package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestTransactions_FiatTransactionsDetails_Less(t *testing.T) {
	t.Parallel()

	firstUUID, err := uuid.FromString("515fb04e-ea91-460b-ad4e-487cb673601e")
	require.NoError(t, err, "failed to parse first UUID.")

	secondUUID, err := uuid.FromString("d68d52a1-aa7c-4301-a27f-611196726edc")
	require.NoError(t, err, "failed to parse second UUID.")

	var (
		uuid1USD = &FiatTransactionDetails{ClientID: firstUUID, Currency: CurrencyUSD}
		uuid1AED = &FiatTransactionDetails{ClientID: firstUUID, Currency: CurrencyAED}
		uuid2USD = &FiatTransactionDetails{ClientID: secondUUID, Currency: CurrencyUSD}
		uuid2AED = &FiatTransactionDetails{ClientID: secondUUID, Currency: CurrencyAED}
	)

	testCases := []struct {
		name        string
		lhs         *FiatTransactionDetails
		rhs         *FiatTransactionDetails
		expectedLHS *FiatTransactionDetails
		expectedRHS *FiatTransactionDetails
	}{
		{
			name:        "UUID1 USD, UUID1 AED - Swap",
			lhs:         uuid1USD,
			rhs:         uuid1AED,
			expectedLHS: uuid1AED,
			expectedRHS: uuid1USD,
		}, {
			name:        "UUID1 USD, UUID2 USD - No swap",
			lhs:         uuid1USD,
			rhs:         uuid2USD,
			expectedLHS: uuid1USD,
			expectedRHS: uuid2USD,
		}, {
			name:        "UUID1 USD, UUID2 AED - No swap",
			lhs:         uuid1USD,
			rhs:         uuid2AED,
			expectedLHS: uuid1USD,
			expectedRHS: uuid2AED,
		}, {
			name:        "UUID1 AED, UUID2 USD - No swap",
			lhs:         uuid1AED,
			rhs:         uuid2USD,
			expectedLHS: uuid1AED,
			expectedRHS: uuid2USD,
		}, {
			name:        "UUID1 AED, UUID2 AED - No swap",
			lhs:         uuid1AED,
			rhs:         uuid2AED,
			expectedLHS: uuid1AED,
			expectedRHS: uuid2AED,
		}, {
			name:        "UUID2 AED, UUID1 USD - Swap.",
			lhs:         uuid2AED,
			rhs:         uuid1USD,
			expectedLHS: uuid1USD,
			expectedRHS: uuid2AED,
		}, {
			name:        "UUID2 AED, UUID2 USD - No swap.",
			lhs:         uuid2AED,
			rhs:         uuid2USD,
			expectedLHS: uuid2AED,
			expectedRHS: uuid2USD,
		}, {
			name:        "UUID2 AED, UUID1 AED - Swap.",
			lhs:         uuid2AED,
			rhs:         uuid1AED,
			expectedLHS: uuid1AED,
			expectedRHS: uuid2AED,
		}, {
			name:        "UUID2 USD, UUID1 USD - Swap.",
			lhs:         uuid2USD,
			rhs:         uuid1USD,
			expectedLHS: uuid1USD,
			expectedRHS: uuid2USD,
		}, {
			name:        "UUID2 USD, UUID1 AED - Swap.",
			lhs:         uuid2USD,
			rhs:         uuid1AED,
			expectedLHS: uuid1AED,
			expectedRHS: uuid2USD,
		}, {
			name:        "UUID1 AED, UUID1 USD - Swap.",
			lhs:         uuid1AED,
			rhs:         uuid1USD,
			expectedLHS: uuid1AED,
			expectedRHS: uuid1USD,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			first, second := test.lhs.Less(test.rhs)

			require.Equal(t, test.expectedLHS, *first, "first parameter did not match expected.")
			require.Equal(t, test.expectedRHS, *second, "second parameter did not match expected.")
		})
	}
}

