package service

import (
	context "context"

	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"google.golang.org/grpc"
)

type rightsGrpcServer struct {
	rights.UnimplementedRightsServer
}

func Register(grpcServer *grpc.Server) {
	rights.RegisterRightsServer(grpcServer, &rightsGrpcServer{})
}

// TODO: add internal token
func (s *rightsGrpcServer) Init(ctx context.Context, in *rights.AccessRuleWithoutId) (*rights.AccessRuleInfo, error) {
	logger.EnpointHit(ctx)
	return Create(ctx, in)
}

func (s *rightsGrpcServer) Create(ctx context.Context, in *rights.AccessRuleWithoutId) (*rights.AccessRuleInfo, error) {
	logger.EnpointHit(ctx)
	err := CheckRight(ctx, in.ApplicationId, rights.AccessRights_ACCESS_RIGHTS_DELETE.Enum())
	if err != nil {
		return nil, err
	}
	return Create(ctx, in)
}

func (s *rightsGrpcServer) Update(ctx context.Context, in *rights.AccessRuleIdAndRight) (*rights.AccessRuleIdAndRight, error) {
	logger.EnpointHit(ctx)
	appID, err := GetAppIDByRightID(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	err = CheckRight(ctx, appID.Id, rights.AccessRights_ACCESS_RIGHTS_DELETE.Enum())
	if err != nil {
		return nil, err
	}
	return Update(ctx, in)
}

func (s *rightsGrpcServer) Delete(ctx context.Context, in *rights.AccessRuleId) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	appID, err := GetAppIDByRightID(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	err = CheckRight(ctx, appID.Id, rights.AccessRights_ACCESS_RIGHTS_DELETE.Enum())
	if err != nil {
		return nil, err
	}

	return Delete(ctx, in)
}

func (s *rightsGrpcServer) Get(ctx context.Context, in *rights.AccessRuleId) (*rights.AccessRuleInfo, error) {
	logger.EnpointHit(ctx)
	appID, err := GetAppIDByRightID(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	err = CheckRight(ctx, appID.Id, rights.AccessRights_ACCESS_RIGHTS_DELETE.Enum())
	if err != nil {
		return nil, err
	}
	return Get(ctx, in)
}

func (s *rightsGrpcServer) List(in *rights.RightsListOptions, stream rights.Rights_ListServer) (err error) {
	logger.EnpointHit(stream.Context())

	err = CheckRight(
		stream.Context(),
		in.AppId.Id,
		rights.AccessRights_ACCESS_RIGHTS_DELETE.Enum(),
	)
	if err != nil {
		return err
	}

	err = List(stream.Context(), in, stream)
	if err != nil {
		return err
	}
	return nil
}

func (s *rightsGrpcServer) ListAvailableApps(in *rights.AvailableAppsListOptions, stream rights.Rights_ListAvailableAppsServer) (err error) {
	logger.EnpointHit(stream.Context())
	return ListAvailableApps(in, stream)
}

func (s *rightsGrpcServer) CheckRight(ctx context.Context, in *rights.AccessRightRequest) (*common.EmptyMessage, error) {
	logger.EnpointHit(ctx)
	err := CheckRight(ctx, in.GetApplicationId().GetId(), in.GetAccessRight().Enum())
	if err != nil {
		return nil, err
	}
	return &common.EmptyMessage{}, nil
}
