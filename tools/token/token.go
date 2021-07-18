package token

import (
	"context"
	"fmt"

	"github.com/badhouseplants/envspotting-users/third_party/redis"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"time"

	"github.com/spf13/viper"
)

var (
	jwtSecret = []byte(viper.GetString("jwt_secret"))
	// TODO: put in environment @allanger
	jwtExpirationTime time.Time
	rtExpirationTime  = 24 * 7 * time.Hour
)

type JWTClaims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}

type RefreshToken struct {
	ID                 string
	BrowserFingerprint string `redis:"browser_fingerprint"`
	UserID             string `redis:"user_id"`
}

type Tokens struct {
	JWT string
	// RT UUID
	RT string
}

func Generate(ctx context.Context, userID string) (*Tokens, error) {
	jwtExpirationTime = time.Now().Add(10000 * time.Second)
	tokens := &Tokens{}
	if err := tokens.NewJWT(userID); err != nil {
		return nil, err
	}

	if err := tokens.NewRT(ctx, userID); err != nil {
		return nil, err
	}
	// Creating refresh token
	header := metadata.Pairs("jwt-token", tokens.JWT, "rt-token", tokens.RT)
	grpc.SendHeader(ctx, header)
	return tokens, nil
}

func Validate(ctx context.Context) error {
	tknStr := metautils.ExtractIncoming(ctx).Get("authorization")
	// Initialize a new instance of `Claims`
	claims := &JWTClaims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return status.Error(codes.Unauthenticated, err.Error())

		}
		return status.Error(codes.Aborted, err.Error())

	}
	if !tkn.Valid {
		return status.Error(codes.Unauthenticated, err.Error())

	}
	return nil
}

func ParseUserID(ctx context.Context) (string, error) {
	tknStr := metautils.ExtractIncoming(ctx).Get("authorization")
	hmacSecretString := jwtSecret // Value
	hmacSecret := []byte(hmacSecretString)
	token, _ := jwt.Parse(tknStr, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := fmt.Sprintf("%v", claims["userId"])
		return userID, nil
	} else {
		return "", status.Error(codes.PermissionDenied, ("wrong jwt token"))
	}
}

func (t *Tokens) NewJWT(userID string) (err error) {
	jwtClaims := &JWTClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwtExpirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	t.JWT, err = token.SignedString(jwtSecret)
 	if err != nil {
		return err
	}
	return nil
}

func (t *Tokens) NewRT(ctx context.Context, userID string) (err error) {
	browserFingerprint, err := getBrowserFingerprint(ctx)
	if err != nil {
		return err
	}
	id := uuid.NewString()
	// Save to redis
	r := redis.Client()
	redCmd := r.HSet(ctx, id,
		"user_id", userID,
		"browser_fingerprint", browserFingerprint,
	)

	if redCmd.Err() != nil {
		return redCmd.Err()
	}
	r.Expire(ctx, id, rtExpirationTime)
	t.RT = id
	return nil
}

func ValidateJWT() {}

func RefreshTokens(ctx context.Context) (*Tokens, error) {
	r := redis.Client()
	rt := &RefreshToken{}
	rtID, err := getRefreshToken(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	userID, err := getUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	oldRT := r.HGetAll(ctx, rtID)
	r.Del(ctx, rtID)
	oldRT.Scan(rt)
	browserFingerprint, err := getBrowserFingerprint(ctx)
	if err != nil {
		return nil, nil
	} else if userID != rt.UserID {
		fmt.Println(rt.UserID)
		fmt.Println(userID)
		return nil, status.Error(codes.PermissionDenied, "refresh token isn't owned by this user")
	} else if browserFingerprint != rt.BrowserFingerprint {
		// TODO: fix error message @allanger
		return nil, status.Error(codes.PermissionDenied, "suspicious activity, browser fingerprint is wrong")
	} else {
		return Generate(ctx, userID)
	}
}

func getBrowserFingerprint(ctx context.Context) (string, error) {
	return "finger", nil
}

func getRefreshToken(ctx context.Context) (string, error) {
	oldRt := metautils.ExtractIncoming(ctx).Get("refresh-token-id")
	if len(oldRt) == 0 {
		return "", nil
	}
	return oldRt, nil
}

func getUserID(ctx context.Context) (string, error) {
	userID := metautils.ExtractIncoming(ctx).Get("user-id")
	if len(userID) == 0 {
		return "", nil
	}
	return userID, nil
}
