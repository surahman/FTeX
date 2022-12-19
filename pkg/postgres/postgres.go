package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
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

// NewPostgres will create a new Postgres configuration by loading it.
func NewPostgres(fs *afero.Fs, logger *logger.Logger) (Postgres, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system or logger supplied")
	}
	return newPostgresImpl(fs, logger)
}

// newCassandraImpl will create a new CassandraImpl configuration and load it from disk.
func newPostgresImpl(fs *afero.Fs, logger *logger.Logger) (c *postgresImpl, err error) {
	c = &postgresImpl{conf: newConfig(), logger: logger}
	if err = c.conf.Load(*fs); err != nil {
		c.logger.Error("failed to load Postgres configurations from disk", zap.Error(err))
		return nil, err
	}
	return
}

// Open will start a database connection pool and establish a connection.
func (p *postgresImpl) Open() (err error) {
	return
}

// Close will close the database connection pool.
func (p *postgresImpl) Close() (err error) {
	return
}

// Ping will ping the database using the connection pool.
func (p *postgresImpl) Ping() (err error) {
	return
}

// Healthcheck will run a lightweight query on the database to ascertain health.
func (p *postgresImpl) Healthcheck() (err error) {
	return
}

// Execute wraps the methods that create, read, update, and delete records from tables on the database.
func (p *postgresImpl) Execute(request func(Postgres, any) (any, error), params any) (any, error) {
	return request(p, params)
}

// verifySession will check to see if a session is established.
func (p *postgresImpl) verifySession() error {
	if p.session == nil || p.session.IsClosed() {
		return errors.New("no session established")
	}
	return nil
}