func TestTransactions_FiatExternalTransfer(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test Fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the Fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// End of test expected totals.
	expectedTotal := decimal.NewFromFloat(52145.79)

	// Test grid.
	testCases := []struct {
		name                 string
		amount               float64
		accountDetails       *FiatTransactionDetails
		errExpectation       require.ErrorAssertionFunc
		nilResultExpectation require.ValueAssertionFunc
	}{
		{
			name: "5443.9786",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "USD",
				Amount:   decimal.NewFromFloat(5443.9786),
			},
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name: "-1293.4321",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "USD",
				Amount:   decimal.NewFromFloat(-1293.4321),
			},
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name: "-4.1235",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "USD",
				Amount:   decimal.NewFromFloat(-4.1235),
			},
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name: "47999.3587",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "USD",
				Amount:   decimal.NewFromFloat(47999.3587),
			},
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name: "Invalid Account",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "GBP",
				Amount:   decimal.NewFromFloat(1234.5678),
			},
			errExpectation:       require.Error,
			nilResultExpectation: require.Nil,
		},
	}

	// Sync tests. This test cannot be run in parallel with other tests and the total must be checked once completed.
	wg := sync.WaitGroup{}
	wg.Add(len(testCases))

	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	defer cancel()

	for _, testCase := range testCases {
		test := testCase

		go func() {
			t.Run(fmt.Sprintf("Transferring %s", test.name), func(t *testing.T) {
				defer wg.Done()

				// External transfer
				transferResult, err := connection.FiatExternalTransfer(ctx, test.accountDetails)
				test.errExpectation(t, err, "error expectation failed.")
				test.nilResultExpectation(t, transferResult, "nil transferResult expectation failed.")

				if transferResult == nil {
					return
				}

				require.False(t, transferResult.TxID.IsNil(), "invalid transaction ID returned.")
				require.False(t, transferResult.ClientID.IsNil(), "invalid client ID returned.")
				require.True(t, transferResult.Currency.Valid(), "invalid currency returned.")
				require.True(t, transferResult.TxTS.Valid, "invalid transaction timestamp returned.")

				// Check for journal entries.
				journalRow, err := connection.Query.FiatGetJournalTransaction(ctx, transferResult.TxID)
				require.NoError(t, err, "failed to retrieve journal entries for transaction.")
				require.Equal(t, 2, len(journalRow), "incorrect number of journal entries.")
			})
		}()
	}

	// Check end of test totals.
	wg.Wait()

	t.Run("Checking end totals", func(t *testing.T) {
		actual, err := connection.Query.FiatGetAccount(ctx, &FiatGetAccountParams{
			ClientID: clientID1,
			Currency: CurrencyUSD,
		})
		require.NoError(t, err, "failed to fiat account.")
		require.Equal(t, expectedTotal, actual.Balance, "end of test expected totals mismatched.")
	})
}

func TestTransactions_FiatTransactionRowLockAndBalanceCheck(t *testing.T) {
	t.Parallel()

	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test Fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the Fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// Initial balances, transaction amounts, and timestamp.
	var (
		balanceClientID1 = decimal.NewFromFloat(52145.77)
		balanceClientID2 = decimal.NewFromFloat(1921.68)
		amount20k        = decimal.NewFromFloat(20987.65)
		amount1k         = decimal.NewFromFloat(1234.56)
		txTimestamp      = pgtype.Timestamptz{}
	)

	require.NoError(t, txTimestamp.Scan(time.Now().UTC()), "failed to create current timestamp.")

	// Configure context for test suite.
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	t.Cleanup(func() {
		cancel()
	})

	// Update base balances in accounts to test from.
	_, err := connection.Query.FiatUpdateAccountBalance(ctx, &FiatUpdateAccountBalanceParams{
		ClientID: clientID1,
		Currency: CurrencyUSD,
		Amount:   balanceClientID1,
		LastTxTs: txTimestamp,
	})
	require.NoError(t, err, "failed to set base balance for Client1 in USD")

	_, err = connection.Query.FiatUpdateAccountBalance(ctx, &FiatUpdateAccountBalanceParams{
		ClientID: clientID2,
		Currency: CurrencyAED,
		Amount:   balanceClientID2,
		LastTxTs: txTimestamp,
	})
	require.NoError(t, err, "failed to set base balance for Client2 in AED")

	// Test grid.
	testCases := []struct {
		name           string
		source         FiatTransactionDetails
		destination    FiatTransactionDetails
		errExpectation require.ErrorAssertionFunc
	}{
		{
			name: "Client1USD 20k, Client2AED 1k",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   amount20k,
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyAED,
				Amount:   amount1k,
			},
			errExpectation: require.NoError,
		}, {
			name: "Client2AED 1k, Client1USD 20k",
			source: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyAED,
				Amount:   amount1k,
			},
			destination: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   amount20k,
			},
			errExpectation: require.NoError,
		}, {
			name: "Client1USD 1k, Client2AED 20k",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(52145.80),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyAED,
				Amount:   amount20k,
			},
			errExpectation: require.Error,
		}, {
			name: "Client2AED 20k, Client1USD 1k",
			source: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyAED,
				Amount:   amount20k,
			},
			destination: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   amount1k,
			},
			errExpectation: require.Error,
		},
	}

	// Run testing grid in parallel.
	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Configure transaction and setup rollback.
			tx, err := connection.pool.Begin(ctx)
			require.NoError(t, err, "failed to start transaction.")

			defer tx.Rollback(ctx) //nolint:errcheck

			queryTx := connection.queries.WithTx(tx)

			err = fiatTransactionRowLockAndBalanceCheck(ctx, queryTx, &test.source, &test.destination)
			test.errExpectation(t, err, "failed error expectation")

			if err != nil {
				require.Contains(t, err.Error(), "insufficient", "failure not related to balance.")
			}
		})
	}
}

