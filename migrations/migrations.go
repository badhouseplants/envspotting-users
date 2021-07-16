package migrations

import (
	"fmt"

	postgresClient "github.com/badhouseplants/envspotting-users/third_party/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

func Migrate() {
	params := postgresClient.NewConnectionParams()
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", params.Username, params.Password, params.Host, params.Port, params.Database)
	// fsrc, err := (&file.File{}).Open("file://migrations")
	// if err != nil {
		// panic(err)
	// }

	m, err := migrate.New(
		"file://migrations/",
		connectionString)
	if err != nil {
		panic(err)
	}
	m.Steps(2)
}
