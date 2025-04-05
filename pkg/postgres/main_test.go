package postgres

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gofrs/uuid"
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

// connection pool to Postgres cluster.
var connection *postgresImpl

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

	// If running on a GitHub Actions runner, use the default credentials for Postgres.
	configFileKey = "test_suite"
	if _, ok := os.LookupEnv(constants.GithubCIKey()); ok == true {
		configFileKey = "github-ci-runner"

		zapLogger.Info("Integration Test running on Github CI runner.")
	}

	// Setup mock filesystem for configs.
	fs := afero.NewMemMapFs()
	if err = fs.MkdirAll(constants.EtcDir(), 0644); err != nil {
		return fmt.Errorf("afero memory mapped file system setup failed: %w", err)
	}

	if err = afero.WriteFile(fs, constants.EtcDir()+constants.PostgresFileName(),
		[]byte(postgresConfigTestData[configFileKey]), 0644); err != nil {
		return fmt.Errorf("afero memory mapped file system write failed: %w", err)
	}

	// Load Postgres configurations for test suite.
	if connection, err = newPostgresImpl(&fs, zapLogger); err != nil {
		return err
	}

	if err = connection.Open(); err != nil {
		return fmt.Errorf("opening Postgres connection failed: %w", err)
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

// insertTestUsers will reset the user's table and create some test user accounts.
func insertTestUsers(t *testing.T) []uuid.UUID {
	t.Helper()

	// Reset the user's table.
	query := "DELETE FROM users WHERE first_name != 'Internal';"
	ctx, cancel := context.WithTimeout(context.TODO(), constants.TwoSeconds())

	defer cancel()

	rows, err := connection.queries.db.Query(ctx, query)
	rows.Close()

	require.NoError(t, err, "failed to wipe users table before reinserting users.")

	clientIDs := make([]uuid.UUID, 0, 5)

	// Insert new users.
	for _, testCase := range getTestUsers() {
		user := testCase

		clientID, err := connection.Query.userCreate(ctx, &user)
		require.NoErrorf(t, err, "failed to insert test user account: %v", err)
		require.False(t, clientID.IsNil(), "failed to retrieve client id from response")

		clientIDs = append(clientIDs, clientID)
	}

	return clientIDs
}

// resetTestFiatAccounts will reset the fiat accounts table and create some test accounts.
func resetTestFiatAccounts(t *testing.T) (uuid.UUID, uuid.UUID) {
	t.Helper()

	// Reset the fiat accounts table.
	query := "TRUNCATE TABLE fiat_accounts CASCADE;"
	ctx, cancel := context.WithTimeout(context.TODO(), constants.TwoSeconds())

	defer cancel()

	rows, err := connection.queries.db.Query(ctx, query)
	rows.Close()

	require.NoError(t, err, "failed to wipe fiat accounts table before reinserting accounts.")

	// Retrieve client ids from users' table.
	clientID1, err := connection.Query.userGetClientId(ctx, "username1")
	require.NoError(t, err, "failed to retrieve username1 client id.")
	clientID2, err := connection.Query.userGetClientId(ctx, "username2")
	require.NoError(t, err, "failed to retrieve username2 client id.")

	// Insert new fiat accounts.
	for _, testCase := range getTestFiatAccounts(clientID1, clientID2) {
		parameters := testCase

		for _, param := range parameters {
			accInfo := param
			rowCount, err := connection.Query.fiatCreateAccount(ctx, &accInfo)
			require.NoError(t, err, "errored whilst trying to insert fiat account.")
			require.NotEqual(t, 0, rowCount, "no rows were added.")
		}
	}

	return clientID1, clientID2
}

// resetTestFiatJournal will reset the fiat journal with base internal and external entries.
func resetTestFiatJournal(t *testing.T, clientID1, clientID2 uuid.UUID) {
	t.Helper()

	// Reset the fiat journal table.
	query := "TRUNCATE TABLE fiat_journal CASCADE;"
	ctx, cancel := context.WithTimeout(context.TODO(), constants.TwoSeconds())

	defer cancel()

	rows, err := connection.queries.db.Query(ctx, query)
	rows.Close()

	require.NoError(t, err, "failed to wipe Fiat journal table before reinserting entries.")

	// Get general ledger entry test cases.
	testCases := getTestFiatJournal(clientID1, clientID2)

	// Insert new Fiat accounts.
	for _, testCase := range testCases {
		parameters := testCase

		result, err := connection.Query.fiatExternalTransferJournalEntry(ctx, &parameters)
		require.NoError(t, err, "failed to insert external Fiat account entry.")
		require.False(t, result.TxID.IsNil(), "returned transaction id is invalid.")
		require.True(t, result.TransactedAt.Valid, "returned transaction time is invalid.")
	}
}

// insertTestInternalFiatGeneralLedger will not reset the journal and will insert some test internal transfers.
func insertTestInternalFiatGeneralLedger(t *testing.T, clientID1, clientID2 uuid.UUID) (
	map[string]fiatInternalTransferJournalEntryParams, map[string]fiatInternalTransferJournalEntryRow) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), constants.TwoSeconds())

	defer cancel()

	// Get journal entry test cases.
	testCases := getTestJournalInternalFiatAccounts(clientID1, clientID2)

	// Mapping for transactions to parameters.
	transactions := make(map[string]fiatInternalTransferJournalEntryRow, len(testCases))

	// Insert new fiat accounts.
	for key, testCase := range testCases {
		parameters := testCase

		row, err := connection.Query.fiatInternalTransferJournalEntry(ctx, &parameters)
		require.NoError(t, err, "errored whilst inserting internal fiat general ledger entry.")
		require.NotEqual(t, 0, row, "no rows were added.")
		transactions[key] = row
	}

	return testCases, transactions
}

// resetTestCryptoAccounts will reset the crypto accounts table and create some test accounts.
func resetTestCryptoAccounts(t *testing.T, clientID1, clientID2 uuid.UUID) {
	t.Helper()

	// Reset the crypto accounts table.
	query := "TRUNCATE TABLE crypto_accounts CASCADE;"
	ctx, cancel := context.WithTimeout(context.TODO(), constants.TwoSeconds())

	defer cancel()

	rows, err := connection.queries.db.Query(ctx, query)
	rows.Close()

	require.NoError(t, err, "failed to wipe crypto accounts table before reinserting accounts.")

	// Insert new crypto accounts.
	for _, testCase := range getTestCryptoAccounts(clientID1, clientID2) {
		parameters := testCase

		for _, param := range parameters {
			accInfo := param
			rowCount, err := connection.Query.cryptoCreateAccount(ctx, &accInfo)
			require.NoError(t, err, "errored whilst trying to insert crypto account.")
			require.NotEqual(t, 0, rowCount, "no rows were added.")
		}
	}
}

// resetTestCryptoAccounts will reset the crypto accounts table and create some test accounts.
func resetTestCryptoJournal(t *testing.T) {
	t.Helper()

	// Reset the fiat accounts table.
	query := "TRUNCATE TABLE crypto_journal CASCADE;"
	ctx, cancel := context.WithTimeout(context.TODO(), constants.TwoSeconds())

	defer cancel()

	rows, err := connection.queries.db.Query(ctx, query)
	rows.Close()

	require.NoError(t, err, "failed to wipe crypto journal table.")
}
