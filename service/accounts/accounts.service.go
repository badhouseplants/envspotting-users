package service

import (
	"context"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	repo "github.com/badhouseplants/envspotting-users/repo/accounts"
	"github.com/badhouseplants/envspotting-users/third_party/postgres"
	"github.com/badhouseplants/envspotting-users/tools/encryption"
	"github.com/badhouseplants/envspotting-users/tools/hasher"
	"github.com/badhouseplants/envspotting-users/tools/token"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authserv "github.com/badhouseplants/envspotting-users/service/authorization"
)

var accrepo repo.AccountStore

var initRepo = func(ctx context.Context) repo.AccountStore {
	if accrepo == nil {
		accrepo = repo.AccountRepo{
			Pool:      postgres.Pool(ctx),
			CreatedAt: time.Now(),
		}
	}
	return accrepo
}

// Create a new user
func Create(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountInfo, error) {
	repo := initRepo(ctx)

	id := uuid.New().String()
	user := &accounts.AccountInfoWithSensitive{
		Id:       id,
		Username: in.GetUsername(),
		Password: hasher.Encrypt(in.GetPassword()),
	}
	in.Reset()

	code, err := repo.CreateUser(ctx, user)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	out := &accounts.AccountInfo{
		Id:       user.Id,
		Username: user.Username,
	}
	return out, nil
}

func SelfGet(ctx context.Context, in *accounts.AccountId) (*accounts.FullAccountInfo, error) {
	repo := initRepo(ctx)
	user, code, err := repo.SelfGetUser(ctx, in)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	user.GitlabToken, code, err = encryption.Decrypt(ctx, user.GitlabToken)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return user, nil

}

func Get(ctx context.Context, in *accounts.AccountId) (*accounts.AccountInfo, error) {
	repo := initRepo(ctx)
	user, code, err := repo.GetUser(ctx, in)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return user, nil
}

func UpdateUser(ctx context.Context, in *accounts.FullAccountInfo) (*accounts.FullAccountInfo, error) {
	var (
		code codes.Code
		err  error
	)
	repo := initRepo(ctx)
	in.GitlabToken, code, err = encryption.Encrypt(ctx, in.GitlabToken)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	_, err = repo.UpdateUser(ctx, in)
	if err != nil {
		return nil, err
	}
	return in, nil
}

func UpdatePassword(ctx context.Context, in *accounts.PasswordUpdate) (*common.EmptyMessage, error) {
	repo := initRepo(ctx)
	creds := &accounts.AccountCreds{
		Username: in.GetUsername(),
		Password: in.GetOldPassword(),
	}
	if err := CheckCreds(ctx, creds); err != nil {
		return nil, err
	}
	newCreds := &accounts.AccountInfoWithSensitive{
		Id:       in.GetId(),
		Username: in.GetUsername(),
		Password: hasher.Encrypt(in.GetNewPassword()),
	}
	code, err := repo.UpdatePassword(ctx, newCreds)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return &common.EmptyMessage{}, nil
}

func List(ctx context.Context, stream accounts.Accounts_ListServer, options *accounts.AccountsListOptions) error {
	repo := initRepo(ctx)
	code, err := repo.ListUsers(ctx, stream, options)
	if err != nil {
		return status.Error(code, err.Error())
	}
	return nil
}

func AddAppToUser(ctx context.Context, in *applications.AppId) (*common.EmptyMessage, error) {
	repo := initRepo(ctx)
	tknStr, code, err := authserv.GetAuthorizationToken(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	userID, code, err := token.ParseUserID(ctx, tknStr)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	code, err = repo.AddAppToUser(ctx, userID, in)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return &common.EmptyMessage{}, nil
}

func GetAppsFromUser(ctx context.Context, in *accounts.AccountId) (*accounts.AccountsApps, error) {
	repo := initRepo(ctx)
	apps, code, err := repo.GetAppsFromUser(ctx, in)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return apps, nil
}

// CheckCreds with login and password
func CheckCreds(ctx context.Context, in *accounts.AccountCreds) error {
	var (
		password string
		err      error
		code     codes.Code
	)
	repo := initRepo(ctx)
	// Get user from the database
	password, code, err = repo.GetPasswordByUsername(ctx, in.Username)
	if err != nil {
		return status.Error(code, err.Error())
	}
	// Check password
	if err = hasher.ComparePasswords(password, in.Password); err != nil {
		return status.Error(code, err.Error())
	}
	return nil
}

func GetGitlabTokenByID(ctx context.Context, id *accounts.AccountId) (*accounts.AccountGitlabToken, error) {
	repo := initRepo(ctx)
	tokenEnc, code, err := repo.GetGitlabTokenByID(ctx, id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	token, code, err := encryption.Decrypt(ctx, tokenEnc)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	return &accounts.AccountGitlabToken{GitlabToken: token}, nil
}