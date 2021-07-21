package token

import (
	"context"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"time"

	"github.com/spf13/viper"
)

var (
	jwtSecret = []byte(viper.GetString("jwt_secret"))
	// TODO: put in environment @allanger
	jwtExpirationTime time.Time
)

type JWTClaims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}

func Generate(ctx context.Context, userID string) (string, codes.Code, error) {
	// FIXME: time
	jwtExpirationTime = time.Now().Add(10000 * time.Second)
	jwtClaims := &JWTClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwtExpirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	tknStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", codes.Internal, err
	}
	return tknStr, codes.OK, nil
}

func Validate(ctx context.Context) (codes.Code, error) {
	tknStr := metautils.ExtractIncoming(ctx).Get("authorization")
	// Initialize a new instance of `Claims`
	claims := &JWTClaims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return codes.Unauthenticated, err
		}
		return codes.Aborted, err
	}
	if !tkn.Valid {
		return codes.Unauthenticated, err
	}
	return codes.OK, nil
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
		return "", codes.PermissionDenied, ("wrong jwt
	}
}

// func RefreshTokens(ctx context.Context) (*Tokens, error) {
	// r := redis.Client()
	// rt := &RefreshToken{}
	// rtID, err := getRefreshToken(ctx)
	// if err != nil {
		// return nil, codes.PermissionDenied, err
	// }
	// userID, err := getUserID(ctx)
	// if err != nil {
		// return nil, codes.PermissionDenied, err
	// }
	// oldRT := r.HGetAll(ctx, rtID)
	// r.Del(ctx, rtID)
	// oldRT.Scan(rt)
	// browserFingerprint, err := getBrowserFingerprint(ctx)
	// if err != nil {
		// return nil, nil
	// } else if userID != rt.UserID {
		// fmt.Println(rt.UserID)
		// fmt.Println(userID)
		// return nil, codes.PermissionDenied, "refresh token isn't owned by this
	// } else if browserFingerprint != rt.BrowserFingerprint {
		// TODO: fix error message @allanger
		// return nil, codes.PermissionDenied, "suspicious activity, browser fingerprint is
	// } else {
		// return Generate(ctx, userID)
	// }
// }
// 