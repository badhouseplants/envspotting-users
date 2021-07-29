package service

import (
	"context"

	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	"github.com/badhouseplants/envspotting-go-proto/models/users/authentication"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"google.golang.org/grpc"
)

type serviceGrpcImpl struct {
	authentication.UnimplementedAuthenticationServer
}

func Register(grpcServer *grpc.Server) {
	authentication.RegisterAuthenticationServer(grpcServer, &serviceGrpcImpl{})
}

func (s *serviceGrpcImpl) SignUp(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountId, error) {
	logger.EnpointHit(ctx)
	return SignUp(ctx, in)
}

func (s *serviceGrpcImpl) SignIn(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountId, error) {
	logger.EnpointHit(ctx)
	return SignIn(ctx, in)
}


