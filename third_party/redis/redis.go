// TODO: Refactor this peace of crap! @allanger
package redis

import (
	"context"

	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var client *redis.Client

// Client returns a redis client
func Client(ctx context.Context) *redis.Client {
	log := logger.GetServerLogger()
	if client == nil {
		NewClient()
	}
	status := client.Ping(ctx)
	if status.Err() != nil {
		log.Error(status.Err())
	}
	return client
}

// NewClient inits a redis client
func NewClient() {
	client = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis_host"),
		Password: viper.GetString("redis_password"),
		DB:       viper.GetInt("redis_database"),
	})
}
