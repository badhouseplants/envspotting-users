package migrations

import (
	"fmt"

	postgresClient "github.com/badhouseplants/envspotting-users/third_party/postgres"
	"github.com/badhouseplants/envspotting-users/tools/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate() {
	log := logger.GetServerLogger()
	params := postgresClient.NewConnectionParams()
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", params.Username, params.Password, params.Host, params.Port, params.Database)
	m, err := migrate.New(
		"file://migrations/scripts",
		connectionString)
	if err != nil {
		log.Error(err)
	}
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			log.Info(err)
		} else {
			log.Error(err)
		}
	}
}
