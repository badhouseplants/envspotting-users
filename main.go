package main

import (
	"fmt"
	"net"

	"github.com/badhouseplants/envspotting-users/migrations"
	accounts "github.com/badhouseplants/envspotting-users/service/accounts"
	authentication "github.com/badhouseplants/envspotting-users/service/authentication"
	authorization "github.com/badhouseplants/envspotting-users/service/authorization"
	rights "github.com/badhouseplants/envspotting-users/service/rights"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

var (
	host string
)

func init() {
	// app variables
	viper.SetDefault("app_host", "0.0.0.0")
	viper.SetDefault("app_port", "9090")
	viper.SetDefault("database_username", "user")
	viper.SetDefault("database_password", "qwertyu9")
	viper.SetDefault("database_name", "aggregator")
	viper.SetDefault("database_host", "localhost")
	viper.SetDefault("database_port", "5432")
	viper.AutomaticEnv() // read in environment variables that match)
}

func main() {
	// log := logger.GetServerLogger()
	migrations.Migrate()
	Serve()
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
	// Disable on prod env
	reflection.Register(grpcServer)
}

func Serve() {
	log := logger.GetServerLogger()
	// seting up grpc server
	listener, err := net.Listen("tcp", getHost())
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer(
		setupGrpcStreamOpts(),
		setupGrpcUnaryOpts(),
	)
	registerServices(grpcServer)
	log.Infof("starting to serve on %s", getHost())
	grpcServer.Serve(listener)
}

func getHost() string {
	host = fmt.Sprintf("%s:%s", viper.GetString("app_host"), viper.GetString("app_port"))
	return host
}
