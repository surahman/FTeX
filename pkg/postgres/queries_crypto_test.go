package postgres

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestQueries_CryptoCreateAccount(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert test users.
	clientIDs := insertTestUsers(t)

	// Insert initial set of test fiat accounts.

	// Insert initial set of test crypto accounts
	resetTestCryptoAccounts(t, clientIDs[0], clientIDs[2])

	err := connection.CryptoCreateAccount(clientIDs[0], "USDC")
	require.NoError(t, err, "failed to create Crypto account.")

	err = connection.FiatCreateAccount(clientIDs[0], "USDC")
	require.Error(t, err, "created duplicate Crypto account.")
}

func TestQueries_CryptoBalanceCurrency(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert test users.
	clientIDs := insertTestUsers(t)

	// Insert initial set of test Crypto accounts.
	resetTestCryptoAccounts(t, clientIDs[0], clientIDs[1])

	testCases := []struct {
		name      string
		ticker    string
		expectErr require.ErrorAssertionFunc
	}{
		{
			name:      "BTC valid",
			ticker:    "BTC",
			expectErr: require.NoError,
		}, {
			name:      "ETH valid",
			ticker:    "ETH",
			expectErr: require.NoError,
		}, {
			name:      "USDT valid",
			ticker:    "USDT",
			expectErr: require.NoError,
		}, {
			name:      "USDC invalid",
			ticker:    "USDC",
			expectErr: require.Error,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := connection.CryptoBalanceCurrency(clientIDs[0], testCase.ticker)
			testCase.expectErr(t, err, "error expectation failed.")
		})
	}
}

func TestQueries_CryptoPurchase(t *testing.T) {
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

	// Reset Crypto Journal entries.
	resetTestCryptoJournal(t)

	// Configure test grid.
	//nolint:dupl
	testCases := []struct {
		name               string
		clientID           uuid.UUID
		fiatCurrency       Currency
		cryptoTicker       string
		fiatDebitAmount    decimal.Decimal
		cryptoCreditAmount decimal.Decimal
		expectErr          require.ErrorAssertionFunc
	}{
		{
			name:               "valid - USD to BTC (first)",
			clientID:           clientID1,
			fiatCurrency:       CurrencyUSD,
			cryptoTicker:       "BTC",
			fiatDebitAmount:    decimal.NewFromFloat(456.78),
			cryptoCreditAmount: decimal.NewFromFloat(13.12345678),
			expectErr:          require.NoError,
		}, {
			name:               "valid - USD to BTC (second)",
			clientID:           clientID1,
			fiatCurrency:       CurrencyUSD,
			cryptoTicker:       "BTC",
			fiatDebitAmount:    decimal.NewFromFloat(2389.33),
			cryptoCreditAmount: decimal.NewFromFloat(104.80808081),
			expectErr:          require.NoError,
		}, {
			name:               "invalid - PKR to BTC",
			clientID:           clientID1,
			fiatCurrency:       CurrencyPKR,
			cryptoTicker:       "BTC",
			fiatDebitAmount:    decimal.NewFromFloat(456.78),
			cryptoCreditAmount: decimal.NewFromFloat(13.12345678),
			expectErr:          require.Error,
		}, {
			name:               "invalid - USD to invalid crypto",
			clientID:           clientID1,
			fiatCurrency:       CurrencyUSD,
			cryptoTicker:       "BAD",
			fiatDebitAmount:    decimal.NewFromFloat(77.99),
			cryptoCreditAmount: decimal.NewFromFloat(4.0000003),
			expectErr:          require.Error,
		}, {
			name:               "invalid - USD insufficient funds",
			clientID:           clientID1,
			fiatCurrency:       CurrencyUSD,
			cryptoTicker:       "BTC",
			fiatDebitAmount:    decimal.NewFromFloat(9999999.99),
			cryptoCreditAmount: decimal.NewFromFloat(6.1100005),
			expectErr:          require.Error,
		},
	}

	// Configure context.
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	t.Cleanup(func() {
		cancel()
	})

	// Insert a test amount to check the final balances against.
	ts1 := pgtype.Timestamptz{}
	require.NoError(t, ts1.Scan(time.Now().UTC()), "time stamp parse failed.")

	_, err := connection.Query.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: clientID1,
		Currency: CurrencyUSD,
		Amount:   decimal.NewFromFloat(5643.17),
		LastTxTs: ts1,
	})
	require.NoError(t, err, "error expectation condition failed.")

	negOne := decimal.NewFromFloat(-1)

	// Configure wait groups for parallel run of all threads.
	wg := sync.WaitGroup{}
	wg.Add(len(testCases))

	// Run test grid.
	for _, testCase := range testCases {
		test := testCase

		go func() {
			defer wg.Done()

			t.Run(test.name, func(t *testing.T) {
				fiatJournal, cryptoJournal, err := connection.CryptoPurchase(
					test.clientID, test.fiatCurrency, test.fiatDebitAmount, test.cryptoTicker, test.cryptoCreditAmount)
				test.expectErr(t, err, "error expectation failed.")

				if err != nil {
					return
				}

				require.Equal(t, test.fiatDebitAmount.Mul(negOne), fiatJournal.Amount, "amount in Fiat Journal mismatched.")
				require.Equal(t, test.cryptoCreditAmount, cryptoJournal.Amount, "amount in Crypto Journal mismatched.")
			})
		}()
	}

	// Wait (tie-threads).
	wg.Wait()
}

