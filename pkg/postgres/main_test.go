package postgres

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"go.uber.org/zap"
)

// postgresConfigTestData is a map Postgres configuration test data.
var postgresConfigTestData = configTestData()

// configFileKey is the name of the Postgres configuration file to use in the tests.
var configFileKey string

// connection pool to Cassandra cluster.
var connection *Postgres

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

	// If running on a GitHub Actions runner use the default credentials for Postgres.
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
	if connection, err = NewPostgres(&fs, zapLogger); err != nil {
		return err
	}

	if err = connection.Open(); err != nil {
		return fmt.Errorf("postgres connection opening failed: %w", err)
	}

	return nil
}

// tearDown will delete the test clusters keyspace.
func tearDown() (err error) {
	if !testing.Short() {
		if err := connection.Close(); err != nil {
			return fmt.Errorf("postgres connection termination failure in test suite: %w", err)
		}
	}

	return
}

// insertTestUsers will reset the users table and create some test user accounts.
func insertTestUsers(t *testing.T) {
	t.Helper()

	// Reset the users table.
	query := "DELETE FROM users WHERE first_name != 'Internal';"
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)

	defer cancel()

	rows, err := connection.Query.db.Query(ctx, query)
	rows.Close()

	require.NoError(t, err, "failed to wipe users table before reinserting users.")

	// Insert new users.
	for key, testCase := range getTestUsers() {
		user := testCase

		t.Run(fmt.Sprintf("Inserting %s", key), func(t *testing.T) {
			clientID, err := connection.Query.createUser(ctx, &user)
			require.NoErrorf(t, err, "failed to insert test user account: %w", err)
			require.True(t, clientID.Valid, "failed to retrieve client id from response")
		})
	}
}

// insertTestFiatAccounts will reset the fiat accounts table and create some test accounts.
func insertTestFiatAccounts(t *testing.T) {
	t.Helper()

	// Reset the fiat accounts table.
	query := "TRUNCATE TABLE fiat_accounts CASCADE;"
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)

	defer cancel()

	rows, err := connection.Query.db.Query(ctx, query)
	rows.Close()

	require.NoError(t, err, "failed to wipe fiat accounts table before reinserting accounts.")

	// Retrieve client ids from users table.
	clientID1, err := connection.Query.getClientIdUser(ctx, "username1")
	require.NoError(t, err, "failed to retrieve username1 client id.")
	clientID2, err := connection.Query.getClientIdUser(ctx, "username2")
	require.NoError(t, err, "failed to retrieve username2 client id.")

	// Insert new fiat accounts.
	for key, testCase := range getTestFiatAccounts(clientID1, clientID2) {
		parameters := testCase

		t.Run(fmt.Sprintf("Inserting %s", key), func(t *testing.T) {
			for _, param := range parameters {
				accInfo := param
				rowCount, err := connection.Query.createFiatAccount(ctx, &accInfo)
				require.NoError(t, err, "errored whilst trying to insert fiat account.")
				require.NotEqual(t, 0, rowCount, "no rows were added.")
			}
		})
	}
}
