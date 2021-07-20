package accrepo

import (
	"errors"
	"fmt"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"google.golang.org/grpc/codes"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"

	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/context"
)

// ApplicationRepo represents methods to store applications
type AccountStore interface {
	// Read
	GetUser(context.Context, *accounts.AccountId) (*accounts.AccountInfo, codes.Code, error)
	SelfGetUser(context.Context, *accounts.AccountId) (*accounts.FullAccountInfo, codes.Code, error)
	ListUsers(context.Context, accounts.Accounts_ListServer, *accounts.AccountsListOptions) (codes.Code, error)
	GetIDByUsername(context.Context, string) (string, codes.Code, error)
	GetPasswordByUsername(context.Context, string) (string, codes.Code, error)
	GetGitlabTokenByID(context.Context, *accounts.AccountId) (string, codes.Code, error)
	// Write
	Create(context.Context, *accounts.AccountInfoWithSensitive) (codes.Code, error)
	UpdateUser(context.Context, *accounts.FullAccountInfo) (codes.Code, error)
	UpdatePassword(context.Context, *accounts.AccountInfoWithSensitive) (codes.Code, error)
	AddAppToUser(context.Context, string, *applications.AppId) (codes.Code, error)
}

type AccountRepo struct {
	Pool      *pgxpool.Pool
	CreatedAt time.Time
}

//Local errors
var (
	// ErrUserNotFoundByID is and example of a good naming
	ErrUserNotFoundByID = func(id string) error {
		return fmt.Errorf("user with this id can't be found: %s", id)
	}
	// ErrUserNotFoundByNameis and example of a good naming
	ErrUserNotFoundByName = func(username string) error {
		return fmt.Errorf("user with this name can't be found: %s", username)
	}
)

// GetUser from database
func (repo AccountRepo) GetUser(ctx context.Context, id *accounts.AccountId) (*accounts.AccountInfo, codes.Code, error) {
	const sql = "SELECT id, username FROM users WHERE id = $1"

	var (
		acc = &accounts.AccountInfo{}
		log = logger.GetGrpcLogger(ctx)
	)

	err := repo.Pool.QueryRow(ctx, sql, id.GetId()).Scan(&acc.Id, &acc.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, codes.NotFound, ErrUserNotFoundByID(id.GetId())
		} else {
			log.Error(err)
			return nil, codes.Internal, err
		}
	}
	return acc, codes.OK, nil
}

// SelfGetUser from database
func (repo AccountRepo) SelfGetUser(ctx context.Context, id *accounts.AccountId) (*accounts.FullAccountInfo, codes.Code, error) {
	const sql = "SELECT id, username, gitlab_token FROM users WHERE id = $1"

	var (
		acc = &accounts.FullAccountInfo{}
		log = logger.GetGrpcLogger(ctx)
	)

	err := repo.Pool.QueryRow(ctx, sql, id.GetId()).Scan(&acc.Id, &acc.Username, &acc.GitlabToken)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, codes.NotFound, ErrUserNotFoundByID(id.GetId())
		} else {
			log.Error(err)
			return nil, codes.Internal, err
		}
	}
	return acc, codes.OK, nil
}

// ListUsers from database (stream)
func (repo AccountRepo) ListUsers(ctx context.Context, stream accounts.Accounts_ListServer, opt *accounts.AccountsListOptions) (codes.Code, error) {
	const sql = "SELECT id, username FROM users WHERE username LIKE $1 LIMIT $2 OFFSET $3"
	var (
		log = logger.GetGrpcLogger(ctx)
		acc = &accounts.AccountInfo{}
	)

	rows, err := repo.Pool.Query(ctx, sql, opt.GetAccountName().GetUsername()+"%", opt.GetPaging().GetCount(), opt.GetPaging().GetPage())
	if err != nil {
		log.Error(err)
		return codes.Internal, err
	}

	for rows.Next() {
		err = rows.Scan(&acc.Id, &acc.Username)
		if err != nil {
			log.Error(err)
			return codes.Internal, err
		}
		if err := stream.Send(acc); err != nil {
			log.Error(err)
			return codes.Internal, err
		}
	}
	return codes.OK, nil
}

