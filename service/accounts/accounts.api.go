package accserv

import (
	"context"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/badhouseplants/envspotting-users/tools/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type accountsGrpcServer struct {
	accounts.UnimplementedAccountsServer
}

func Register(grpcServer *grpc.Server) {
	accounts.RegisterAccountsServer(grpcServer, &accountsGrpcServer{})
}

func (s *accountsGrpcServer) Create(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountInfo, error) {
	logger.EnpointHit(ctx)
	return Create(ctx, in)
}

func (s *accountsGrpcServer) UpdateUser(ctx context.Context, in *accounts.AccountInfo) (*accounts.AccountInfo, error) {
	log := logger.EnpointHit(ctx)
	idFromToken, err := token.ParseUserID(ctx)
	if err != nil {
		log.Error(err)
		return nil, err
	} else if in.Id != idFromToken {
		return nil, status.Error(codes.PermissionDenied, "do not tru to update other users")
	}
	return UpdateUser(ctx, in)
}

func (s *accountsGrpcServer) UpdatePassword(ctx context.Context, in *accounts.PasswordUpdate) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	idFromToken, err := token.ParseUserID(ctx)
	if err != nil {
		return nil, err
	} else if in.Id != idFromToken {
		return nil, status.Error(codes.PermissionDenied, "fuck you, do not tru to update other users")
	}
	return UpdatePassword(ctx, in)
}

func (s *accountsGrpcServer) Get(ctx context.Context, in *accounts.AccountId) (*accounts.AccountInfo, error) {
	logger.EnpointHit(ctx)
	return Get(ctx, in)
}

func (s *accountsGrpcServer) List(in *accounts.AccountName, stream accounts.Accounts_ListServer) error {
	logger.EnpointHit(stream.Context())
	err := List(stream.Context(), stream, in)
	if err != nil {
		return err
	}
	return nil
}

func (s *accountsGrpcServer) AddAppToUser(ctx context.Context, in *applications.AppId) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	// Check if user is  can read the application
	err := rs.CheckRight(ctx, in.Id, rights.AccessRights_READ.Enum())
	if err != nil {
		return nil, err
	}
	idFromToken, err := token.ParseUserID(ctx)
	if err != nil {
		return nil, err
	} else if in.Id != idFromToken {
		return nil, status.Error(codes.PermissionDenied, "fuck you, do not tru to update other users")
	}
	return AddAppToUser(ctx, in)
}
