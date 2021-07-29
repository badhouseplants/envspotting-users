package service

import (
	"context"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	repo "github.com/badhouseplants/envspotting-users/repo/accounts"
	accountsService "github.com/badhouseplants/envspotting-users/service/accounts"
	authService "github.com/badhouseplants/envspotting-users/service/authorization"
	"github.com/badhouseplants/envspotting-users/third_party/postgres"
	"github.com/badhouseplants/envspotting-users/tools/hasher"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// SingUp creates a user and generates tokens
func SignUp(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountId, error) {
	log := logger.GetGrpcLogger(ctx)
	repo := initRepo(ctx)

	user, err := accountsService.Create(ctx, in)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	userID, code, err := repo.GetIDByUsername(ctx, user.Username)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	_, err = authService.GenerateToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	accountId := &accounts.AccountId{
		Id: userID,
	}

	return accountId, nil
}

// SignIn with login and password
func SignIn(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountId, error) {
	repo := initRepo(ctx)
	// Get user from the database
	password, code, err := repo.GetPasswordByUsername(ctx, in.Username)
	if err != nil {
		return nil, status.Errorf(code, err.Error())
	}
	// Check password
	if err = hasher.ComparePasswords(password, in.Password); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	userID, code, err := repo.GetIDByUsername(ctx, in.Username)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	_, err = authService.GenerateToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	accountId := &accounts.AccountId{
		Id: userID,
	}

	return accountId, nil
}
