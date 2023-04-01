package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)
	resetTestFiatJournal(t, clientID1, clientID2)

	// End of test expected totals.
	expectedTotal := pgtype.Numeric{}
	require.NoError(t, expectedTotal.Scan("52145.77"))

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
				Amount:   5443.9786,
			},
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name: "-1293.4321",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "USD",
				Amount:   -1293.4321,
			},
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name: "-4.1235",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "USD",
				Amount:   -4.1235,
			},
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name: "47999.3587",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "USD",
				Amount:   47999.3587,
			},
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name: "Invalid Account",
			accountDetails: &FiatTransactionDetails{
				ClientID: clientID1,
				Currency: "GBP",
				Amount:   1234.5678,
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
				require.True(t, transferResult.LastTx.Valid, "invalid last transaction returned.")
				require.True(t, transferResult.Balance.Valid, "invalid balance returned.")

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
