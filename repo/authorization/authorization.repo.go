package repo

import (
	"context"
	"errors"
	"time"

	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
)

type AuthorizationStore interface {
	GetRefreshToken(context.Context, *RefreshToken) (*RefreshToken, codes.Code, error)
	SetRefreshToken(context.Context, *RefreshToken) (codes.Code, error)
	DelRefreshToken(context.Context, *RefreshToken) (codes.Code, error)
}

type AuthorizationRepo struct {
	Redis     *redis.Client
	CreatedAt time.Time
}

type RefreshToken struct {
	ID                 string
	BrowserFingerprint string `redis:"browser_fingerprint"`
	UserID             string `redis:"user_id"`
}

var (
	rtExpirationTime = func() time.Duration {
		return viper.GetDuration("refresh_token_expiry") * time.Hour
	}
)

var (
	errRTNotFound = errors.New("refresh token not found")
)

func (repo AuthorizationRepo) GetRefreshToken(ctx context.Context, rt *RefreshToken) (*RefreshToken, codes.Code, error) {
	log := logger.GetServerLogger()
	oldRT := repo.Redis.HGetAll(ctx, rt.ID)
	if len(oldRT.Val()) == 0 {
		return nil, codes.PermissionDenied, errRTNotFound
	}
	if err := oldRT.Scan(rt); err != nil {
		log.Error(err)
		return nil, codes.Internal, err
	}
	return rt, codes.OK, nil
}

func (repo AuthorizationRepo) SetRefreshToken(ctx context.Context, rt *RefreshToken) (codes.Code, error) {
	log := logger.GetGrpcLogger(ctx)
	redCmd := repo.Redis.HSet(ctx, rt.ID,
		"user_id", rt.UserID,
		"browser_fingerprint", rt.BrowserFingerprint,
	)
	if redCmd.Err() != nil {
		log.Error(redCmd.Err())
		return codes.Internal, redCmd.Err()
	}
	repo.Redis.Expire(ctx, rt.ID, rtExpirationTime())
	return codes.OK, nil
}

func (repo AuthorizationRepo) DelRefreshToken(ctx context.Context, rt *RefreshToken) (codes.Code, error) {
	log := logger.GetServerLogger()
	redCmd := repo.Redis.Del(ctx, rt.ID)
	if redCmd.Err() != nil {
		log.Error(redCmd.Err())
		return codes.Internal, redCmd.Err()
	}
	return codes.OK, nil
}
