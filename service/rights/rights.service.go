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
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
)

var rightsrepo repo.RightsStore

var initRepo = func(ctx context.Context) repo.RightsStore {
	if rightsrepo == nil {
		rightsrepo = repo.RightsRepo{
			Pool:      postgres.Pool(ctx),
			CreatedAt: time.Now(),
		}
	}
	return rightsrepo
}

func Create(ctx context.Context, in *rights.AccessRuleWithoutId) (*rights.AccessRuleInfo, error) {
	repo := initRepo(ctx)
	rightOut := &rights.AccessRuleInfo{
		Id:            uuid.NewString(),
		UserId:        in.UserId,
		ApplicationId: in.ApplicationId,
		AccessRight:   rights.AccessRights(in.AccessRight),
	}
	code, err := repo.CreateRight(ctx, rightOut)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return rightOut, nil
}

func Update(ctx context.Context, in *rights.AccessRuleIdAndRight) (*rights.AccessRuleIdAndRight, error) {
	repo := initRepo(ctx)
	code, err := repo.UpdateRight(ctx, in)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return in, nil
}

func Delete(ctx context.Context, in *rights.AccessRuleId) (*common.EmptyMessage, error) {
	repo := initRepo(ctx)
	code, err := repo.DeleteRight(ctx, in)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return &common.EmptyMessage{}, nil
}

func Get(ctx context.Context, in *rights.AccessRuleId) (*rights.AccessRuleInfo, error) {
	repo := initRepo(ctx)
	rightOut, code, err := repo.GetRight(ctx, in)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return rightOut, nil
}

func List(ctx context.Context, in *rights.RightsListOptions, stream rights.Rights_ListServer) (err error) {
	repo := initRepo(ctx)
	code, err := repo.ListRights(ctx, stream, in)
	if err != nil {
		return status.Error(code, err.Error())
	}
	return nil
}

func CheckRight(ctx context.Context, applicationID string, right *rights.AccessRights) (err error) {
	repo := initRepo(ctx)
	userID, err := token.ParseUserID(ctx)
	if err != nil {
		return err
	}
	access := &rights.AccessRuleInfo{
		UserId:        userID,
		ApplicationId: applicationID,
	}
	access, code, err := repo.GetAccessRight(ctx, access)
	if err != nil || access.AccessRight < rights.AccessRights(*right) {
		return status.Error(code, err.Error())
	}
	return nil
}

func GetAppIDByRightID(ctx context.Context, rightID string) (*applications.AppId, error) {
	repo := initRepo(ctx)
	appID, code, err := repo.GetAppIDByRightID(ctx, rightID)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return appID, err
}
