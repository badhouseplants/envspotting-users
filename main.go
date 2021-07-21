package main

import (
	"github.com/badhouseplants/envspotting-users/migrations"
	"github.com/badhouseplants/envspotting-users/server"
	"github.com/spf13/viper"
)

const AppVersion = "0.0.1"

func init() {
	// app variables
	viper.SetDefault("environment", "dev")
	// server variables
	viper.SetDefault("envspotting_users_host", "0.0.0.0")
	viper.SetDefault("envspotting_users_port", "9090")
	// database variables
	viper.SetDefault("database_username", "docker_user")
	viper.SetDefault("database_password", "qwertyu9")
	viper.SetDefault("database_name", "users")
	viper.SetDefault("database_host", "localhost")
	viper.SetDefault("database_port", "5432")
	// redis variables
	viper.SetDefault("redis_host", "localhost:6379")
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", "0")
	// auth tokens variables
	viper.SetDefault("refresh_token_expiry", "24") // hours
	// read environment variables that match
	viper.AutomaticEnv()
}

func main() {
	if err := migrations.Migrate(); err != nil {
		panic(err)
	}
	if err := server.Serve(); err != nil {
		panic(err)
	}
}
