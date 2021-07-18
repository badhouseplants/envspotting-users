package accserv

import (
	"context"
	"time"

	"github.com/badhouseplants/envspotting-users/models/users/accounts"
	accrepo "github.com/badhouseplants/envspotting-users/repo/accounts"
	"github.com/badhouseplants/envspotting-users/third_party/postgres"
	"github.com/badhouseplants/envspotting-users/tools/hasher"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var initRepo = func() *accrepo.AccountRepo {
	return &accrepo.AccountRepo{
		Pool:      postgres.Pool(),
		CreatedAt: time.Now(),
	}
}

// Create a new user
func (s *serviceGrpcImpl) Create(ctx context.Context, in *accounts.AccountCreds) (*accounts.AccountInfo, error) {
	repo := initRepo()
	log := logger.EnpointHit(ctx)
	// Fill user struct
	id := uuid.New().String()
	user := &accounts.AccountInfoWithSensitive{
		Id:       id,
		Username: in.GetUsername(),
		Password: hasher.Encrypt(in.GetPassword()),
	}
	in.Reset()

	// Clear sensitive data
	log.Debug("calling repo create method")

	code, err := repo.Create(ctx, user)
	if code != codes.OK {
		return nil, status.Error(code, err.Error())
	}

	out := &accounts.AccountInfo{
		Id:       user.Id,
		Username: user.Username,
	}
	return out, nil
}

func (s *serviceGrpcImpl) UpdateUser(ctx context.Context, in *accounts.AccountInfo) (*accounts.AccountInfo, error) {
	repo := initRepo()
	// log := logger.EnpointHit(ctx)
	// idFromToken, err := token.ParseUserID(ctx)
	// if err != nil {
		// log.Error(err)
		// return nil, err
	// } else if in.Id != idFromToken {
		// return nil, status.Error(codes.PermissionDenied, "do not tru to update other users")
	// }
	_, err := repo.UpdateUser(ctx, in)
	if err != nil {
		return nil, err
	}
	return in, nil

}
