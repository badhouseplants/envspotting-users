package repo

import (
	"context"
	"time"

	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
)

type AuthorizationStore interface {
	GetRefreshToken(context.Context, *RefreshToken) (*RefreshToken, codes.Code, error)
	SetRefreshToken(context.Context, *RefreshToken) (codes.Code, error)
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

func (repo AuthorizationRepo) GetRefreshToken(ctx context.Context, refreshToken *RefreshToken) (*RefreshToken, codes.Code, error) {
	log := logger.GetServerLogger()
	rt := &RefreshToken{}
	oldRT := repo.Redis.HGetAll(ctx, refreshToken.ID)
	redCmd := repo.Redis.Del(ctx, refreshToken.ID)
	if redCmd.Err() != nil {
		log.Error(redCmd.Err())
		return nil, codes.Internal, redCmd.Err()
	}
	if err := oldRT.Scan(rt); err != nil {
		log.Error(redCmd.Err())
		return nil, codes.Internal, redCmd.Err()
	}
	return rt, codes.OK, nil
}

func (repo AuthorizationRepo) SetRefreshToken(ctx context.Context, rt *RefreshToken) (codes.Code, error) {
	log := logger.GetGrpcLogger(ctx)
	log.Info("HERERERERER")
	redCmd := repo.Redis.HSet(ctx, rt.ID,
		"user_id", rt.ID,
		"browser_fingerprint", rt.BrowserFingerprint,
	)
	log.Info("HERERERERER1")
	if redCmd.Err() != nil {
		log.Error(redCmd.Err())
		return codes.Internal, redCmd.Err()
	}
	repo.Redis.Expire(ctx, rt.ID, rtExpirationTime())
	return codes.OK, nil
}
