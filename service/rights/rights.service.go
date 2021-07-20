package service

import (
	"context"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/apps/applications"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/rights"
	repo "github.com/badhouseplants/envspotting-users/repo/rights"
	"github.com/badhouseplants/envspotting-users/third_party/postgres"
	"github.com/badhouseplants/envspotting-users/tools/token"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
)

var initRepo = func() repo.RightStore {
	rightsrepo := repo.RightStore{
		Pool:      postgres.Pool(),
		CreatedAt: time.Now(),
	}
	return rightsrepo
}


func Create(ctx context.Context, in *rights.AccessRuleWithoutId) (*rights.AccessRuleInfo, error) {
	repo := initRepo()
	rightOut := &rights.AccessRuleInfo{
		Id:            uuid.NewString(),
		UserId:        in.UserId,
		ApplicationId: in.ApplicationId,
		AccessRight:   rights.AccessRights(in.AccessRight),
	}
	if err := repo.Create(ctx, rightOut); err != nil {
		return nil, err
	}
	return rightOut, nil
}

func Update(ctx context.Context, in *rights.AccessRuleIdAndRight) (*rights.AccessRuleIdAndRight, error) {
	repo := initRepo()
	err := repo.Update(ctx, in)
	if err != nil {
		return nil, err
	}
	return in, nil
}

func Delete(ctx context.Context, in *rights.AccessRuleId) (*common.EmptyMessage, error) {
	repo := initRepo()
	err := repo.Delete(ctx, in)
	if err != nil {
		return nil, err
	}
	return &common.EmptyMessage{}, nil
}

func Get(ctx context.Context, in *rights.AccessRuleId) (*rights.AccessRuleInfo, error) {
	repo := initRepo()
	rightOut, err := repo.Get(ctx, in)
	if err != nil {
		return nil, err
	}
	return rightOut, nil
}

func List(ctx context.Context, in *rights.RightsListOptions, stream rights.Rights_ListServer) (err error) {
	repo := initRepo()
	err = repo.List(ctx, stream, in)
	if err != nil {
		return err
	}
	return nil
}

func CheckRight(ctx context.Context, applicationID string, right *rights.AccessRights) (err error) {
	repo := initRepo()
	userID, err := token.ParseUserID(ctx)
	if err != nil {
		return err
	}
	access := &rights.AccessRuleInfo{
		UserId:        userID,
		ApplicationId: applicationID,
	}
	access, err = repo.GetAccessRight(ctx, access)
	if err != nil || access.AccessRight < rights.AccessRights(*right)  {
		return status.Error(codes.PermissionDenied, "action is not allowed by access rules")
	}
	return nil
}

func GetAppIDByRightID(ctx context.Context, rightID string) (*applications.AppId, error) {
	repo := initRepo()
	appId, err := repo.GetAppIDByRightID(ctx, rightID)
	if err != nil {
		return nil, err
	}
	return &applications.AppId{
		Id: appId,
	}, err
}

