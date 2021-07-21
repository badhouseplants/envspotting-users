package server

import (
	"fmt"
	"net"

	"github.com/badhouseplants/envspotting-users/tools/logger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/spf13/viper"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	accounts "github.com/badhouseplants/envspotting-users/service/accounts"
	authentication "github.com/badhouseplants/envspotting-users/service/authentication"
	authorization "github.com/badhouseplants/envspotting-users/service/authorization"
	rights "github.com/badhouseplants/envspotting-users/service/rights"
)

// Serve starts grpc server
func Serve() error {
	log := logger.GetServerLogger()
	listener, err := net.Listen("tcp", getHost())
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer(
		setupGrpcStreamOpts(),
		setupGrpcUnaryOpts(),
	)
	registerServices(grpcServer)
	log.Infof("starting to serve on %s", getHost())
	if err = grpcServer.Serve(listener); err != nil {
		return err
	}
	return nil
}

func getHost() string {
	return fmt.Sprintf("%s:%s", viper.GetString("envspotting_users_host"), viper.GetString("envspotting_users_port"))
}

func setupGrpcUnaryOpts() grpc.ServerOption {
	return grpc_middleware.WithUnaryServerChain(
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.UnaryServerInterceptor(logger.GrpcLogrusEntry, logger.GrpcLogrusOpts...),
	)
}

func setupGrpcStreamOpts() grpc.ServerOption {
	return grpc_middleware.WithStreamServerChain(
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.StreamServerInterceptor(logger.GrpcLogrusEntry, logger.GrpcLogrusOpts...),
	)
}

func registerServices(grpcServer *grpc.Server) {
	accounts.Register(grpcServer)
	authentication.Register(grpcServer)
	authorization.Register(grpcServer)
	rights.Register(grpcServer)
	if viper.GetString("environment") != "prod" {
		reflection.Register(grpcServer)
	}
}
