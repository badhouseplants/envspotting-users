package service

import (
	"context"

	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	"github.com/badhouseplants/envspotting-go-proto/models/users/authorization"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"google.golang.org/grpc"
)

type authorizationGrpcImpl struct {
	authorization.UnimplementedAuthorizationServer
}

func Register(grpcServer *grpc.Server) {
	authorization.RegisterAuthorizationServer(grpcServer, &authorizationGrpcImpl{})
}

func (s *authorizationGrpcImpl) RefreshToken(ctx context.Context, in *accounts.AccountId) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	return RefreshToken(ctx, in)
}

func (s *authorizationGrpcImpl) ParseIdFromToken(ctx context.Context, in *common.EmptyMessage) (*accounts.AccountId, error) {
	logger.EnpointHit(ctx)
	return ParseIdFromToken(ctx, in)
}

func (s *authorizationGrpcImpl) ValidateToken(ctx context.Context, in *common.EmptyMessage) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	return ValidateToken(ctx)
}
