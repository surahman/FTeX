package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

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

	xferDetails := FiatAccountDetails{
		ClientID: clientID1,
		Currency: "USD",
	}

	// Test grid.
	testCases := []struct {
		name                 string
		amount               float64
		accountDetails       *FiatAccountDetails
		errExpectation       require.ErrorAssertionFunc
		nilResultExpectation require.ValueAssertionFunc
	}{
		{
			name:                 "5443.9786",
			amount:               5443.9786,
			accountDetails:       &xferDetails,
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name:                 "-1293.4321",
			amount:               -1293.4321,
			accountDetails:       &xferDetails,
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name:                 "-4.1235",
			amount:               -4.1235,
			accountDetails:       &xferDetails,
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name:                 "47999.3587",
			amount:               47999.3587,
			accountDetails:       &xferDetails,
			errExpectation:       require.NoError,
			nilResultExpectation: require.NotNil,
		}, {
			name:   "Invalid Account",
			amount: 1234.5678,
			accountDetails: &FiatAccountDetails{
				ClientID: clientID1,
				Currency: "GBP",
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
				transferResult, err := connection.FiatExternalTransfer(ctx, test.accountDetails, test.amount)
				test.errExpectation(t, err, "error expectation failed.")
				test.nilResultExpectation(t, transferResult, "nil transferResult expectation failed.")

				if transferResult == nil {
					return
				}

				require.True(t, transferResult.TxID.Valid, "invalid transaction ID returned.")
				require.True(t, transferResult.ClientID.Valid, "invalid client ID returned.")
				require.True(t, transferResult.Currency.Valid(), "invalid currency returned.")
				require.True(t, transferResult.TxTS.Valid, "invalid transaction timestamp returned.")
				require.True(t, transferResult.LastTx.Valid, "invalid last transaction returned.")
				require.True(t, transferResult.Balance.Valid, "invalid balance returned.")

				// Check for journal entries.
				journalRow, err := connection.Query.fiatGetJournalTransaction(ctx, transferResult.TxID)
				require.NoError(t, err, "failed to retrieve journal entries for transaction.")
				require.Equal(t, 2, len(journalRow), "incorrect number of journal entries.")
			})
		}()
	}

	// Check end of test totals.
	wg.Wait()

	t.Run("Checking end totals", func(t *testing.T) {
		actual, err := connection.Query.fiatGetAccount(ctx, &fiatGetAccountParams{
			ClientID: xferDetails.ClientID,
			Currency: xferDetails.Currency,
		})
		require.NoError(t, err, "failed to fiat account.")
		require.Equal(t, expectedTotal, actual.Balance, "end of test expected totals mismatched.")
	})
}
