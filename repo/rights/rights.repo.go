package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/codes"
)

type RightsStore interface {
	// Read
	GetRight(context.Context, *rights.AccessRuleId) (*rights.AccessRuleInfo, codes.Code, error)
	ListRights(context.Context, rights.Rights_ListServer, *rights.RightsListOptions) (codes.Code, error)
	GetAccessRight(context.Context, *rights.AccessRuleInfo) (*rights.AccessRuleInfo, codes.Code, error)
	GetAppIDByRightID(context.Context, string) (*applications.AppId, codes.Code, error)
	// Write
	CreateRight(context.Context, *rights.AccessRuleInfo) (codes.Code, error)
	UpdateRight(context.Context, *rights.AccessRuleIdAndRight) (codes.Code, error)
	DeleteRight(context.Context, *rights.AccessRuleId) (codes.Code, error)
}
type RightsRepo struct {
	Pool      *pgxpool.Conn
	CreatedAt time.Time
}

var (
	errRightAlreadyExistsWithID = func(id string) error {
		return fmt.Errorf("right with this id already exists: %s", id)
	}
	errRightNotFoundByID = func(id string) error {
		return fmt.Errorf("user with this id can't be found: %s", id)
	}
)

// CreateRight adds a new right to database
func (store RightsRepo) CreateRight(ctx context.Context, in *rights.AccessRuleInfo) (codes.Code, error) {
	const sql = "INSERT INTO rights (id, user_id, application_id, access_right) VALUES ($1, $2, $3, $4::text::user_rights)"

	var log = logger.GetGrpcLogger(ctx)

	_, err := store.Pool.Exec(ctx, sql, in.GetId(), in.GetUserId(), in.GetApplicationId(), parseAccessRightsEnum(in.GetAccessRight().String()))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return codes.AlreadyExists, errRightAlreadyExistsWithID(in.Id)
			default:
				log.Error(err)
				return codes.Internal, err
			}
		}
	}
	return codes.OK, nil
}

// UpdateRight in database
func (store RightsRepo) UpdateRight(ctx context.Context, in *rights.AccessRuleIdAndRight) (codes.Code, error) {
	const sql = "UPDATE rights SET access_right=$2 WHERE id=$1 RETURNING *"

	var log = logger.GetGrpcLogger(ctx)

	tag, err := store.Pool.Exec(ctx, sql, in.GetId(), parseAccessRightsEnum(in.GetAccessRight().String()))
	if tag.RowsAffected() == 0 {
		return codes.NotFound, errRightNotFoundByID(in.Id)
	}
	if err != nil {
		log.Error(err)
		return codes.Internal, err
	}
	return codes.OK, nil
}

// DeleteRight from databse
func (store RightsRepo) DeleteRight(ctx context.Context, in *rights.AccessRuleId) (codes.Code, error) {
	const sql = "DELETE FROM rights WHERE id = $1"

	var log = logger.GetGrpcLogger(ctx)

	tag, err := store.Pool.Exec(ctx, sql, in.GetId())
	if tag.RowsAffected() == 0 {
		return codes.NotFound, errRightNotFoundByID(in.Id)
	}
	if err != nil {
		log.Error(err)
		return codes.Internal, err
	}
	return codes.OK, nil
}

// GetRight from database
func (store RightsRepo) GetRight(ctx context.Context, in *rights.AccessRuleId) (*rights.AccessRuleInfo, codes.Code, error) {

	const sql = "SELECT id, user_id, application_id, access_right FROM rights WHERE id=$1"

	var (
		log         = logger.GetGrpcLogger(ctx)
		err         error
		rightOut    = &rights.AccessRuleInfo{}
		accessRight string
	)

	err = store.Pool.QueryRow(ctx, sql, in.GetId()).Scan(
		&rightOut.Id,
		&rightOut.UserId,
		&rightOut.ApplicationId,
		&accessRight,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, codes.NotFound, errRightNotFoundByID(in.Id)
		} else {
			log.Error(err)
			return nil, codes.Internal, err
		}
	}

	rightOut.AccessRight = rights.AccessRights(rights.AccessRights_value[parseAccessRightsEnum(accessRight)])
	return rightOut, codes.OK, nil
}

// ListRights from database (stream)
func (store RightsRepo) ListRights(ctx context.Context, stream rights.Rights_ListServer, options *rights.RightsListOptions) (codes.Code, error) {
	const sql = "SELECT id, user_id, application_id, access_right FROM rights WHERE application_id=$1"

	var (
		log         = logger.GetGrpcLogger(ctx)
		right       = &rights.AccessRuleInfo{}
		accessRight string
	)

	// TODO: Catch panic
	rows, err := store.Pool.Query(ctx, sql, options.AppId.Id)
	if err != nil {
		log.Error(err)
		return codes.Internal, err
	}
	// Stream applications
	for rows.Next() {
		// Scan apps into struct
		err = rows.Scan(&right.Id, &right.UserId, &right.ApplicationId, &accessRight)
		if err != nil {
			log.Error(err)
			return codes.Internal, err
		}
		right.AccessRight = rights.AccessRights(rights.AccessRights_value[parseAccessRightsEnum(accessRight)])
		// Send stream
		if err := stream.Send(right); err != nil {
			log.Error(err)
			return codes.Internal, err
		}
	}
	return codes.OK, nil
}

// GetAccessRight returns only access_right enum
func (store RightsRepo) GetAccessRight(ctx context.Context, in *rights.AccessRuleInfo) (*rights.AccessRuleInfo, codes.Code, error) {
	const sql = "SELECT access_right FROM rights WHERE user_id=$1 AND application_id=$2"

	var (
		log         = logger.GetGrpcLogger(ctx)
		accessRight string
		err         error
	)

	err = store.Pool.QueryRow(ctx, sql, in.GetUserId(), in.GetApplicationId()).Scan(
		&accessRight,
	)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return nil, codes.NotFound, errRightNotFoundByID(in.Id)
		}
		log.Error(err)
		return nil, codes.Internal, err
	}

	in.AccessRight = rights.AccessRights(rights.AccessRights_value[parseAccessRightsEnum(accessRight)])
	return in, codes.OK, nil
}

// GetAppIDByRightID from database
func (store RightsRepo) GetAppIDByRightID(ctx context.Context, rightID string) (*applications.AppId, codes.Code, error) {
	const sql = "SELECT application_id FROM rights WHERE id = $1"

	var appID = &applications.AppId{}

	var log = logger.GetGrpcLogger(ctx)
	err := store.Pool.QueryRow(ctx, sql, rightID).Scan(&appID.Id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, codes.NotFound, errRightNotFoundByID(rightID)
		} else {
			log.Error(err)
			return nil, codes.Internal, err
		}
	}
	return appID, codes.OK, nil
}

/*
 * Utils
 */

// Parse protobuf enum to the postgres enum format
var parseAccessRightsEnum = func(in string) (out string) {
	var prefix = "ACCESS_RIGHTS_"
	if strings.HasPrefix(in, prefix) {
		out = strings.ReplaceAll(in, prefix, "")
		return out
	}
	out = fmt.Sprintf("%s%s", prefix, in)
	return out
}
