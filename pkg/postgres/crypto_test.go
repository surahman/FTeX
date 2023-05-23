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

	// Configure test grid.
	txIDValid1, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate first tx id for USD to BTC.")

	txIDValid2, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate second tx id for USD to BTC.")

	txIDPKR, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for USD to BTC.")

	txIDBAD, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for bad crypto ticker.")

	txIDNoFunds, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for insufficient funds.")

	testCases := []struct {
		name      string
		params    *cryptoPurchaseParams
		expectErr require.ErrorAssertionFunc
	}{
		{
			name: "valid - USD to BTC (first)",
			params: &cryptoPurchaseParams{
				TransactionID:      txIDValid1,
				ClientID:           clientID1,
				FiatCurrency:       CurrencyUSD,
				CryptoTicker:       "BTC",
				FiatDebitAmount:    decimal.NewFromFloat(456.78),
				CryptoCreditAmount: decimal.NewFromFloat(13.12345678),
			},
			expectErr: require.NoError,
		}, {
			name: "valid - USD to BTC (second)",
			params: &cryptoPurchaseParams{
				TransactionID:      txIDValid2,
				ClientID:           clientID1,
				FiatCurrency:       CurrencyUSD,
				CryptoTicker:       "BTC",
				FiatDebitAmount:    decimal.NewFromFloat(2389.33),
				CryptoCreditAmount: decimal.NewFromFloat(104.80808081),
			},
			expectErr: require.NoError,
		}, {
			name: "invalid - PKR to BTC",
			params: &cryptoPurchaseParams{
				TransactionID:      txIDPKR,
				ClientID:           clientID1,
				FiatCurrency:       CurrencyPKR,
				CryptoTicker:       "BTC",
				FiatDebitAmount:    decimal.NewFromFloat(456.78),
				CryptoCreditAmount: decimal.NewFromFloat(13.12345678),
			},
			expectErr: require.Error,
		}, {
			name: "invalid - USD to invalid crypto",
			params: &cryptoPurchaseParams{
				TransactionID:      txIDBAD,
				ClientID:           clientID1,
				FiatCurrency:       CurrencyUSD,
				CryptoTicker:       "BAD",
				FiatDebitAmount:    decimal.NewFromFloat(77.99),
				CryptoCreditAmount: decimal.NewFromFloat(4.0000003),
			},
			expectErr: require.Error,
		}, {
			name: "invalid - USD insufficient funds",
			params: &cryptoPurchaseParams{
				TransactionID:      txIDNoFunds,
				ClientID:           clientID1,
				FiatCurrency:       CurrencyUSD,
				CryptoTicker:       "BTC",
				FiatDebitAmount:    decimal.NewFromFloat(9999999.99),
				CryptoCreditAmount: decimal.NewFromFloat(6.1100005),
			},
			expectErr: require.Error,
		},
	}

	// Configure context.
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	t.Cleanup(func() {
		cancel()
	})

	// Insert a test amount to check the final balances against.
	ts1 := pgtype.Timestamptz{}
	require.NoError(t, ts1.Scan(time.Now().UTC()), "time stamp 1 parse failed.")

	_, err = connection.Query.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: clientID1,
		Currency: CurrencyUSD,
		Amount:   decimal.NewFromFloat(5643.17),
		LastTxTs: ts1,
	})
	require.NoError(t, err, "error expectation condition failed.")

	// Configure wait groups for parallel run of all threads.
	wg := sync.WaitGroup{}
	wg.Add(len(testCases))

	// Run test grid.
	for _, testCase := range testCases {
		test := testCase

		go func() {
			defer wg.Done()

			t.Run(test.name, func(t *testing.T) {
				err := connection.Query.cryptoPurchase(ctx, test.params)
				test.expectErr(t, err, "error expectation failed.")
			})
		}()
	}

	// Wait (tie-threads).
	wg.Wait()

	// Verify results.
	//nolint:contextcheck
	t.Run("check end results", func(t *testing.T) {
		negOne := decimal.NewFromFloat(-1)

		fiatOpsAcc, err := connection.Query.userGetClientId(ctx, "fiat-currencies")
		require.NoError(t, err, "failed to retrieve Fiat operations user id.")

		cryptoOpsAcc, err := connection.Query.userGetClientId(ctx, "crypto-currencies")
		require.NoError(t, err, "failed to retrieve Fiat operations user id.")

		// Check balances.
		fiatAccount, err := connection.FiatBalanceCurrency(clientID1, CurrencyUSD)
		require.NoError(t, err, "failed to retrieve Fiat account balance.")
		require.Equal(t, fiatAccount.Balance, decimal.NewFromFloat(2797.06), "Fiat balance mismatch.")

		cryptoAccount, err := connection.CryptoBalanceCurrency(clientID1, "BTC")
		require.NoError(t, err, "failed to retrieve Crypto account balance.")
		require.Equal(t, cryptoAccount.Balance, decimal.NewFromFloat(117.93153759), "Crypto balance mismatch.")

		// Check Fiat Journal entries.
		fiatJournal, err := connection.FiatTxDetailsCurrency(clientID1, txIDValid1)
		require.NoError(t, err, "failed to retrieve Fiat journal for first valid purchase.")
		require.Equal(t, 1, len(fiatJournal), "invalid Fiat journal count for first purchase.")
		require.Equal(t, testCases[0].params.FiatDebitAmount.Mul(negOne), fiatJournal[0].Amount,
			"first Fiat debit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetailsCurrency(fiatOpsAcc, txIDValid1)
		require.NoError(t, err, "failed to retrieve ops Fiat journal for first valid purchase.")
		require.Equal(t, 1, len(fiatJournal), "invalid ops Fiat journal count for first purchase.")
		require.Equal(t, testCases[0].params.FiatDebitAmount, fiatJournal[0].Amount,
			"first ops Fiat debit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetailsCurrency(clientID1, txIDValid2)
		require.NoError(t, err, "failed to retrieve Fiat journal for second valid purchase.")
		require.Equal(t, 1, len(fiatJournal), "invalid Fiat journal count for second purchase.")
		require.Equal(t, testCases[1].params.FiatDebitAmount.Mul(negOne), fiatJournal[0].Amount,
			"second Fiat debit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetailsCurrency(fiatOpsAcc, txIDValid2)
		require.NoError(t, err, "failed to retrieve ops Fiat journal for second valid purchase.")
		require.Equal(t, 1, len(fiatJournal), "invalid ops Fiat journal count for second purchase.")
		require.Equal(t, testCases[1].params.FiatDebitAmount, fiatJournal[0].Amount,
			"second ops Fiat debit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetailsCurrency(clientID1, txIDPKR)
		require.NoError(t, err, "failed to retrieve Fiat journal for PKR purchase.")
		require.Equal(t, 0, len(fiatJournal), "Fiat journal entry for PKR purchase found.")

		fiatJournal, err = connection.FiatTxDetailsCurrency(clientID1, txIDBAD)
		require.NoError(t, err, "failed to retrieve Fiat journal for invalid crypto ticker purchase.")
		require.Equal(t, 0, len(fiatJournal), "Fiat journal entry for invalid crypto ticker purchase found.")

		fiatJournal, err = connection.FiatTxDetailsCurrency(clientID1, txIDNoFunds)
		require.NoError(t, err, "failed to retrieve Fiat journal for insufficient funds purchase.")
		require.Equal(t, 0, len(fiatJournal), "Fiat journal entry for insufficient funds purchase found.")

		// Check Crypto Journal entries.
		cryptoJournal, err := connection.CryptoTxDetailsCurrency(clientID1, txIDValid1)
		require.NoError(t, err, "failed to retrieve Crypto journal for first valid purchase.")
		require.Equal(t, 1, len(cryptoJournal), "invalid Crypto journal count for first purchase.")
		require.Equal(t, testCases[0].params.CryptoCreditAmount, cryptoJournal[0].Amount,
			"first Crypto credit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetailsCurrency(cryptoOpsAcc, txIDValid1)
		require.NoError(t, err, "failed to retrieve ops Crypto journal for first valid purchase.")
		require.Equal(t, 1, len(cryptoJournal), "invalid ops Crypto journal count for first purchase.")
		require.Equal(t, testCases[0].params.CryptoCreditAmount.Mul(negOne), cryptoJournal[0].Amount,
			"first ops Crypto debit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetailsCurrency(clientID1, txIDValid2)
		require.NoError(t, err, "failed to retrieve Crypto journal for second valid purchase.")
		require.Equal(t, 1, len(cryptoJournal), "invalid Crypto journal count for second purchase.")
		require.Equal(t, testCases[1].params.CryptoCreditAmount, cryptoJournal[0].Amount,
			"second Crypto credit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetailsCurrency(cryptoOpsAcc, txIDValid2)
		require.NoError(t, err, "failed to retrieve ops Crypto journal for second valid purchase.")
		require.Equal(t, 1, len(cryptoJournal), "invalid ops Crypto journal count for second purchase.")
		require.Equal(t, testCases[1].params.CryptoCreditAmount.Mul(negOne), cryptoJournal[0].Amount,
			"second Crypto debit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetailsCurrency(clientID1, txIDPKR)
		require.NoError(t, err, "failed to retrieve Crypto journal for PKR purchase.")
		require.Equal(t, 0, len(cryptoJournal), "Crypto journal entry for PKR purchase found.")

		cryptoJournal, err = connection.CryptoTxDetailsCurrency(clientID1, txIDBAD)
		require.NoError(t, err, "failed to retrieve Crypto journal for invalid crypto ticker purchase.")
		require.Equal(t, 0, len(cryptoJournal), "Crypto journal entry for invalid crypto ticker purchase found.")

		cryptoJournal, err = connection.CryptoTxDetailsCurrency(clientID1, txIDNoFunds)
		require.NoError(t, err, "failed to retrieve Crypto journal for insufficient funds purchase.")
		require.Equal(t, 0, len(cryptoJournal), "Crypto journal entry for insufficient funds purchase found.")
	})
}
