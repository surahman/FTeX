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

func TestCrypto_CryptoCreateAccount(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Insert an initial set of test crypto accounts.
	resetTestCryptoAccounts(t, clientID1, clientID2)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Account collisions test.
	for key, testCase := range getTestCryptoAccounts(clientID1, clientID2) {
		parameters := testCase

		t.Run("Inserting "+key, func(t *testing.T) {
			for _, param := range parameters {
				accInfo := param
				rowCount, err := connection.Query.cryptoCreateAccount(ctx, &accInfo)
				require.Error(t, err, "did not error whilst inserting duplicate crypto account.")
				require.Zero(t, rowCount, "rows were added.")
			}
		})
	}
}

func TestCrypto_CryptoPurchase(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Insert the initial set of test fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// Insert an initial set of test crypto accounts.
	resetTestCryptoAccounts(t, clientID1, clientID2)

	// Reset Crypto Journal entries.
	resetTestCryptoJournal(t)

	// Configure test grid.
	txIDValid1, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate first tx id for USD to BTC.")

	txIDValid2, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate second tx id for USD to BTC.")

	txIDPKR, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for PKR to BTC.")

	txIDBAD, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for bad crypto ticker.")

	txIDNoFunds, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for insufficient funds.")

	testCases := []struct { //nolint:dupl
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
		fiatAccount, err := connection.FiatBalance(clientID1, CurrencyUSD)
		require.NoError(t, err, "failed to retrieve Fiat account balance.")
		require.Equal(t, fiatAccount.Balance, decimal.NewFromFloat(2797.06), "Fiat balance mismatch.")

		cryptoAccount, err := connection.CryptoBalance(clientID1, "BTC")
		require.NoError(t, err, "failed to retrieve Crypto account balance.")
		require.Equal(t, cryptoAccount.Balance, decimal.NewFromFloat(117.93153759), "Crypto balance mismatch.")

		// Check Fiat Journal entries.
		fiatJournal, err := connection.FiatTxDetails(clientID1, txIDValid1)
		require.NoError(t, err, "failed to retrieve Fiat journal for first valid purchase.")
		require.Len(t, fiatJournal, 1, "invalid Fiat journal count for first purchase.")
		require.Equal(t, testCases[0].params.FiatDebitAmount.Mul(negOne), fiatJournal[0].Amount,
			"first Fiat debit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetails(fiatOpsAcc, txIDValid1)
		require.NoError(t, err, "failed to retrieve ops Fiat journal for first valid purchase.")
		require.Len(t, fiatJournal, 1, "invalid ops Fiat journal count for first purchase.")
		require.Equal(t, testCases[0].params.FiatDebitAmount, fiatJournal[0].Amount,
			"first ops Fiat debit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetails(clientID1, txIDValid2)
		require.NoError(t, err, "failed to retrieve Fiat journal for second valid purchase.")
		require.Len(t, fiatJournal, 1, "invalid Fiat journal count for second purchase.")
		require.Equal(t, testCases[1].params.FiatDebitAmount.Mul(negOne), fiatJournal[0].Amount,
			"second Fiat debit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetails(fiatOpsAcc, txIDValid2)
		require.NoError(t, err, "failed to retrieve ops Fiat journal for second valid purchase.")
		require.Len(t, fiatJournal, 1, "invalid ops Fiat journal count for second purchase.")
		require.Equal(t, testCases[1].params.FiatDebitAmount, fiatJournal[0].Amount,
			"second ops Fiat debit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetails(clientID1, txIDPKR)
		require.NoError(t, err, "failed to retrieve Fiat journal for PKR purchase.")
		require.Empty(t, fiatJournal, "Fiat journal entry for PKR purchase found.")

		fiatJournal, err = connection.FiatTxDetails(clientID1, txIDBAD)
		require.NoError(t, err, "failed to retrieve Fiat journal for invalid crypto ticker purchase.")
		require.Empty(t, fiatJournal, "Fiat journal entry for invalid crypto ticker purchase found.")

		fiatJournal, err = connection.FiatTxDetails(clientID1, txIDNoFunds)
		require.NoError(t, err, "failed to retrieve Fiat journal for insufficient funds purchase.")
		require.Empty(t, fiatJournal, "Fiat journal entry for insufficient funds purchase found.")

		// Check Crypto Journal entries.
		cryptoJournal, err := connection.CryptoTxDetails(clientID1, txIDValid1)
		require.NoError(t, err, "failed to retrieve Crypto journal for first valid purchase.")
		require.Len(t, cryptoJournal, 1, "invalid Crypto journal count for first purchase.")
		require.Equal(t, testCases[0].params.CryptoCreditAmount, cryptoJournal[0].Amount,
			"first Crypto credit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetails(cryptoOpsAcc, txIDValid1)
		require.NoError(t, err, "failed to retrieve ops Crypto journal for first valid purchase.")
		require.Len(t, cryptoJournal, 1, "invalid ops Crypto journal count for first purchase.")
		require.Equal(t, testCases[0].params.CryptoCreditAmount.Mul(negOne), cryptoJournal[0].Amount,
			"first ops Crypto debit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetails(clientID1, txIDValid2)
		require.NoError(t, err, "failed to retrieve Crypto journal for second valid purchase.")
		require.Len(t, cryptoJournal, 1, "invalid Crypto journal count for second purchase.")
		require.Equal(t, testCases[1].params.CryptoCreditAmount, cryptoJournal[0].Amount,
			"second Crypto credit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetails(cryptoOpsAcc, txIDValid2)
		require.NoError(t, err, "failed to retrieve ops Crypto journal for second valid purchase.")
		require.Len(t, cryptoJournal, 1, "invalid ops Crypto journal count for second purchase.")
		require.Equal(t, testCases[1].params.CryptoCreditAmount.Mul(negOne), cryptoJournal[0].Amount,
			"second Crypto debit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetails(clientID1, txIDPKR)
		require.NoError(t, err, "failed to retrieve Crypto journal for PKR purchase.")
		require.Empty(t, cryptoJournal, "Crypto journal entry for PKR purchase found.")

		cryptoJournal, err = connection.CryptoTxDetails(clientID1, txIDBAD)
		require.NoError(t, err, "failed to retrieve Crypto journal for invalid crypto ticker purchase.")
		require.Empty(t, cryptoJournal, "Crypto journal entry for invalid crypto ticker purchase found.")

		cryptoJournal, err = connection.CryptoTxDetails(clientID1, txIDNoFunds)
		require.NoError(t, err, "failed to retrieve Crypto journal for insufficient funds purchase.")
		require.Empty(t, cryptoJournal, "Crypto journal entry for insufficient funds purchase found.")
	})
}

func TestCrypto_CryptoSell(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Insert the initial set of test fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// Insert an initial set of test crypto accounts.
	resetTestCryptoAccounts(t, clientID1, clientID2)

	// Reset Crypto Journal entries.
	resetTestCryptoJournal(t)

	// Configure test grid.
	txIDValid1, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate first tx id for BTC to USD.")

	txIDValid2, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate second tx id for BTC to USD.")

	txIDPKR, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for BTC to PKR.")

	txIDBAD, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for bad crypto ticker.")

	txIDNoFunds, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate tx id for insufficient funds.")

	testCases := []struct { //nolint:dupl
		name      string
		params    *cryptoSellParams
		expectErr require.ErrorAssertionFunc
	}{
		{
			name: "valid - BTC to USD (first)",
			params: &cryptoSellParams{
				TransactionID:     txIDValid1,
				ClientID:          clientID1,
				FiatCurrency:      CurrencyUSD,
				CryptoTicker:      "BTC",
				FiatCreditAmount:  decimal.NewFromFloat(992.91),
				CryptoDebitAmount: decimal.NewFromFloat(9.11992012),
			},
			expectErr: require.NoError,
		}, {
			name: "valid - BTC to USD (second)",
			params: &cryptoSellParams{
				TransactionID:     txIDValid2,
				ClientID:          clientID1,
				FiatCurrency:      CurrencyUSD,
				CryptoTicker:      "BTC",
				FiatCreditAmount:  decimal.NewFromFloat(7765.32),
				CryptoDebitAmount: decimal.NewFromFloat(11.40404049),
			},
			expectErr: require.NoError,
		}, {
			name: "invalid - BTC to PKR",
			params: &cryptoSellParams{
				TransactionID:     txIDPKR,
				ClientID:          clientID1,
				FiatCurrency:      CurrencyPKR,
				CryptoTicker:      "BTC",
				FiatCreditAmount:  decimal.NewFromFloat(555.11),
				CryptoDebitAmount: decimal.NewFromFloat(88888.12345678),
			},
			expectErr: require.Error,
		}, {
			name: "invalid - invalid crypto to USD",
			params: &cryptoSellParams{
				TransactionID:     txIDBAD,
				ClientID:          clientID1,
				FiatCurrency:      CurrencyUSD,
				CryptoTicker:      "BAD",
				FiatCreditAmount:  decimal.NewFromFloat(77.99),
				CryptoDebitAmount: decimal.NewFromFloat(4.0000003),
			},
			expectErr: require.Error,
		}, {
			name: "invalid - Crypto insufficient funds",
			params: &cryptoSellParams{
				TransactionID:     txIDNoFunds,
				ClientID:          clientID1,
				FiatCurrency:      CurrencyUSD,
				CryptoTicker:      "BTC",
				FiatCreditAmount:  decimal.NewFromFloat(9999999.99),
				CryptoDebitAmount: decimal.NewFromFloat(9191919191.1100005),
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
		Amount:   decimal.NewFromFloat(64320.27),
		LastTxTs: ts1,
	})
	require.NoError(t, err, "error expectation condition failed.")

	_, _, err = connection.CryptoPurchase(
		clientID1, CurrencyUSD, decimal.NewFromFloat(22.22), "BTC", decimal.NewFromFloat(4444.4444))
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
				err := connection.Query.cryptoSell(ctx, test.params)
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
		fiatAccount, err := connection.FiatBalance(clientID1, CurrencyUSD)
		require.NoError(t, err, "failed to retrieve Fiat account balance.")
		require.Equal(t, fiatAccount.Balance, decimal.NewFromFloat(73056.28), "Fiat balance mismatch.")

		cryptoAccount, err := connection.CryptoBalance(clientID1, "BTC")
		require.NoError(t, err, "failed to retrieve Crypto account balance.")
		require.Equal(t, cryptoAccount.Balance, decimal.NewFromFloat(4423.92043939), "Crypto balance mismatch.")

		// Check Fiat Journal entries.
		fiatJournal, err := connection.FiatTxDetails(clientID1, txIDValid1)
		require.NoError(t, err, "failed to retrieve Fiat journal for first valid sale.")
		require.Len(t, fiatJournal, 1, "invalid Fiat journal count for first sale.")
		require.Equal(t, testCases[0].params.FiatCreditAmount, fiatJournal[0].Amount,
			"first Fiat credit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetails(fiatOpsAcc, txIDValid1)
		require.NoError(t, err, "failed to retrieve ops Fiat journal for first valid sale.")
		require.Len(t, fiatJournal, 1, "invalid ops Fiat journal count for first sale.")
		require.Equal(t, testCases[0].params.FiatCreditAmount.Mul(negOne), fiatJournal[0].Amount,
			"first ops Fiat credit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetails(clientID1, txIDValid2)
		require.NoError(t, err, "failed to retrieve Fiat journal for second valid sale.")
		require.Len(t, fiatJournal, 1, "invalid Fiat journal count for second sale.")
		require.Equal(t, testCases[1].params.FiatCreditAmount, fiatJournal[0].Amount,
			"second Fiat credit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetails(fiatOpsAcc, txIDValid2)
		require.NoError(t, err, "failed to retrieve ops Fiat journal for second valid sale.")
		require.Len(t, fiatJournal, 1, "invalid ops Fiat journal count for second sale.")
		require.Equal(t, testCases[1].params.FiatCreditAmount.Mul(negOne), fiatJournal[0].Amount,
			"second ops Fiat credit amount mismatched.")

		fiatJournal, err = connection.FiatTxDetails(clientID1, txIDPKR)
		require.NoError(t, err, "failed to retrieve Fiat journal for PKR sale.")
		require.Empty(t, fiatJournal, "Fiat journal entry for PKR sale found.")

		fiatJournal, err = connection.FiatTxDetails(clientID1, txIDBAD)
		require.NoError(t, err, "failed to retrieve Fiat journal for invalid crypto ticker sale.")
		require.Empty(t, fiatJournal, "Fiat journal entry for invalid crypto ticker sale found.")

		fiatJournal, err = connection.FiatTxDetails(clientID1, txIDNoFunds)
		require.NoError(t, err, "failed to retrieve Fiat journal for insufficient funds sale.")
		require.Empty(t, fiatJournal, "Fiat journal entry for insufficient funds sale found.")

		// Check Crypto Journal entries.
		cryptoJournal, err := connection.CryptoTxDetails(clientID1, txIDValid1)
		require.NoError(t, err, "failed to retrieve Crypto journal for first valid sale.")
		require.Len(t, cryptoJournal, 1, "invalid Crypto journal count for first sale.")
		require.Equal(t, testCases[0].params.CryptoDebitAmount.Mul(negOne), cryptoJournal[0].Amount,
			"first Crypto debit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetails(cryptoOpsAcc, txIDValid1)
		require.NoError(t, err, "failed to retrieve ops Crypto journal for first valid sale.")
		require.Len(t, cryptoJournal, 1, "invalid ops Crypto journal count for first sale.")
		require.Equal(t, testCases[0].params.CryptoDebitAmount, cryptoJournal[0].Amount,
			"first ops Crypto debit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetails(clientID1, txIDValid2)
		require.NoError(t, err, "failed to retrieve Crypto journal for second valid sale.")
		require.Len(t, cryptoJournal, 1, "invalid Crypto journal count for second sale.")
		require.Equal(t, testCases[1].params.CryptoDebitAmount.Mul(negOne), cryptoJournal[0].Amount,
			"second Crypto debit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetails(cryptoOpsAcc, txIDValid2)
		require.NoError(t, err, "failed to retrieve ops Crypto journal for second valid sale.")
		require.Len(t, cryptoJournal, 1, "invalid ops Crypto journal count for second sale.")
		require.Equal(t, testCases[1].params.CryptoDebitAmount, cryptoJournal[0].Amount,
			"second Crypto credit amount mismatched.")

		cryptoJournal, err = connection.CryptoTxDetails(clientID1, txIDPKR)
		require.NoError(t, err, "failed to retrieve Crypto journal for PKR sale.")
		require.Empty(t, cryptoJournal, "Crypto journal entry for PKR sale found.")

		cryptoJournal, err = connection.CryptoTxDetails(clientID1, txIDBAD)
		require.NoError(t, err, "failed to retrieve Crypto journal for invalid crypto ticker sale.")
		require.Empty(t, cryptoJournal, "Crypto journal entry for invalid crypto ticker purchase sale.")

		cryptoJournal, err = connection.CryptoTxDetails(clientID1, txIDNoFunds)
		require.NoError(t, err, "failed to retrieve Crypto journal for insufficient funds sale.")
		require.Empty(t, cryptoJournal, "Crypto journal entry for insufficient funds sale found.")
	})
}
func TestCrypto_CryptoGetAllAccounts(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Insert the initial set of test fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// Insert an initial set of test crypto accounts.
	resetTestCryptoAccounts(t, clientID1, clientID2)

	// Reset Crypto Journal entries.
	resetTestCryptoJournal(t)

	// Testing grid.
	testCases := []struct {
		name           string
		ticker         string
		clientID       uuid.UUID
		limitCnt       int32
		expectedRowCnt int
	}{
		{
			name:           "ClientID 1 - ALL",
			ticker:         "BTC",
			clientID:       clientID1,
			limitCnt:       3,
			expectedRowCnt: 3,
		}, {
			name:           "ClientID 1 - Limit 1",
			ticker:         "USDT",
			clientID:       clientID1,
			limitCnt:       1,
			expectedRowCnt: 1,
		}, {
			name:           "ClientID 1 - Base ETH",
			ticker:         "ETH",
			clientID:       clientID1,
			limitCnt:       3,
			expectedRowCnt: 2,
		}, {
			name:           "ClientID 1 - Base USDT",
			ticker:         "USDT",
			clientID:       clientID1,
			limitCnt:       3,
			expectedRowCnt: 1,
		}, {
			name:           "ClientID 1 - Base XRP",
			ticker:         "XRP",
			clientID:       clientID1,
			limitCnt:       3,
			expectedRowCnt: 0,
		}, {
			name:           "Non-existent",
			ticker:         "NON-EXIST",
			clientID:       uuid.UUID{},
			limitCnt:       3,
			expectedRowCnt: 0,
		},
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	for _, testCase := range testCases {
		t.Run("Retrieving "+testCase.name, func(t *testing.T) {
			rows, err := connection.Query.cryptoGetAllAccounts(ctx, &cryptoGetAllAccountsParams{
				ClientID: testCase.clientID,
				Ticker:   testCase.ticker,
				Limit:    testCase.limitCnt,
			})
			require.NoError(t, err, "error expectation failed.")
			require.Len(t, rows, testCase.expectedRowCnt, "expected row count mismatch.")
		})
	}
}

func TestCrypto_CryptoGetAllJournalTransactionPaginated(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Insert the initial set of test fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// Insert an initial set of test crypto accounts.
	resetTestCryptoAccounts(t, clientID1, clientID2)

	// Reset Crypto Journal entries.
	resetTestCryptoJournal(t)

	// Context setup for no hold-and-wait.
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)

	defer cancel()

	// Prepare the journals.
	{
		_, err := connection.FiatExternalTransfer(ctx, &FiatTransactionDetails{
			ClientID: clientID1,
			Currency: CurrencyUSD,
			Amount:   decimal.NewFromFloat(10203040.56),
		})
		require.NoError(t, err, "failed to deposit Fiat money for client 1.")

		_, err = connection.FiatExternalTransfer(ctx, &FiatTransactionDetails{
			ClientID: clientID2,
			Currency: CurrencyUSD,
			Amount:   decimal.NewFromFloat(10304055.78),
		})
		require.NoError(t, err, "failed to deposit Fiat money for client 2.")

		parameters := getTestCryptoPurchaseParams(clientID1, clientID2)
		for range 4 {
			for _, item := range parameters {
				parameter := item
				for idx := range 3 {
					parameter[idx].TransactionID, err = uuid.NewV4()
					require.NoError(t, err, "failed to generate tx id.")

					err := connection.Query.cryptoPurchase(ctx, &parameter[idx])
					require.NoError(t, err, "error expectation failed.")
				}
			}
		}
	}

	// Setup time intervals.
	var (
		timePoint    = time.Now().UTC()
		minuteAhead  = pgtype.Timestamptz{}
		minuteBehind = pgtype.Timestamptz{}
		hourAhead    = pgtype.Timestamptz{}
		hourBehind   = pgtype.Timestamptz{}
	)

	require.NoError(t, minuteAhead.Scan(timePoint.Add(time.Minute)))
	require.NoError(t, minuteBehind.Scan(timePoint.Add(-time.Minute)))
	require.NoError(t, hourAhead.Scan(timePoint.Add(time.Hour)))
	require.NoError(t, hourBehind.Scan(timePoint.Add(-time.Hour)))

	// Test grid.
	testCases := []struct { //nolint:dupl
		name         string
		expectedCont int
		parameters   cryptoGetAllJournalTransactionsPaginatedParams
	}{
		{
			name:         "ClientID1 BTC: Before-After",
			expectedCont: 4,
			parameters: cryptoGetAllJournalTransactionsPaginatedParams{
				ClientID:  clientID1,
				Ticker:    "BTC",
				Offset:    0,
				Limit:     4,
				StartTime: minuteBehind,
				EndTime:   minuteAhead,
			},
		}, {
			name:         "ClientID1 BTC: Before-After, 2 items page 1",
			expectedCont: 2,
			parameters: cryptoGetAllJournalTransactionsPaginatedParams{
				ClientID:  clientID1,
				Ticker:    "BTC",
				Offset:    0,
				Limit:     2,
				StartTime: minuteBehind,
				EndTime:   minuteAhead,
			},
		}, {
			name:         "ClientID1 BTC: Before-After, 2 items page 2",
			expectedCont: 2,
			parameters: cryptoGetAllJournalTransactionsPaginatedParams{
				ClientID:  clientID1,
				Ticker:    "BTC",
				Offset:    2,
				Limit:     4,
				StartTime: minuteBehind,
				EndTime:   minuteAhead,
			},
		}, {
			name:         "ClientID1 BTC: Before-After, 3 items page 2",
			expectedCont: 3,
			parameters: cryptoGetAllJournalTransactionsPaginatedParams{
				ClientID:  clientID1,
				Ticker:    "BTC",
				Offset:    1,
				Limit:     4,
				StartTime: minuteBehind,
				EndTime:   minuteAhead,
			},
		}, {
			name:         "ClientID1 BTC: Before",
			expectedCont: 0,
			parameters: cryptoGetAllJournalTransactionsPaginatedParams{
				ClientID:  clientID1,
				Ticker:    "BTC",
				Offset:    0,
				Limit:     4,
				StartTime: hourBehind,
				EndTime:   minuteBehind,
			},
		}, {
			name:         "ClientID1 BTC: After",
			expectedCont: 0,
			parameters: cryptoGetAllJournalTransactionsPaginatedParams{
				ClientID:  clientID1,
				Ticker:    "BTC",
				Offset:    0,
				Limit:     4,
				StartTime: minuteAhead,
				EndTime:   hourAhead,
			},
		}, {
			name:         "ClientID2 - ETH: Before-After",
			expectedCont: 4,
			parameters: cryptoGetAllJournalTransactionsPaginatedParams{
				ClientID:  clientID2,
				Ticker:    "ETH",
				Offset:    0,
				Limit:     4,
				StartTime: minuteBehind,
				EndTime:   minuteAhead,
			},
		}, {
			name:         "ClientID2 - XFR: Before-After",
			expectedCont: 0,
			parameters: cryptoGetAllJournalTransactionsPaginatedParams{
				ClientID:  clientID2,
				Ticker:    "XFR",
				Offset:    0,
				Limit:     4,
				StartTime: minuteBehind,
				EndTime:   minuteAhead,
			},
		},
	}

	for _, testCase := range testCases {
		parameters := testCase.parameters
		t.Run("Retrieving "+testCase.name, func(t *testing.T) {
			rows, err := connection.Query.cryptoGetAllJournalTransactionsPaginated(ctx, &parameters)
			require.NoError(t, err, "error expectation failed.")
			require.Len(t, rows, testCase.expectedCont, "expected row count mismatch.")
		})
	}
}
