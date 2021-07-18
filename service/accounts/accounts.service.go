package accserv

import (
	"context"
	"fmt"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	accrepo "github.com/badhouseplants/envspotting-users/repo/accounts"
	"github.com/badhouseplants/envspotting-users/third_party/postgres"
	"github.com/badhouseplants/envspotting-users/tools/hasher"
	"github.com/badhouseplants/envspotting-users/tools/token"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var initRepo = func() *accrepo.AccountRepo {
	return &accrepo.AccountRepo{
		Pool:      postgres.Pool(),
		CreatedAt: time.Now(),
	}
}

// Create a new user
func Create(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountInfo, error) {
	repo := initRepo()
	// Fill user struct and clear a struct with a password
	id := uuid.New().String()
	user := &accounts.AccountInfoWithSensitive{
		Id:       id,
		Username: in.GetUsername(),
		Password: hasher.Encrypt(in.GetPassword()),
	}
	in.Reset()
	code, err := repo.Create(ctx, user)
	if code != codes.OK {
		return nil, status.Error(code, err.Error())
	}

	out := &accounts.AccountInfo{
		Id:       user.Id,
		Username: user.Username,
	}
	return out, nil
}

func UpdateUser(ctx context.Context, in *accounts.AccountInfo) (*accounts.AccountInfo, error) {
	repo := initRepo()
	_, err := repo.UpdateUser(ctx, in)
	if err != nil {
		return nil, err
	}
	return in, nil
}

func UpdatePassword(ctx context.Context, in *accounts.PasswordUpdate) (*common.EmptyMessage, error) {
	creds := &accounts.AccountCreds{
		Username: in.GetUsername(),
		Password: in.GetOldPassword(),
	}
	if err := CheckCreds(ctx, store, creds); err != nil {
		return nil, err
	}
	newCreds := &accounts.AccountInfoWithSensitive{
		Id:       in.GetId(),
		Username: in.GetUsername(),
		Password: hasher.Encrypt(in.GetNewPassword()),
	}
	err := store.UpdatePassword(ctx, newCreds)
	if err != nil {
		return nil, err
	}
	return &common.EmptyMessage{}, nil
}

func List(ctx context.Context, stream accounts.Accounts_ListServer, username *accounts.AccountName) error {
	err := store.List(ctx, stream, username)
	if err != nil {
		return err
	}
	return nil
}

func AddAppToUser(ctx context.Context, in *applications.AppId) (*common.EmptyMessage, error) {
	store := &repo.AccountStore{
		Pool:      postgres.Pool(),
		CreatedAt: time.Now(),
	}
	userID, err := token.ParseUserID(ctx)
	if err != nil {
		return nil, err
	}
	err = store.AddAppToUser(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	return &common.EmptyMessage{}, nil
}

// CheckCreds with login and password
func CheckCreds(ctx context.Context, in *accounts.AccountCreds) error {
	var (
		password string
		err      error
	)
	// Get user from the database
	password, err = store.GetPasswordByUsername(ctx, in.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return status.Errorf(codes.NotFound, fmt.Sprintf("User not found: %s", in.Username))
		}
		return status.Error(codes.Internal, err.Error())
	}
	// Check password
	if err = hasher.ComparePasswords(password, in.Password); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}
	return nil
}

func GetGitlabTokenByID(ctx context.Context) (string, error) {
	store := &repo.AccountStore{
		Pool:      postgres.Pool(),
		CreatedAt: time.Now(),
	}
	userID, err := token.ParseUserID(ctx)
	if err != nil {
		return "", err
	}
	token, err := store.GetGitlabTokenByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return token, nil

}