func TestQueries_CryptoSell(t *testing.T) {
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

	// Reset Crypto Journal entries.
	resetTestCryptoJournal(t)

	// Configure test grid.
	//nolint:dupl
	testCases := []struct {
		name              string
		clientID          uuid.UUID
		fiatCurrency      Currency
		cryptoTicker      string
		fiatCreditAmount  decimal.Decimal
		cryptoDebitAmount decimal.Decimal
		expectErr         require.ErrorAssertionFunc
	}{
		{
			name:              "valid - BTC to USD (first)",
			clientID:          clientID1,
			fiatCurrency:      CurrencyUSD,
			cryptoTicker:      "BTC",
			fiatCreditAmount:  decimal.NewFromFloat(992.91),
			cryptoDebitAmount: decimal.NewFromFloat(9.11992012),
			expectErr:         require.NoError,
		}, {
			name:              "valid - BTC to USD (second)",
			clientID:          clientID1,
			fiatCurrency:      CurrencyUSD,
			cryptoTicker:      "BTC",
			fiatCreditAmount:  decimal.NewFromFloat(7765.32),
			cryptoDebitAmount: decimal.NewFromFloat(11.40404049),
			expectErr:         require.NoError,
		}, {
			name:              "invalid - BTC to PKR",
			clientID:          clientID1,
			fiatCurrency:      CurrencyPKR,
			cryptoTicker:      "BTC",
			fiatCreditAmount:  decimal.NewFromFloat(555.11),
			cryptoDebitAmount: decimal.NewFromFloat(88888.12345678),
			expectErr:         require.Error,
		}, {
			name:              "invalid - invalid crypto to USD",
			clientID:          clientID1,
			fiatCurrency:      CurrencyUSD,
			cryptoTicker:      "BAD",
			fiatCreditAmount:  decimal.NewFromFloat(77.99),
			cryptoDebitAmount: decimal.NewFromFloat(4.0000003),
			expectErr:         require.Error,
		}, {
			name:              "invalid - Crypto insufficient funds",
			clientID:          clientID1,
			fiatCurrency:      CurrencyUSD,
			cryptoTicker:      "BTC",
			fiatCreditAmount:  decimal.NewFromFloat(9999999.99),
			cryptoDebitAmount: decimal.NewFromFloat(9191919191.1100005),
			expectErr:         require.Error,
		},
	}

	// Configure context.
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	t.Cleanup(func() {
		cancel()
	})

	// Insert a test amount to check the final balances against.
	ts1 := pgtype.Timestamptz{}
	require.NoError(t, ts1.Scan(time.Now().UTC()), "time stamp parse failed.")

	_, err := connection.Query.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: clientID1,
		Currency: CurrencyUSD,
		Amount:   decimal.NewFromFloat(64320.27),
		LastTxTs: ts1,
	})
	require.NoError(t, err, "error expectation condition failed.")

	_, _, err = connection.CryptoPurchase(
		clientID1, CurrencyUSD, decimal.NewFromFloat(22.22), "BTC", decimal.NewFromFloat(4444.4444))
	require.NoError(t, err, "error expectation condition failed.")

	negOne := decimal.NewFromFloat(-1)

	// Configure wait groups for parallel run of all threads.
	wg := sync.WaitGroup{}
	wg.Add(len(testCases))

	// Run test grid.
	for _, testCase := range testCases {
		test := testCase

		go func() {
			defer wg.Done()

			t.Run(test.name, func(t *testing.T) {
				fiatJournal, cryptoJournal, err := connection.CryptoSell(
					test.clientID, test.fiatCurrency, test.fiatCreditAmount, test.cryptoTicker, test.cryptoDebitAmount)
				test.expectErr(t, err, "error expectation failed.")

				if err != nil {
					return
				}

				require.Equal(t, test.fiatCreditAmount, fiatJournal.Amount, "amount in Fiat Journal mismatched.")
				require.Equal(t, test.cryptoDebitAmount.Mul(negOne), cryptoJournal.Amount, "amount in Crypto Journal mismatched.")
			})
		}()
	}

	// Wait (tie-threads).
	wg.Wait()
}
