package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestCrypto_CryptoCreateAccount(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Insert initial set of test crypto accounts.
	resetTestCryptoAccounts(t, clientID1, clientID2)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Account collisions test.
	for key, testCase := range getTestCryptoAccounts(clientID1, clientID2) {
		parameters := testCase

		t.Run(fmt.Sprintf("Inserting %s", key), func(t *testing.T) {
			for _, param := range parameters {
				accInfo := param
				rowCount, err := connection.Query.cryptoCreateAccount(ctx, &accInfo)
				require.Error(t, err, "did not error whilst inserting duplicate crypto account.")
				require.Equal(t, int64(0), rowCount, "rows were added.")
			}
		})
	}
}

func TestCrypto_PurchaseCryptocurrency(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Insert the initial set of test fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// Insert initial set of test crypto accounts.
	resetTestCryptoAccounts(t, clientID1, clientID2)

	var (
		amount1Ts = time.Now().UTC()
		amount1   = decimal.NewFromFloat(5643.17)
		ts1       = pgtype.Timestamptz{}
	)

	require.NoError(t, ts1.Scan(amount1Ts), "time stamp 1 parse failed.")

	// Configure context.
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Insert a test amount to check the final balances against.
	_, err := connection.Query.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: clientID1,
		Currency: CurrencyUSD,
		Amount:   amount1,
		LastTxTs: ts1,
	})
	require.NoError(t, err, "error expectation condition failed.")
}
