package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/badhouseplants/envspotting-go-proto/models/common"
	"github.com/badhouseplants/envspotting-go-proto/models/users/accounts"
	repo "github.com/badhouseplants/envspotting-users/repo/authorization"
	"github.com/badhouseplants/envspotting-users/third_party/redis"
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
	errUserIDNotProvided       = errors.New("user id is not passed via metadata (user-id header)")
	errBFNotProvided           = errors.New("browser fingerprint is not passed via metadata (browser-fingerprint header)")
	errAuthTokenNotProvided    = errors.New("jwt token is not provided via metadata (authorization header)")
	errRefreshTokenNotProvided = errors.New("refresh token is not provided via metadata (refresh-token header)")
	errStrangeActivity         = errors.New("strange activity (wrong browser-fingerprint is provided)")
	errRefreshTokenNotOwned    = errors.New("refresh token is not owned by this user")
)

// RefreshToken create a new pair of tokens and returns via metadata
func RefreshToken(ctx context.Context, userID *accounts.AccountId) (*common.EmptyMessage, error) {
	var (
		err  error
		code codes.Code
	)
	initRepo(ctx)
	tknStr, code, err := getRefreshToken(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	bf, code, err := getBrowserFingerprint(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}

	rt := &repo.RefreshToken{
		ID: tknStr,
	}
	_, code, err = authrepo.GetRefreshToken(ctx, rt)

	switch {
	case err != nil:
		return nil, status.Error(code, err.Error())
	case rt.BrowserFingerprint != bf:
		return nil, status.Error(codes.PermissionDenied, errStrangeActivity.Error())
	case rt.UserID != userID.Id:
		fmt.Printf("%s - %s", rt.UserID, userID)
		return nil, status.Error(codes.PermissionDenied, errRefreshTokenNotOwned.Error())
	default:
		code, err = authrepo.DelRefreshToken(ctx, rt)
		if err != nil {
			return nil, status.Error(code, err.Error())
		}
		_, err = GenerateToken(ctx, userID.Id)
		if err != nil {
			return nil, err
		}
	}
	return &common.EmptyMessage{}, nil
}

func ParseIdFromToken(ctx context.Context, in *common.EmptyMessage) (*accounts.AccountId, error) {
	var (
		id     *accounts.AccountId
		userID string
		tknStr string
		err    error
		code   codes.Code
	)
	tknStr, code, err = GetAuthorizationToken(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	userID, code, err = token.ParseUserID(ctx, tknStr)
	id = &accounts.AccountId{
		Id: userID,
	}
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	return id, nil
}

func ValidateToken(ctx context.Context) (*common.EmptyMessage, error) {
	tknStr, code, err := GetAuthorizationToken(ctx)
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	code, err = token.Validate(ctx, tknStr)
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
	if err != nil {
		return nil, status.Error(code, err.Error())
	}
	header := metadata.Pairs("authorization", jwtToken, "refresh-token", refreshToken.ID)
	grpc.SendHeader(ctx, header)
	return &common.EmptyMessage{}, nil
}

func getBrowserFingerprint(ctx context.Context) (string, codes.Code, error) {
	bf := metautils.ExtractIncoming(ctx).Get("browser-fingerprint")
	if len(bf) != 0 {
		return bf, codes.OK, nil
	}

	return "", codes.PermissionDenied, errBFNotProvided
}

func getRefreshToken(ctx context.Context) (string, codes.Code, error) {
	oldRt := metautils.ExtractIncoming(ctx).Get("refresh-token")
	if len(oldRt) == 0 {
		return "", codes.PermissionDenied, errRefreshTokenNotProvided
	}
	return oldRt, codes.OK, nil
}

func getUserID(ctx context.Context) (string, codes.Code, error) {
	userID := metautils.ExtractIncoming(ctx).Get("user-id")
	if len(userID) == 0 {
		return "", codes.PermissionDenied, errUserIDNotProvided
	}
	return userID, codes.OK, nil
}

func GetAuthorizationToken(ctx context.Context) (string, codes.Code, error) {
	tknStr := metautils.ExtractIncoming(ctx).Get("authorization")
	if len(tknStr) == 0 {
		return "", codes.PermissionDenied, errAuthTokenNotProvided
	}
	return tknStr, codes.OK, nil
}
