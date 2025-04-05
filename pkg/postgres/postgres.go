package postgres

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"go.uber.org/zap"
)

// Mock Postgres SQLC Querier interface stub generation. This is local to the Postgres package.
//go:generate mockgen -destination=querier_mocks.go -package=postgres github.com/surahman/FTeX/pkg/postgres Querier

// Mock Postgres interface stub generation.
//go:generate mockgen -destination=../mocks/mock_postgres.go -package=mocks github.com/surahman/FTeX/pkg/postgres Postgres

// Postgres is the interface through which the queries will be executed on the database.
//
//nolint:interfacebloat
type Postgres interface {
	// Open will establish a pooled connection to the Postgres database and ping it to ensure the connection is open.
	Open() error

	// Close will close an established Postgres pooled connection to support a healthy database.
	Close() error

	// Healthcheck will run a Ping command on an established Postgres database connection.
	Healthcheck() error

	// UserRegister will create a user account in the Postgres database.
	UserRegister(userAccount *modelsPostgres.UserAccount) (uuid.UUID, error)

	// UserCredentials will retrieve the ClientID and hashed password associated with a provided username.
	UserCredentials(username string) (uuid.UUID, string, error)

	// UserGetInfo will retrieve the account information associated with a Client ID.
	UserGetInfo(clientID uuid.UUID) (modelsPostgres.User, error)

	// UserDelete will delete the account information associated with a Client ID.
	UserDelete(clientID uuid.UUID) error

	// UserIsDeleted is the interface through which external methods can check if a user account is soft-deleted.
	UserIsDeleted(clientID uuid.UUID) (bool, error)

	// FiatCreateAccount will open an account associated with a Client ID for a specific currency.
	FiatCreateAccount(clientID uuid.UUID, ticker Currency) error

	// FiatExternalTransfer will transfer Fiat funds into an account associated with a Client ID for a specific
	// currency.
	FiatExternalTransfer(ctx context.Context, txDetails *FiatTransactionDetails) (*FiatAccountTransferResult, error)

	// FiatInternalTransfer will transfer Fiat funds for a specific Client ID between two Fiat currency accounts for
	// that client.
	FiatInternalTransfer(ctx context.Context, source *FiatTransactionDetails, destination *FiatTransactionDetails) (
		*FiatAccountTransferResult, *FiatAccountTransferResult, error)

	// FiatBalance is the interface through which external methods can retrieve a Fiat account balance for a specific
	// currency.
	FiatBalance(clientID uuid.UUID, ticker Currency) (FiatAccount, error)

	// FiatTxDetails is the interface through which external methods can retrieve a Fiat transaction details for a
	// specific transaction.
	FiatTxDetails(clientID uuid.UUID, txID uuid.UUID) ([]FiatJournal, error)

	// FiatBalancePaginated is the interface through which external methods can retrieve all Fiat account balances
	// for a specific client.
	FiatBalancePaginated(clientID uuid.UUID, ticker Currency, pageSize int32) ([]FiatAccount, error)

	// FiatTransactionsPaginated is the interface through which external methods can retrieve transactions on a
	// Fiat account for a specific client during a specific month.
	FiatTransactionsPaginated(clientID uuid.UUID, ticker Currency, pageSize int32, offset int32,
		start pgtype.Timestamptz, end pgtype.Timestamptz) ([]FiatJournal, error)

	// CryptoCreateAccount is the interface through which external methods can create a Crypto account.
	CryptoCreateAccount(clientID uuid.UUID, ticker string) error

	// CryptoBalance is the interface through which external methods can retrieve a Fiat-account balance for a specific
	// cryptocurrency.
	CryptoBalance(clientID uuid.UUID, ticker string) (CryptoAccount, error)

	// CryptoTxDetails is the interface through which external methods can retrieve a Crypto transaction details for a
	// specific transaction.
	CryptoTxDetails(clientID uuid.UUID, txID uuid.UUID) ([]CryptoJournal, error)

	// CryptoPurchase is the interface through which external methods can purchase a specific Cryptocurrency.
	CryptoPurchase(clientID uuid.UUID, fiatTicker Currency, fiatAmount decimal.Decimal, cryptoTicker string,
		cryptoAmount decimal.Decimal) (*FiatJournal, *CryptoJournal, error)

	// CryptoSell is the interface through which external methods can sell a specific Cryptocurrency.
	CryptoSell(clientID uuid.UUID, fiatTicker Currency, fiatAmount decimal.Decimal, cryptoTicker string,
		cryptoAmount decimal.Decimal) (*FiatJournal, *CryptoJournal, error)

	// CryptoBalancesPaginated is the interface through which external methods can retrieve all Crypto account balances
	// for a specific client.
	CryptoBalancesPaginated(clientID uuid.UUID, ticker string, pageSize int32) ([]CryptoAccount, error)

	// CryptoTransactionsPaginated is the interface through which external methods can retrieve transactions on a Crypto
	// account for a specific client during a specific month.
	CryptoTransactionsPaginated(clientID uuid.UUID, cryptoTicker string, pageSize int32, offset int32,
		start pgtype.Timestamptz, end pgtype.Timestamptz) ([]CryptoJournal, error)
}

