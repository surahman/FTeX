package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFiat_CreateFiatAccount(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Account collisions test.
	for key, testCase := range getTestFiatAccounts(clientID1, clientID2) {
		parameters := testCase

		t.Run(fmt.Sprintf("Inserting %s", key), func(t *testing.T) {
			for _, param := range parameters {
				accInfo := param
				rowCount, err := connection.Query.createFiatAccount(ctx, &accInfo)
				require.Error(t, err, "did not error whilst inserting duplicate fiat account.")
				require.Equal(t, int64(0), rowCount, "rows were added.")
			}
		})
	}
}

func TestFiat_RowLockFiatAccount(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Get general ledger entry test cases.
	testCases := []struct {
		name         string
		parameter    rowLockFiatAccountParams
		errExpected  require.ErrorAssertionFunc
		boolExpected require.BoolAssertionFunc
	}{
		{
			name: "Client1 - USD",
			parameter: rowLockFiatAccountParams{
				ClientID: clientID1,
				Currency: CurrencyUSD,
			},
			errExpected:  require.NoError,
			boolExpected: require.True,
		}, {
			name: "Client1 - AED",
			parameter: rowLockFiatAccountParams{
				ClientID: clientID1,
				Currency: CurrencyAED,
			},
			errExpected:  require.NoError,
			boolExpected: require.True,
		}, {
			name: "Client1 - CAD",
			parameter: rowLockFiatAccountParams{
				ClientID: clientID1,
				Currency: CurrencyCAD,
			},
			errExpected:  require.NoError,
			boolExpected: require.True,
		}, {
			name: "Client2 - USD",
			parameter: rowLockFiatAccountParams{
				ClientID: clientID2,
				Currency: CurrencyUSD,
			},
			errExpected:  require.NoError,
			boolExpected: require.True,
		}, {
			name: "Client2 - AED",
			parameter: rowLockFiatAccountParams{
				ClientID: clientID2,
				Currency: CurrencyAED,
			},
			errExpected:  require.NoError,
			boolExpected: require.True,
		}, {
			name: "Client2 - CAD",
			parameter: rowLockFiatAccountParams{
				ClientID: clientID2,
				Currency: CurrencyCAD,
			},
			errExpected:  require.NoError,
			boolExpected: require.True,
		}, {
			name: "Client1 - Not Found",
			parameter: rowLockFiatAccountParams{
				ClientID: clientID1,
				Currency: CurrencyEUR,
			},
			errExpected:  require.Error,
			boolExpected: require.False,
		},
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Insert new fiat accounts.
	for _, testCase := range testCases {
		test := testCase

		t.Run(fmt.Sprintf("Inserting %s", test.name), func(t *testing.T) {
			balance, err := connection.Query.rowLockFiatAccount(ctx, &test.parameter)
			test.errExpected(t, err, "error expectation condition failed.")
			test.boolExpected(t, balance.Valid, "invalid balance.")
		})
	}
}

func TestFiat_GeneralLedgerExternalFiatAccount(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the test
	resetTestFiatGeneralLedger(t, clientID1, clientID2)
}

func TestFiat_GeneralLedgerAccountTxFiatAccount(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the test
	resetTestFiatGeneralLedger(t, clientID1, clientID2)

	// Get general ledger entry test cases.
	testCases := []struct {
		name        string
		parameter   generalLedgerAccountTxFiatAccountParams
		errExpected require.ErrorAssertionFunc
	}{
		{
			name: "Client1 - USD",
			parameter: generalLedgerAccountTxFiatAccountParams{
				ClientID: clientID1,
				Currency: CurrencyUSD,
			},
			errExpected: require.NoError,
		}, {
			name: "Client1 - AED",
			parameter: generalLedgerAccountTxFiatAccountParams{
				ClientID: clientID1,
				Currency: CurrencyAED,
			},
			errExpected: require.NoError,
		}, {
			name: "Client1 - CAD",
			parameter: generalLedgerAccountTxFiatAccountParams{
				ClientID: clientID1,
				Currency: CurrencyCAD,
			},
			errExpected: require.NoError,
		}, {
			name: "Client2 - USD",
			parameter: generalLedgerAccountTxFiatAccountParams{
				ClientID: clientID2,
				Currency: CurrencyUSD,
			},
			errExpected: require.NoError,
		}, {
			name: "Client2 - AED",
			parameter: generalLedgerAccountTxFiatAccountParams{
				ClientID: clientID2,
				Currency: CurrencyAED,
			},
			errExpected: require.NoError,
		}, {
			name: "Client2 - CAD",
			parameter: generalLedgerAccountTxFiatAccountParams{
				ClientID: clientID2,
				Currency: CurrencyCAD,
			},
			errExpected: require.NoError,
		}, {
			name: "Client1 - Not Found",
			parameter: generalLedgerAccountTxFiatAccountParams{
				ClientID: clientID1,
				Currency: CurrencyEUR,
			},
			errExpected: require.NoError,
		},
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Insert new fiat accounts.
	for _, testCase := range testCases {
		test := testCase

		t.Run(fmt.Sprintf("Inserting %s", test.name), func(t *testing.T) {
			results, err := connection.Query.generalLedgerAccountTxFiatAccount(ctx, &test.parameter)
			test.errExpected(t, err, "error expectation condition failed.")
			for _, result := range results {
				require.Equal(t, test.parameter.Currency, result.Currency, "currency type mismatch.")
				require.True(t, result.ClientID.Valid, "invalid UUID.")
				require.True(t, result.TxID.Valid, "invalid TX ID.")
				require.True(t, result.Ammount.Valid, "invalid amount.")
				require.True(t, result.TransactedAt.Valid, "invalid TX time.")
			}
		})
	}
}
