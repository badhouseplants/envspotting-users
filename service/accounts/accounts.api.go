package accserv

import (
	"github.com/badhouseplants/envspotting-users/models/users/accounts"
	"google.golang.org/grpc"
)

type serviceGrpcImpl struct {
	accounts.UnimplementedAccountsServer
}

func Register(grpcServer *grpc.Server) {
	accounts.RegisterAccountsServer(grpcServer, &serviceGrpcImpl{})
}
