package constants

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEtcDir(t *testing.T) {
	require.Equal(t, configEtcDir, GetEtcDir(), "Incorrect etc directory")
}

func TestGetHomeDir(t *testing.T) {
	require.Equal(t, configHomeDir, GetHomeDir(), "Incorrect home directory")
}

func TestGetBaseDir(t *testing.T) {
	require.Equal(t, configBaseDir, GetBaseDir(), "Incorrect base directory")
}

func TestGetGithubCIKey(t *testing.T) {
	require.Equal(t, githubCIKey, GetGithubCIKey(), "Incorrect Github CI environment key")
}

func TestGetLoggerFileName(t *testing.T) {
	require.Equal(t, loggerConfigFileName, GetLoggerFileName(), "Incorrect logger filename")
}

func TestGetLoggerPrefix(t *testing.T) {
	require.Equal(t, loggerPrefix, GetLoggerPrefix(), "Incorrect Zap logger environment prefix")
}

func TestGetPostgresFileName(t *testing.T) {
	require.Equal(t, postgresConfigFileName, GetPostgresFileName(), "Incorrect Postgres filename")
}

func TestGetPostgresPrefix(t *testing.T) {
	require.Equal(t, postgresPrefix, GetPostgresPrefix(), "Incorrect Postgres environment prefix")
}

func TestGetPostgresDSN(t *testing.T) {
	require.Equal(t, postgresDSN, GetPostgresDSN(), "Incorrect Postgres DSN format string")
}

func TestGetTestDatabaseName(t *testing.T) {
	require.Equal(t, testDatabaseName, GetTestDatabaseName(), "Incorrect test suite database name")
}

func TestGetRedisFileName(t *testing.T) {
	require.Equal(t, redisConfigFileName, GetRedisFileName(), "Incorrect Redis filename")
}

func TestGetRedisPrefix(t *testing.T) {
	require.Equal(t, redisPrefix, GetRedisPrefix(), "Incorrect Redis prefix")
}

func TestGetQuotesFileName(t *testing.T) {
	require.Equal(t, quotesConfigFileName, GetQuotesFileName(), "Incorrect Quotes filename")
}

func TestGetQuotesPrefix(t *testing.T) {
	require.Equal(t, quotesPrefix, GetQuotesPrefix(), "Incorrect Quotes prefix")
}

func TestGetAuthFileName(t *testing.T) {
	require.Equal(t, authConfigFileName, GetAuthFileName(), "Incorrect authentication filename")
}

func TestGetAuthPrefix(t *testing.T) {
	require.Equal(t, authPrefix, GetAuthPrefix(), "Incorrect authorization environment prefix")
}

func TestGetHTTPRESTFileName(t *testing.T) {
	require.Equal(t, restConfigFileName, GetHTTPRESTFileName(), "Incorrect HTTP REST filename")
}

func TestGetHTTPRESTPrefix(t *testing.T) {
	require.Equal(t, restPrefix, GetHTTPRESTPrefix(), "Incorrect HTTP REST environment prefix")
}

func TestGetDeleteUserAccountConfirmation(t *testing.T) {
	require.Equal(t, deleteUserAccountConfirmation, GetDeleteUserAccountConfirmation(),
		"Incorrect user account deletion confirmation message.")
}

func TestGetDecimalPlacesFiat(t *testing.T) {
	require.Equal(t, fiatDecimalPlaces, GetDecimalPlacesFiat(), "Incorrect Fiat currency decimal places.")
}

func TestGetFiatOfferTTL(t *testing.T) {
	require.Equal(t, fiatOfferTTL, GetFiatOfferTTL(), "Incorrect Fiat offer TTL.")
}

func TestGetCryptoOfferTTL(t *testing.T) {
	require.Equal(t, cryptoOfferTTL, GetCryptoOfferTTL(), "Incorrect Crypto offer TTL.")
}
