package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RightStore struct {
	Pool      *pgxpool.Pool
	CreatedAt time.Time
}

func (store RightStore) Create(ctx context.Context, in *rights.AccessRuleInfo) (err error) {
	const sql = "INSERT INTO rights (id, user_id, application_id, access_right) VALUES ($1, $2, $3, $4::text::user_rights)"
	var log = logger.GetGrpcLogger(ctx)
	log.Info(in.AccessRight)
	comtag, err := store.Pool.Exec(ctx, sql, in.GetId(), in.GetUserId(), in.GetApplicationId(), in.GetAccessRight().String())
	log.Info(comtag.String())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return status.Error(codes.AlreadyExists, err.Error())
			// case pgerrcode.Cons
			default:
				log.Error(err)
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
	return nil
}

func (store RightStore) Update(ctx context.Context, in *rights.AccessRuleIdAndRight) (err error) {
	const sql = "UPDATE rights SET access_right=$2 WHERE id=$1 RETURNING *"
	var log = logger.GetGrpcLogger(ctx)
	tag, err := store.Pool.Exec(ctx, sql, in.GetId(), in.GetAccessRight().String())
	if tag.RowsAffected() == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("user right with this id can't be found: %s", in.GetId()))
	}
	if err != nil {
		log.Error(err)
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (store RightStore) Delete(ctx context.Context, in *rights.AccessRuleId) (err error) {
	const sql = "DELETE FROM rights WHERE id = $1"
	var log = logger.GetGrpcLogger(ctx)
	tag, err := store.Pool.Exec(ctx, sql, in.GetId())
	if tag.RowsAffected() == 0 {
		return status.Error(codes.NotFound, fmt.Sprintf("user right with this id can't be found: %s", in.GetId()))
	}
	if err != nil {
		log.Error(err)
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (store RightStore) Get(ctx context.Context, in *rights.AccessRuleId) (*rights.AccessRuleInfo, error) {
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
			return nil, status.Error(codes.NotFound, fmt.Sprintf("right with this id can't be found: %s", in.Id))
		} else {
			log.Error(err)
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	rightOut.AccessRight = rights.AccessRights(rights.AccessRights_value[accessRight])
	return rightOut, nil
}

func (store RightStore) List(ctx context.Context, stream rights.Rights_ListServer, options *rights.RightsListOptions) error {
	const sql = "SELECT id, user_id, application_id, access_right FROM rights WHERE application_id=$1"
	var (
		log         = logger.GetGrpcLogger(ctx)
		right       = &rights.AccessRuleInfo{}
		accessRight string
	)
	// Get contours
	// TODO: Catch panic
	rows, err := store.Pool.Query(ctx, sql, options.AppId.Id)
	if err != nil {
		log.Error(err)
		return status.Error(codes.Internal, err.Error())
	}
	// Stream applications
	for rows.Next() {
		// Scan apps into struct
		err = rows.Scan(&right.Id, &right.UserId, &right.ApplicationId, &accessRight)
		if err != nil {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
		right.AccessRight = rights.AccessRights(rights.AccessRights_value[accessRight])
		// Send stream
		if err := stream.Send(right); err != nil {
			log.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
	}
	return nil
}

func (store RightStore) GetAccessRight(ctx context.Context, access *rights.AccessRuleInfo) (*rights.AccessRuleInfo, error) {
	const sql = "SELECT access_right FROM rights WHERE user_id=$1 AND application_id=$2"
	var (
		log         = logger.GetGrpcLogger(ctx)
		accessRight string
		err         error
	)
	err = store.Pool.QueryRow(ctx, sql, access.GetUserId(), access.GetApplicationId()).Scan(
		&accessRight,
	)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			return nil, status.Error(codes.NotFound, "right cannot be found")
		}

		log.Error(err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	access.AccessRight = rights.AccessRights(rights.AccessRights_value[accessRight])
	return access, nil
}

func (store RightStore) GetAppIDByRightID(ctx context.Context, rightID string) (string, error) {
	const sql = "SELECT application_id FROM rights WHERE id = $1"
	var appID string
	var log = logger.GetGrpcLogger(ctx)
	err := store.Pool.QueryRow(ctx, sql, rightID).Scan(&appID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			default:
				log.Error(err)
				return "", status.Error(codes.Internal, err.Error())
			}
		}
	}
	return appID, nil

}

