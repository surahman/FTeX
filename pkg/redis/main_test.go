package redis

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// redisConfigTestData is a map of Redis configuration test data.
var redisConfigTestData = configTestData()

// invalidServerAddr ia a bad/invalid server address used in test suite.
var invalidServerAddr = "127.0.0.1:7777"

// configFileKey is the redis configuration key name that is loaded based on whether the test is run locally or on GHA.
var configFileKey string

// connection pool to Redis cluster.
var connection Redis

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

// setup will configure the connection to the test clusters keyspace.
func setup() error {
	if testing.Short() {
		zapLogger.Warn("Short test: Skipping Redis integration tests")

		return nil
	}

	// If running on a GitHub Actions runner use the default credentials for Postgres.
	configFileKey = "test_suite"
	if _, ok := os.LookupEnv(constants.GithubCIKey()); ok == true {
		configFileKey = "github-ci-runner"

		zapLogger.Info("Integration Test running on Github CI runner.")
	}

	conf := config{}
	if err := yaml.Unmarshal([]byte(redisConfigTestData[configFileKey]), &conf); err != nil {
		return fmt.Errorf("failed to parse test suite Redis configs %w", err)
	}

	connection = &redisImpl{conf: &conf, logger: zapLogger}
	if err := connection.Open(); err != nil {
		return fmt.Errorf("failed to test suite Redis logger %w", err)
	}

	return nil
}

// tearDown will delete the test clusters keyspace.
func tearDown() error {
	if !testing.Short() {
		if err := connection.Close(); err != nil {
			return fmt.Errorf("redis test suite failed to close connection %w", err)
		}

		return nil
	}

	return nil
}
