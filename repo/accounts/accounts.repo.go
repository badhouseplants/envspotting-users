package accrepo

import (
	"errors"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"google.golang.org/grpc/codes"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/context"
)

// ApplicationRepo represents methods to store applications
type AccountStore interface {
	Create(context.Context, *accounts.AccountInfoWithSensitive) (error, codes.Code)
}

type AccountRepo struct {
	Pool      *pgxpool.Pool
	CreatedAt time.Time
}

func (repo AccountRepo) Create(ctx context.Context, user *accounts.AccountInfoWithSensitive) (code codes.Code, err error) {
	// SQL
	const sql = "INSERT INTO users (id, username, password) VALUES ($1, $2, $3)"

	var log = logger.GetGrpcLogger(ctx)
	// Not checking zero rows affected condition
	_, err = repo.Pool.Exec(ctx, sql, user.GetId(), user.GetUsername(), user.GetPassword())
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return codes.AlreadyExists, err
			default:
				log.Error(err)
				return codes.Internal, err
			}
		}
	}
	return codes.OK, nil
}

func (repo AccountRepo) UpdateUser(ctx context.Context, user *accounts.AccountInfo) (core codes.Code, err error) {
	const sql = `
	UPDATE users SET
  	username = $2,
  	gitlab_token = $3,
	WHERE id = $1
	`
	var (
		log = logger.GetGrpcLogger(ctx)
		tag pgconn.CommandTag
	)

	tag, err = repo.Pool.Exec(ctx, sql, user.GetId(), user.GetUsername(), user.GetGitlabToken())

	if tag.RowsAffected() == 0 {
		return codes.NotFound, errors.New("user with this id cannot be found")
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return codes.AlreadyExists, err
			default:
				log.Error(err)
				return codes.Internal, err
			}
		}
	}
	return codes.OK, err
}
