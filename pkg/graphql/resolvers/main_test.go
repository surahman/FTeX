package graphql

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
)

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

// testAuthHeaderKey is the authorization header key.
var testAuthHeaderKey = "Authorization"

// testUserQuery is the test user account related mutations and queries.
var testUserQuery = getUsersQuery()

// testFiatQuery is the test Fiat related mutations and queries.
var testFiatQuery = getFiatQuery()

// testCryptoQuery is the test Crypto related mutations and queries.
var testCryptoQuery = getCryptoQuery()

func TestMain(m *testing.M) {
	var err error
	// Configure logger.
	if zapLogger, err = logger.NewTestLogger(); err != nil {
		log.Printf("Test suite logger setup failed: %v\n", err)
		os.Exit(1)
	}

	// Setup test space.
	if err = setup(); err != nil {
		zapLogger.Error("Test suite setup failure", zap.Error(err))
		os.Exit(1)
	}

	// Run test suite.
	exitCode := m.Run()

	// Cleanup test space.
	if err = tearDown(); err != nil {
		zapLogger.Error("Test suite teardown failure:", zap.Error(err))
		os.Exit(1)
	}

	os.Exit(exitCode)
}

// setup will configure the auth test object.
func setup() error {
	return nil
}

// tearDown will delete the test clusters keyspace.
func tearDown() error {
	return nil
}

// verifyErrorReturned will check an HTTP response to ensure an error was returned.
func verifyErrorReturned(t *testing.T, response map[string]any) {
	t.Helper()

	value, ok := response["data"]
	require.True(t, ok, "data key expected but not set")
	require.Nil(t, value, "data value should be set to nil")

	value, ok = response["errors"]
	require.True(t, ok, "error key expected but not set")
	require.NotNil(t, value, "error value should not be nil")
}
