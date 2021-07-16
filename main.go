package main

import (
	"fmt"

	"github.com/badhouseplants/envspotting-users/migrations"
	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("database_username", "user")
	viper.SetDefault("database_password", "qwertyu9")
	viper.SetDefault("database_name", "aggregator")
	viper.SetDefault("database_host", "localhost")
	viper.SetDefault("database_port", "5432")
	migrations.Migrate()

	fmt.Println("asd")
}
