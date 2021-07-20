package service

import (
	"context"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/common"
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

var initRepo = func() repo.AccountStore {
	var accrepo repo.AccountStore
	accrepo = repo.AccountRepo{
		Pool:      postgres.Pool(),
		CreatedAt: time.Now(),
	}
	return accrepo
}

func SignUp(ctx context.Context, in *accounts.AccountCreds) (*common.EmptyMessage, error) {
	repo := initRepo()
	log := logger.GetGrpcLogger(ctx)
	_, err := accountsService.Create(ctx, in)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	userId, code, err := repo.GetIDByUsername(ctx, in.Username)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	out, err := authService.GenerateToken(ctx, userId)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SignIn with login and password
func SignIn(ctx context.Context, in *accounts.AccountCreds) (*common.EmptyMessage, error) {
	repo := initRepo()
	// Get user from the database
	password, code, err := repo.GetPasswordByUsername(ctx, in.Username)
	if err != nil {
		return nil, status.Errorf(code, err.Error())
	}
	// Check password
	if err = hasher.ComparePasswords(password, in.Password); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	userId, code, err := repo.GetIDByUsername(ctx, in.Username)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	out, err := authService.GenerateToken(ctx, userId)
	if err != nil {
		return nil, err
	}
	return out, nil
}
