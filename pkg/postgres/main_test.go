package postgres

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
)

// postgresConfigTestData is a map Postgres configuration test data.
var postgresConfigTestData = configTestData()

// configFileKey is the name of the Postgres configuration file to use in the tests.
var configFileKey string

// testConnection is the connection pool to the Postgres test database.
type testConnection struct {
	db Postgres // Test database connection.
	// mu sync.RWMutex // Mutex to enforce sequential test execution.
}

// connection pool to Cassandra cluster.
var connection testConnection

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
		zapLogger.Warn("Short test: Skipping Postgres integration tests")

		return nil
	}

	var err error

	// If running on a GitHub Actions runner use the default credentials for Cassandra.
	configFileKey = "test_suite"
	if _, ok := os.LookupEnv(constants.GetGithubCIKey()); ok == true {
		configFileKey = "github-ci-runner"

		zapLogger.Info("Integration Test running on Github CI runner.")
	}

	// Setup mock filesystem for configs.
	fs := afero.NewMemMapFs()
	if err = fs.MkdirAll(constants.GetEtcDir(), 0644); err != nil {
		return fmt.Errorf("afero memory mapped file system setup failed: %w", err)
	}

	if err = afero.WriteFile(fs, constants.GetEtcDir()+constants.GetPostgresFileName(),
		[]byte(postgresConfigTestData[configFileKey]), 0644); err != nil {
		return fmt.Errorf("afero memory mapped file system write failed: %w", err)
	}

	// Load Postgres configurations for test suite.
	if connection.db, err = newPostgresImpl(&fs, zapLogger); err != nil {
		return err
	}

	if err = connection.db.Open(); err != nil {
		return fmt.Errorf("postgres connection opening failed: %w", err)
	}

	return nil
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	if !testing.Short() {
		if err := connection.db.Close(); err != nil {
			return fmt.Errorf("postgres connection termination failure in test suite: %w", err)
		}
	}

	return
}
