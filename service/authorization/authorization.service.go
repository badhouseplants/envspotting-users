package service

import (
	"context"

	"github.com/badhouseplants/envspotting-users/tools/token"
	"github.com/badhouseplants/envspotting-go-proto/models/common"
)

// Create the JWT key used to create the signature
func RefreshToken(ctx context.Context, in *common.EmptyMessage) (*common.EmptyMessage, error) {
	_, err := token.RefreshTokens(ctx)
	if err != nil {
		return nil, err
	}
	return &common.EmptyMessage{}, nil
}

func GenerateToken(ctx context.Context, userID string) (*common.EmptyMessage, error) {
	_, err := token.Generate(ctx, userID)
	if err != nil { 
		return nil, err
	}
	return &common.EmptyMessage{}, nil
}
