package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/surahman/FTeX/pkg/logger"
)

// Mock Cassandra interface stub generation.
//go:generate mockgen -destination=../mocks/mock_postgres.go -package=mocks github.com/surahman/FTeX/pkg/postgres Postgres

// Postgres is the interface through which the database can be accessed. Created to support mock testing.
type Postgres interface {
	// Open will create a connection pool and establish a connection to the database backend.
	Open() error

	// Close will shut down the connection pool and ensure that the connection to the database backend is terminated correctly.
	Close() error

	// Ping will pint the database. This can be used to open an initial connection support by some database drivers that operate
	// with lazy connections.
	Ping() error

	// Healthcheck runs a lightweight healthcheck on the database backend.
	Healthcheck() error

	// Execute will execute statements or run a lightweight transaction on the database backend, leveraging the connection pool.
	Execute(func(Postgres, any) (any, error), any) (any, error)
}

// Check to ensure the Postgres interface has been implemented.
var _ Postgres = &postgresImpl{}

// postgresImpl implements the Postgres interface and contains the logic to interface with the database.
type postgresImpl struct {
	conf    *config
	session *pgx.Conn
	logger  *logger.Logger
}
