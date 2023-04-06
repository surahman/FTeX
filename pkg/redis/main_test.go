package redis

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// redisConfigTestData is a map of Redis configuration test data.
var redisConfigTestData = configTestData()

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
func setup() (err error) {
	if testing.Short() {
		zapLogger.Warn("Short test: Skipping Redis integration tests")

		return
	}

	conf := config{}
	if err = yaml.Unmarshal([]byte(redisConfigTestData["test_suite"]), &conf); err != nil {
		return
	}

	connection = &redisImpl{conf: &conf, logger: zapLogger}
	if err = connection.Open(); err != nil {
		return
	}

	return
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	return
}
