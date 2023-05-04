package utilities

import (
	"log"
	"os"
	"testing"

	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
)

// testAuth is the Authorization object.
var testAuth auth.Auth

// expirationDuration is the time in seconds that a JWT will be valid for.
var expirationDuration int64 = 10

// refreshThreshold is the time in seconds before expiration that a JWT can be refreshed in.
var refreshThreshold int64 = 99

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

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
//
//nolint:unparam
func setup() error {
	testAuth = auth.TestAuth(zapLogger, expirationDuration, refreshThreshold)

	return nil
}

// tearDown will delete the test clusters keyspace.
func tearDown() error {
	return nil
}
