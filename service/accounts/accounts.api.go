package service

import (
	"context"
	"errors"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/badhouseplants/envspotting-users/tools/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountsGrpcServer struct {
	accounts.UnimplementedAccountsServer
}

type tokenGrpcServer struct {
	accounts.UnimplementedTokensServer
}

func Register(grpcServer *grpc.Server) {
	accounts.RegisterAccountsServer(grpcServer, &accountsGrpcServer{})
	accounts.RegisterTokensServer(grpcServer, &tokenGrpcServer{})
}

const ErrSelfOperation = "you not allowed to perform this opertaions for another accounts"

func (s *accountsGrpcServer) Create(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountInfo, error) {
	logger.EnpointHit(ctx)
	return Create(ctx, in)
}

func (s *accountsGrpcServer) UpdateUser(ctx context.Context, in *accounts.FullAccountInfo) (*accounts.FullAccountInfo, error) {
	logger.EnpointHit(ctx)
	code, err := checkSelfOperation(ctx, in.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return UpdateUser(ctx, in)
}

func (s *accountsGrpcServer) UpdatePassword(ctx context.Context, in *accounts.PasswordUpdate) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	code, err := checkSelfOperation(ctx, in.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return UpdatePassword(ctx, in)
}

func (s *accountsGrpcServer) Get(ctx context.Context, in *accounts.AccountId) (*accounts.AccountInfo, error) {
	logger.EnpointHit(ctx)
	return Get(ctx, in)
}

func (s *accountsGrpcServer) List(in *accounts.AccountsListOptions, stream accounts.Accounts_ListServer) error {
	logger.EnpointHit(stream.Context())
	err := List(stream.Context(), stream, in)
	if err != nil {
		return err
	}
	return nil
}

func (s *accountsGrpcServer) AddAppToUser(ctx context.Context, in *applications.AppId) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	// err := CheckRight(ctx, in.Id, rights.AccessRights_ACCESS_RIGHTS_READ_UNSPECIFIED.Enum())
	err := errors.New("Asdf")
	if err != nil {
		return nil, err
	}
	code, err := checkSelfOperation(ctx, in.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return AddAppToUser(ctx, in)
}

func (s *accountsGrpcServer) SelfGet(ctx context.Context, in *accounts.AccountId) (*accounts.FullAccountInfo, error) {
	logger.EnpointHit(ctx)
	code, err := checkSelfOperation(ctx, in.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return SelfGet(ctx, in)
}

func (s *tokenGrpcServer) GetGitlabTokenByAccountID(ctx context.Context, in *accounts.AccountId) (*accounts.GitlabToken, error) {
	logger.EnpointHit(ctx)
	code, err := checkSelfOperation(ctx, in.Id)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return nil, nil
}

var checkSelfOperation = func(ctx context.Context, id string) (codes.Code, error) {
	log := logger.GetGrpcLogger(ctx)
	idFromToken, err := token.ParseUserID(ctx)
	if err != nil {
		log.Error(err)
		return codes.Internal, err
	} else if id != idFromToken {
		return codes.PermissionDenied, errors.New(ErrSelfOperation)
	}
	return codes.OK, nil
}
