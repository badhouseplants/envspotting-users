package service

import (
	"context"
	"errors"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	repo "github.com/badhouseplants/envspotting-users/repo/authorization"
	"github.com/badhouseplants/envspotting-users/third_party/redis"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/badhouseplants/envspotting-users/tools/token"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var authrepo repo.AuthorizationStore

var initRepo = func(ctx context.Context) repo.AuthorizationStore {
	if authrepo == nil {
		authrepo = repo.AuthorizationRepo{
			Redis:     redis.Client(ctx),
			CreatedAt: time.Now(),
		}
	}
	return authrepo
}

var (
	errUserIDNotProvided = errors.New("user id is not passed via metadata")
	errBFNotProvided     = errors.New("browser fingerprint is not passed via metadata")
)

// RefreshToken create a new pair of tokens and returns via metadata
func RefreshToken(ctx context.Context, in *common.EmptyMessage) (*common.EmptyMessage, error) {
	// _, err := token.RefreshTokens(ctx)
	// if err != nil {
	// return nil, err
	// }
	return &common.EmptyMessage{}, nil
}

func ParseIdFromToken(ctx context.Context, in *common.EmptyMessage) (*accounts.AccountId, error) {
	var (
		id     *accounts.AccountId
		userID string
		err    error
		log    = logger.GetGrpcLogger(ctx)
	)
	userID, err = token.ParseUserID(ctx)
	id = &accounts.AccountId{
		Id: userID,
	}
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return id, nil
}

func ValidateToken(ctx context.Context) (*common.EmptyMessage, error) {
	code, err := token.Validate(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return &common.EmptyMessage{}, nil
}

func GenerateToken(ctx context.Context, userID string) (*common.EmptyMessage, error) {
	initRepo(ctx)
	jwtToken, code, err := token.Generate(ctx, userID)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	browserFingerprint, code, err := getBrowserFingerprint(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	refreshToken := &repo.RefreshToken{
		ID:                 uuid.NewString(),
		BrowserFingerprint: browserFingerprint,
		UserID:             userID,
	}
	code, err = authrepo.SetRefreshToken(ctx, refreshToken)

	header := metadata.Pairs("jwt-token", jwtToken, "rt-token", refreshToken.ID)
	grpc.SendHeader(ctx, header)
	return &common.EmptyMessage{}, nil
}

func getBrowserFingerprint(ctx context.Context) (string, codes.Code, error) {
	userID := metautils.ExtractIncoming(ctx).Get("browser-fingerprint")
	if len(userID) == 0 {
		return "", codes.PermissionDenied, errBFNotProvided
	}
	return userID, codes.OK, nil
}

func getRefreshToken(ctx context.Context) (string, error) {
	oldRt := metautils.ExtractIncoming(ctx).Get("refresh-token-id")
	if len(oldRt) == 0 {
		return "", nil
	}
	return oldRt, nil
}

func getUserID(ctx context.Context) (string, codes.Code, error) {
	userID := metautils.ExtractIncoming(ctx).Get("user-id")
	if len(userID) == 0 {
		return "", codes.PermissionDenied, errUserIDNotProvided
	}
	return userID, codes.OK, nil
}
