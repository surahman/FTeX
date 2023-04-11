package quotes

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
)

// quotesConfigTestData is a map Quotes configuration test data.
var quotesConfigTestData = configTestData()

// testConfigs is the name of the Quotes configuration to use in the tests.
var testConfigs string

// zapLogger is the Zap logger used strictly for the test suite in this package.
var zapLogger *logger.Logger

func TestMain(m *testing.M) {
	// Parse commandline flags to check for short tests.
	flag.Parse()

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

// setup will configure the connection to the test database.
func setup() error {
	if testing.Short() {
		zapLogger.Warn("Short test: Skipping Quotes integration tests")

		return nil
	}

	// If running on a GitHub Actions runner use the secret stored in the GitHub Actions Secrets.
	testConfigs = "test_suite"

	if _, ok := os.LookupEnv(constants.GetGithubCIKey()); ok {
		zapLogger.Info("Integration Test running on Github CI runner.")
		zapLogger.Warn("*** Please ensure that the Quotes configurations are upto date in GHA Secrets ***")

		if testConfigs, ok = os.LookupEnv(constants.GetQuotesPrefix() + "_CI_CONFIGS"); !ok {
			msg := "failed to load Quotes configs from GitHub Actions Secrets"
			zapLogger.Error(msg)

			return fmt.Errorf(msg)
		}
	}

	return nil
}

// tearDown will delete the test clusters keyspace.
func tearDown() error {
	return nil
}
