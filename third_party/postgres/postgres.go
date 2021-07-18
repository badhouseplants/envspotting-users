package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/badhouseplants/envspotting-users/tools/logger"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
)

// TODO: add context here

// ConnectionParams to open database pool
type ConnectionParams struct {
	Username string
	Password string
	Database string
	Host     string
	Port     string
}

func NewConnectionParams() *ConnectionParams {
	return &ConnectionParams{
		Username: viper.GetString("database_username"),
		Password: viper.GetString("database_password"),
		Database: viper.GetString("database_name"),
		Host:     viper.GetString("database_host"),
		Port:     viper.GetString("database_port"),
	}
}

var (
	pool    *pgxpool.Pool
	retries = 5
)

// Pool return db coonection pool
func Pool() *pgxpool.Pool {
	var log = logger.GetServerLogger()
	if pool == nil {
		err := OpenConnectionPool()
		if err != nil {
			log.Fatal(err)
			return nil
		}
	}
	return pool
}

// OpenConnectionPool opens new connection pool
func OpenConnectionPool() (err error) {
	params := NewConnectionParams() // TODO: Refactor
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", params.Username, params.Password, params.Host, params.Port, params.Database)
	for i := 0; i < retries; i++ {
		pool, err = pgxpool.Connect(context.Background(), connectionString)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return err
	}

	return nil
}
