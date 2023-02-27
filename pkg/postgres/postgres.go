package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
)

// Mock Postgres interface stub generation.
//nolint:lll
//go:generate mockgen -destination=../mocks/mock_postgres.go -package=mocks github.com/surahman/FTeX/pkg/postgres Postgres

// Postgres is the interface through which the database can be accessed. Created to support mock testing.
type Postgres interface {
	// Open will create a connection pool and establish a connection to the database backend.
	Open() error

	// Close will shut down the connection pool and ensure that the connection to the database backend is terminated.
	Close() error

	// Healthcheck runs a ping healthcheck on the database backend.
	Healthcheck() error

	// Execute will execute statements or run a transaction on the database backend, leveraging the connection pool.
	Execute(func(Postgres, any) (any, error), any) (any, error)
}

// Check to ensure the Postgres interface has been implemented.
var _ Postgres = &postgresImpl{}

// postgresImpl implements the Postgres interface and contains the logic to interface with the database.
type postgresImpl struct {
	conf   config
	pool   *pgxpool.Pool
	logger *logger.Logger
}

// NewPostgres will create a new Postgres configuration by loading it.
func NewPostgres(fs *afero.Fs, logger *logger.Logger) (Postgres, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system or logger supplied")
	}

	return newPostgresImpl(fs, logger)
}

// newCPostgresImpl will create a new postgresImpl configuration and load it from disk.
func newPostgresImpl(fs *afero.Fs, logger *logger.Logger) (c *postgresImpl, err error) {
	c = &postgresImpl{conf: newConfig(), logger: logger}
	if err = c.conf.Load(*fs); err != nil {
		c.logger.Error("failed to load Postgres configurations from disk", zap.Error(err))

		return nil, err
	}

	return
}

// Open will start a database connection pool and establish a connection.
func (p *postgresImpl) Open() error {
	var err error
	if err = p.verifySession(); err == nil {
		return errors.New("connection is already established to Postgres")
	}

	var pgxConfig *pgxpool.Config

	if pgxConfig, err = pgxpool.ParseConfig(fmt.Sprintf(constants.GetPostgresDSN(),
		p.conf.Authentication.Username,
		p.conf.Authentication.Password,
		p.conf.Connection.Host,
		p.conf.Connection.Port,
		p.conf.Connection.Database,
		p.conf.Connection.Timeout)); err != nil {
		p.logger.Error("failed to parse Postgres DSN", zap.Error(err))

		return err
	}

	pgxConfig.MaxConns = p.conf.Pool.MaxConns
	pgxConfig.MinConns = p.conf.Pool.MinConns
	pgxConfig.HealthCheckPeriod = p.conf.Pool.HealthCheckPeriod

	if p.pool, err = pgxpool.NewWithConfig(context.Background(), pgxConfig); err != nil {
		p.logger.Error("failed to configure Postgres connection", zap.Error(err))

		return err
	}

	// Binary Exponential Backoff connection to Postgres. The lazy connection can be opened via a ping to the database.
	if err = p.createSessionRetry(); err != nil {
		return err
	}

	return nil
}

// Close will close the database connection pool.
func (p *postgresImpl) Close() (err error) {
	if err = p.verifySession(); err != nil {
		msg := "no established Postgres connection to close"
		p.logger.Error(msg)

		return errors.New(msg)
	}

	p.pool.Close()

	return
}

// Healthcheck will run a ping on the database to ascertain health.
func (p *postgresImpl) Healthcheck() (err error) {
	if err = p.verifySession(); err != nil {
		return
	}

	return p.pool.Ping(context.Background())
}

// Execute wraps the methods that create, read, update, and delete records from tables on the database.
func (p *postgresImpl) Execute(request func(Postgres, any) (any, error), params any) (any, error) {
	return request(p, params)
}

// verifySession will check to see if a session is established.
func (p *postgresImpl) verifySession() error {
	if p.pool == nil || p.pool.Ping(context.Background()) != nil {
		return errors.New("no session established")
	}

	return nil
}

// createSessionRetry will attempt to open the connection using binary exponential back-off.
// Stop on the first success or fail after the last one.
func (p *postgresImpl) createSessionRetry() (err error) {
	for attempt := 1; attempt <= p.conf.Connection.MaxConnAttempts; attempt++ {
		waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		p.logger.Info(fmt.Sprintf("Attempting connection to Postgres database in %s...", waitTime),
			zap.String("attempt", strconv.Itoa(attempt)))
		time.Sleep(waitTime)

		if err = p.pool.Ping(context.Background()); err == nil {
			return nil
		}
	}
	p.logger.Error("unable to establish connection to Postgres database", zap.Error(err))

	return
}
