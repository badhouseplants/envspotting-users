module github.com/badhouseplants/envspotting-users

go 1.16

require (
	github.com/badhouseplants/envspotting-go-proto v0.0.5-0.20210720214240-5a54bcbf4042
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis/v8 v8.11.0
	github.com/golang-migrate/migrate/v4 v4.14.1
	github.com/google/uuid v1.2.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/jackc/pgconn v1.9.0
	github.com/jackc/pgerrcode v0.0.0-20201024163028-a0d42d470451
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.12.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.8.1
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	google.golang.org/grpc v1.39.0
)
