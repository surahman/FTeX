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
