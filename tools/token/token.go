package token

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/codes"

	"time"

	"github.com/spf13/viper"
)

var (
	jwtSecret = []byte(viper.GetString("jwt_secret"))
	// TODO: put in environment @allanger
)

type JWTClaims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}

func Generate(ctx context.Context, userID string) (string, codes.Code, error) {
	// FIXME: time
	jwtExpirationTime := time.Now().Add( viper.GetDuration("jwt_token_expiry") * time.Minute)
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
	authStr := fmt.Sprintf("Bearer: %s", tknStr)
	return authStr, codes.OK, nil
}

func Validate(ctx context.Context, tknStr string) (codes.Code, error) {
	tknStr = strings.ReplaceAll(tknStr, "Bearer: ", "")
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

func ParseUserID(ctx context.Context, authStr string) (string, codes.Code, error) {
	tknStr := strings.ReplaceAll(authStr, "Bearer: ", "")
	hmacSecretString := jwtSecret // Value
	hmacSecret := []byte(hmacSecretString)
	token, _ := jwt.Parse(tknStr, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := fmt.Sprintf("%v", claims["userId"])
		return userID, codes.OK, nil
	} else {
		return "", codes.PermissionDenied, errors.New("wrong jwt token is provided")
	}
}