// Check to ensure the Postgres interface has been implemented.
var _ Postgres = &postgresImpl{}

// postgresImpl contains objects required to interface with the database.
type postgresImpl struct {
	conf    config
	pool    *pgxpool.Pool
	logger  *logger.Logger
	queries *Queries
	Query   Querier
}

// NewPostgres will create a new Postgres configuration by loading it.
func NewPostgres(fs *afero.Fs, logger *logger.Logger) (Postgres, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system or logger supplied")
	}

	return newPostgresImpl(fs, logger)
}

// newPostgresImpl will create a new postgresImpl configuration and load it from disk.
func newPostgresImpl(fs *afero.Fs, logger *logger.Logger) (*postgresImpl, error) {
	if fs == nil || logger == nil {
		return nil, errors.New("nil file system or logger supplied")
	}

	postgres := &postgresImpl{conf: newConfig(), logger: logger}
	if err := postgres.conf.Load(*fs); err != nil {
		postgres.logger.Error("failed to load Postgres configurations from disk", zap.Error(err))

		return nil, err
	}

	return postgres, nil
}

// Open will start a database connection pool and establish a connection.
func (p *postgresImpl) Open() error {
	var err error
	if err = p.verifySession(); err == nil {
		return errors.New("connection is already established to Postgres")
	}

	var pgxConfig *pgxpool.Config

	if pgxConfig, err = pgxpool.ParseConfig(fmt.Sprintf(constants.PostgresDSN(),
		p.conf.Authentication.Username,
		p.conf.Authentication.Password,
		p.conf.Connection.Host,
		p.conf.Connection.Port,
		p.conf.Connection.Database,
		p.conf.Connection.Timeout)); err != nil {
		msg := "failed to parse Postgres DSN"
		p.logger.Error(msg, zap.Error(err))

		return fmt.Errorf(msg+"%w", err)
	}

	pgxConfig.MaxConns = p.conf.Pool.MaxConns
	pgxConfig.MinConns = p.conf.Pool.MinConns
	pgxConfig.HealthCheckPeriod = p.conf.Pool.HealthCheckPeriod

	if p.pool, err = pgxpool.NewWithConfig(context.Background(), pgxConfig); err != nil {
		msg := "failed to configure Postgres connection"
		p.logger.Error(msg, zap.Error(err))

		return fmt.Errorf(msg+"%w", err)
	}

	// Binary Exponential Backoff connection to Postgres. The lazy connection can be opened via a ping to the database.
	if err = p.createSessionRetry(); err != nil {
		return err
	}

	// Setup SQLC DBTX interface.
	p.queries = New(p.pool)
	p.Query = p.queries

	return nil
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
func (p *postgresImpl) Healthcheck() error {
	var err error
	if err = p.verifySession(); err != nil {
		return err
	}

	if err = p.pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("postgres cluster ping failed: %w", err)
	}

	return nil
}