func TestTransactions_FiatInternalTransfer(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test Fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the Fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	var (
		expectedTotalClientID1 = decimal.NewFromFloat(616115.59)
		expectedTotalClientID2 = decimal.NewFromFloat(284321.37)
		balanceClientID1       = decimal.NewFromFloat(521459.77)
		balanceClientID2       = decimal.NewFromFloat(192103.68)
		txTimestamp            = pgtype.Timestamptz{}
	)

	require.NoError(t, txTimestamp.Scan(time.Now().UTC()), "failed to create current timestamp.")

	// Configure context for test suite.
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	defer cancel()

	// Update base balances in accounts to test from.
	_, err := connection.Query.FiatUpdateAccountBalance(ctx, &FiatUpdateAccountBalanceParams{
		ClientID: clientID1,
		Currency: CurrencyUSD,
		Amount:   balanceClientID1,
		LastTxTs: txTimestamp,
	})
	require.NoError(t, err, "failed to set base balance for Client1 in USD")

	_, err = connection.Query.FiatUpdateAccountBalance(ctx, &FiatUpdateAccountBalanceParams{
		ClientID: clientID2,
		Currency: CurrencyCAD,
		Amount:   balanceClientID2,
		LastTxTs: txTimestamp,
	})
	require.NoError(t, err, "failed to set base balance for Client2 in AED")

	// Test grid.
	testCases := []struct {
		name           string
		source         FiatTransactionDetails
		destination    FiatTransactionDetails
		errExpectation require.ErrorAssertionFunc
	}{
		{
			name: "Client1USD 6830.69, Client2CAD 10182.72",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(6830.69),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(10182.72),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client2CAD 9300.58, Client1USD 11894.37",
			source: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(9300.58),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(11894.37),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client1USD 5741.18, Client2CAD 7678.79",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(5741.18),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(7678.79),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client2CAD 5034.36, Client1USD 2469.99",
			source: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(5034.36),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(2469.99),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client1USD 14657.84, Client2CAD 14763.92,",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(14657.84),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(14763.92),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client2CAD 12517.73, Client1USD 12828.39",
			source: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(12517.73),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(12828.39),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client1USD 7887.40, Client2CAD 10453.91",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(7887.40),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(10453.91),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client2CAD 7838.29, Client1USD 6783.08",
			source: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(7838.29),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(6783.08),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client1USD 14287.55, Client2CAD 2407.57",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(14287.55),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(2407.57),
			},
			errExpectation: require.NoError,
		}, {
			name: "Client2CAD 12039.82, Client1USD 11275.33",
			source: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(12039.82),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(11275.33),
			},
			errExpectation: require.NoError,
		}, {
			name: "Insufficient funds",
			source: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyCAD,
				Amount:   decimal.NewFromFloat(999999.82),
			},
			destination: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(11275.33),
			},
			errExpectation: require.Error,
		},
	}

	// Configure wait groups for parallel run of all threads.
	wg := sync.WaitGroup{}
	wg.Add(len(testCases))

	// Run testing grid in parallel.
	for _, testCase := range testCases {
		test := testCase

		go func() {
			t.Run(test.name, func(t *testing.T) {
				defer wg.Done()

				srcResult, dstResult, err := connection.FiatInternalTransfer(ctx, &test.source, &test.destination)
				test.errExpectation(t, err, "failed error expectation")

				if err != nil {
					return
				}

				require.False(t, srcResult.ClientID.IsNil(), "source client id is invalid.")
				require.Equal(t, test.source.ClientID, srcResult.ClientID, "source client id mismatch.")
				require.Equal(t, test.source.Currency, srcResult.Currency, "source currency mismatch.")
				require.False(t, srcResult.TxID.IsNil(), "source transaction id is invalid.")
				require.True(t, srcResult.TxTS.Valid, "source transaction timestamp is invalid.")

				require.False(t, dstResult.ClientID.IsNil(), "destination client id is invalid.")
				require.Equal(t, test.destination.ClientID, dstResult.ClientID, "destination client id mismatch.")
				require.Equal(t, test.destination.Currency, dstResult.Currency, "destination currency mismatch.")
				require.False(t, dstResult.TxID.IsNil(), "destination transaction id is invalid.")
				require.True(t, dstResult.TxTS.Valid, "destination transaction timestamp is invalid.")

				// Check for journal entries.
				journalRow, err := connection.Query.FiatGetJournalTransaction(ctx, dstResult.TxID)
				require.NoError(t, err, "failed to retrieve journal entries for transaction.")
				require.Equal(t, 2, len(journalRow), "incorrect number of journal entries.")
			})
		}()
	}

	// Wait (tie-threads).
	wg.Wait()

	t.Run("Checking end totals", func(t *testing.T) {
		client1, err := connection.Query.FiatGetAccount(ctx, &FiatGetAccountParams{
			ClientID: clientID1,
			Currency: CurrencyUSD,
		})
		require.NoError(t, err, "failed to fiat account.")
		require.Equal(t, expectedTotalClientID1, client1.Balance, "client 1's balance mismatched.")

		client2, err := connection.Query.FiatGetAccount(ctx, &FiatGetAccountParams{
			ClientID: clientID2,
			Currency: CurrencyCAD,
		})
		require.NoError(t, err, "failed to fiat account.")
		require.Equal(t, expectedTotalClientID2, client2.Balance, "client 2's balance mismatched.")
	})
}