// GetPasswordByUsername returns a hashed password from database
func (repo AccountRepo) GetPasswordByUsername(ctx context.Context, username string) (string, codes.Code, error) {
	const sql = "SELECT password FROM users WHERE username = $1"

	var (
		log      = logger.GetGrpcLogger(ctx)
		password string
	)

	err := repo.Pool.QueryRow(ctx, sql, username).Scan(&password)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", codes.NotFound, ErrUserNotFoundByName(username)
		} else {
			log.Error(err)
			return "", codes.Internal, err
		}
	}
	return password, codes.OK, nil
}

// GetIDByUsername  returns ID from database
func (repo AccountRepo) GetIDByUsername(ctx context.Context, username string) (string, codes.Code, error) {
	const sql = "SELECT id FROM users WHERE username = $1"

	var (
		log = logger.GetGrpcLogger(ctx)
		id  string
	)

	err := repo.Pool.QueryRow(ctx, sql, username).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", codes.NotFound, ErrUserNotFoundByName(username)
		} else {
			log.Error(err)
			return "", codes.Internal, err
		}
	}
	return id, codes.OK, nil
}

// GetGitlabTokenByID return ecnrypted gitlab token from database
func (repo AccountRepo) GetGitlabTokenByID(ctx context.Context, id *accounts.AccountId) (string, codes.Code, error) {
	const sql = "SELECT gitlab_token FROM users WHERE id = $1"

	var (
		log   = logger.GetGrpcLogger(ctx)
		token string
	)

	err := repo.Pool.QueryRow(ctx, sql, id.GetId()).Scan(&token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", codes.NotFound, ErrUserNotFoundByID(id.GetId())
		} else {
			log.Error(err)
			return "", codes.Internal, err
		}
	}
	return token, codes.OK, nil
}

// Create a new user  in database
func (repo AccountRepo) Create(ctx context.Context, acc *accounts.AccountInfoWithSensitive) (code codes.Code, err error) {
	// SQL
	const sql = "INSERT INTO users (id, username, password) VALUES ($1, $2, $3)"

	var log = logger.GetGrpcLogger(ctx)
	// Not checking zero rows affected condition
	_, err = repo.Pool.Exec(ctx, sql, acc.GetId(), acc.GetUsername(), acc.GetPassword())
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

// UpdateUser is upadting account data in database
func (repo AccountRepo) UpdateUser(ctx context.Context, acc *accounts.FullAccountInfo) (codes.Code, error) {
	const sql = "UPDATE users SET username = $2, gitlab_token = $3, WHERE id = $1"

	var (
		log = logger.GetGrpcLogger(ctx)
	)

	tag, err := repo.Pool.Exec(ctx, sql, acc.GetId(), acc.GetUsername(), acc.GetGitlabToken())
	if tag.RowsAffected() == 0 {
		return codes.NotFound, ErrUserNotFoundByID(acc.GetId())
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

// UpdatePassword id updating account password in database
func (repo AccountRepo) UpdatePassword(ctx context.Context, acc *accounts.AccountInfoWithSensitive) (codes.Code, error) {
	const sql = "UPDATE users SET password=$3 WHERE id=$1 AND username=$2 RETURNING *"

	var log = logger.GetGrpcLogger(ctx)

	tag, err := repo.Pool.Exec(ctx, sql, acc.GetId(), acc.GetUsername(), acc.GetPassword())
	if tag.RowsAffected() == 0 {
		return codes.NotFound, ErrUserNotFoundByID(acc.GetId())
	}
	if err != nil {
		log.Error(err)
		return codes.Internal, err
	}
	return codes.OK, nil
}

// AddAppToUser appends a new application id in users.applications column
func (repo AccountRepo) AddAppToUser(ctx context.Context, userID string, appID *applications.AppId) (codes.Code, error) {
	const sql = "UPDATE users SET applications = array_append(applications, $2) WHERE id=$1"

	var (
		log = logger.GetGrpcLogger(ctx)
	)

	tag, err := repo.Pool.Exec(ctx, sql, userID, appID.Id)
	if tag.RowsAffected() == 0 {
		return codes.NotFound, ErrUserNotFoundByID(userID)
	}
	if err != nil {
		log.Error(err)
		return codes.Internal, err
	}
	return codes.OK, nil
}
