package postgres

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
)

func TestTransactions_FiatTransactionsDetails_LessComparator(t *testing.T) {
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

	// Insert an initial set of test Fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the Fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// End of test-expected totals.
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

	ftexID, err := connection.queries.userGetClientId(ctx, constants.SpecialAccountFiat())
	require.NoError(t, err, "failed to retrieve FTeX internal ID.")

	for _, testCase := range testCases {
		test := testCase

		go func() {
			t.Run("Transferring "+test.name, func(t *testing.T) {
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
				journalEntry, err := connection.Query.fiatGetJournalTransaction(ctx, &fiatGetJournalTransactionParams{
					ClientID: test.accountDetails.ClientID,
					TxID:     transferResult.TxID,
				})
				require.NoError(t, err, "failed to retrieve journal entries for deposit.")
				require.Len(t, journalEntry, 1, "incorrect journal entry for deposit.")

				journalEntry, err = connection.Query.fiatGetJournalTransaction(ctx, &fiatGetJournalTransactionParams{
					ClientID: ftexID,
					TxID:     transferResult.TxID,
				})
				require.NoError(t, err, "failed to retrieve internal journal entries for internal.")
				require.Len(t, journalEntry, 1, "incorrect journal entry for internal.")
			})
		}()
	}

	// Check the end of test totals.
	wg.Wait()

	t.Run("Checking end totals", func(t *testing.T) {
		actual, err := connection.Query.fiatGetAccount(ctx, &fiatGetAccountParams{
			ClientID: clientID1,
			Currency: CurrencyUSD,
		})
		require.NoError(t, err, "failed to fiat account.")
		require.Equal(t, expectedTotal, actual.Balance, "end of test expected totals mismatched.")
	})
}

func TestTransactions_FiatExternalTransfer_Mock(t *testing.T) {
	t.Parallel()

	txDetails := FiatTransactionDetails{}
	rowLockDecimal := decimal.Decimal{}
	journalEntryRow := fiatExternalTransferJournalEntryRow{}
	accountBalanceRow := fiatUpdateAccountBalanceRow{}

	// Test grid.
	testCases := []struct {
		name                string
		expectedErrMsg      string
		rowLockReturn       *decimal.Decimal
		rowLockError        error
		rowLockTimes        int
		extJournalReturn    *fiatExternalTransferJournalEntryRow
		extJournalError     error
		extJournalTimes     int
		updateBalanceReturn *fiatUpdateAccountBalanceRow
		updateBalanceError  error
		updateBalanceTimes  int
	}{
		{
			name:                "Row lock failure.",
			expectedErrMsg:      "row lock failure",
			rowLockTimes:        1,
			rowLockReturn:       &rowLockDecimal,
			rowLockError:        errors.New("row lock failure"),
			extJournalTimes:     0,
			extJournalReturn:    &journalEntryRow,
			extJournalError:     nil,
			updateBalanceTimes:  0,
			updateBalanceReturn: &accountBalanceRow,
			updateBalanceError:  nil,
		}, {
			name:                "Journal entry failure.",
			expectedErrMsg:      "journal entry failure",
			rowLockTimes:        1,
			rowLockReturn:       &rowLockDecimal,
			rowLockError:        nil,
			extJournalTimes:     1,
			extJournalReturn:    &journalEntryRow,
			extJournalError:     errors.New("journal entry failure"),
			updateBalanceTimes:  0,
			updateBalanceReturn: &accountBalanceRow,
			updateBalanceError:  nil,
		}, {
			name:                "Account balance update failure.",
			expectedErrMsg:      "account balance update failure",
			rowLockTimes:        1,
			rowLockReturn:       &rowLockDecimal,
			rowLockError:        nil,
			extJournalTimes:     1,
			extJournalReturn:    &journalEntryRow,
			extJournalError:     nil,
			updateBalanceTimes:  1,
			updateBalanceReturn: &accountBalanceRow,
			updateBalanceError:  errors.New("account balance update failure"),
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run("Failure: "+test.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockQuerier := NewMockQuerier(mockCtrl)

			// Configure mock expectations.
			gomock.InOrder(
				mockQuerier.EXPECT().
					fiatRowLockAccount(gomock.Any(), gomock.Any()).
					Return(*test.rowLockReturn, test.rowLockError).
					Times(test.rowLockTimes),

				mockQuerier.EXPECT().
					fiatExternalTransferJournalEntry(gomock.Any(), gomock.Any()).
					Return(*test.extJournalReturn, test.extJournalError).
					Times(test.extJournalTimes),

				mockQuerier.EXPECT().
					fiatUpdateAccountBalance(gomock.Any(), gomock.Any()).
					Return(*test.updateBalanceReturn, test.updateBalanceError).
					Times(test.updateBalanceTimes),
			)

			// Check for error.
			_, err := fiatExternalTransfer(context.TODO(), connection.logger, mockQuerier, &txDetails)
			require.Error(t, err, "failed to get error.")
			require.Contains(t, err.Error(), test.expectedErrMsg, "error messages mismatched.")
		})
	}
}

func TestTransactions_FiatTransactionRowLockAndBalanceCheck(t *testing.T) {
	t.Parallel()

	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test Fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the Fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	// Initial balances, transaction amounts, and timestamp.
	var (
		balanceClientID1 = decimal.NewFromFloat(52145.77)
		balanceClientID2 = decimal.NewFromFloat(1921.68)
		amount20k        = decimal.NewFromFloat(20987.65)
		amount1k         = decimal.NewFromFloat(1234.56)
		amountNegative   = decimal.NewFromFloat(-1010.11)
		txTimestamp      = pgtype.Timestamptz{}
	)

	require.NoError(t, txTimestamp.Scan(time.Now().UTC()), "failed to create current timestamp.")

	// Configure context for test suite.
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	t.Cleanup(func() {
		cancel()
	})

	// Update base balances in accounts to test from.
	_, err := connection.Query.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: clientID1,
		Currency: CurrencyUSD,
		Amount:   balanceClientID1,
		LastTxTs: txTimestamp,
	})
	require.NoError(t, err, "failed to set base balance for Client1 in USD")

	_, err = connection.Query.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: clientID2,
		Currency: CurrencyAED,
		Amount:   balanceClientID2,
		LastTxTs: txTimestamp,
	})
	require.NoError(t, err, "failed to set base balance for Client2 in AED")

	// Test grid.
	testCases := []struct {
		name           string
		expectErrMsg   string
		source         FiatTransactionDetails
		destination    FiatTransactionDetails
		errExpectation require.ErrorAssertionFunc
	}{
		{
			name:         "Client1USD 20k, Client2AED 1k",
			expectErrMsg: "insufficient balance",
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
			name:         "Client2AED 1k, Client1USD 20k",
			expectErrMsg: "insufficient balance",
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
			name:         "Client1USD 1k, Client2AED 20k",
			expectErrMsg: "insufficient balance",
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
			name:         "Client2AED 20k, Client1USD 1k",
			expectErrMsg: "insufficient balance",
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
		}, {
			name:         "Client1USD Negative, Client2AED 1k",
			expectErrMsg: "negative",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   amountNegative,
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyAED,
				Amount:   amount1k,
			},
			errExpectation: require.Error,
		}, {
			name:         "Client1USD 20k, Client2AED Negative",
			expectErrMsg: "negative",
			source: FiatTransactionDetails{
				ClientID: clientID1,
				Currency: CurrencyUSD,
				Amount:   amount20k,
			},
			destination: FiatTransactionDetails{
				ClientID: clientID2,
				Currency: CurrencyAED,
				Amount:   amountNegative,
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
				require.Contains(t, err.Error(), test.expectErrMsg, "failure not related to balance.")
			}
		})
	}
}

func TestTransactions_FiatTransactionRowLockAndBalanceCheck_mock(t *testing.T) {
	t.Parallel()

	firstUUID, err := uuid.FromString("515fb04e-ea91-460b-ad4e-487cb673601e")
	require.NoError(t, err, "failed to parse first UUID.")

	secondUUID, err := uuid.FromString("d68d52a1-aa7c-4301-a27f-611196726edc")
	require.NoError(t, err, "failed to parse second UUID.")

	var (
		uuid1USD = &FiatTransactionDetails{ClientID: firstUUID, Currency: CurrencyUSD}
		uuid2USD = &FiatTransactionDetails{ClientID: secondUUID, Currency: CurrencyUSD}
	)

	testCases := []struct {
		name                 string
		expectedErrMsg       string
		srcAccount           *FiatTransactionDetails
		dstAccount           *FiatTransactionDetails
		firstRowLockBalance  decimal.Decimal
		firstRowLockErr      error
		firstRowLockTimes    int
		secondRowLockBalance decimal.Decimal
		secondRowLockErr     error
		secondRowLockTimes   int
	}{
		{
			name:                 "First row lock failure.",
			expectedErrMsg:       "first row lock failure",
			srcAccount:           uuid1USD,
			dstAccount:           uuid2USD,
			firstRowLockBalance:  decimal.Decimal{},
			firstRowLockErr:      errors.New("first row lock failure"),
			firstRowLockTimes:    1,
			secondRowLockBalance: decimal.Decimal{},
			secondRowLockErr:     nil,
			secondRowLockTimes:   0,
		}, {
			name:                 "Second row lock failure.",
			expectedErrMsg:       "second row lock failure",
			srcAccount:           uuid1USD,
			dstAccount:           uuid2USD,
			firstRowLockBalance:  decimal.Decimal{},
			firstRowLockErr:      nil,
			firstRowLockTimes:    1,
			secondRowLockBalance: decimal.Decimal{},
			secondRowLockErr:     errors.New("second row lock failure"),
			secondRowLockTimes:   1,
		}, {
			name:           "Insufficient balance failure.",
			expectedErrMsg: "insufficient balance",
			srcAccount: &FiatTransactionDetails{
				ClientID: uuid1USD.ClientID,
				Currency: CurrencyUSD,
				Amount:   decimal.NewFromFloat(101.1),
			},
			dstAccount:           uuid2USD,
			firstRowLockBalance:  decimal.Decimal{},
			firstRowLockErr:      nil,
			firstRowLockTimes:    1,
			secondRowLockBalance: decimal.NewFromFloat(100.0),
			secondRowLockErr:     nil,
			secondRowLockTimes:   1,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockQuerier := NewMockQuerier(mockCtrl)

			// Configure mock expectations.
			gomock.InOrder(
				mockQuerier.EXPECT().
					fiatRowLockAccount(gomock.Any(), gomock.Any()).
					Return(test.firstRowLockBalance, test.firstRowLockErr).
					Times(test.firstRowLockTimes),

				mockQuerier.EXPECT().
					fiatRowLockAccount(gomock.Any(), gomock.Any()).
					Return(test.secondRowLockBalance, test.secondRowLockErr).
					Times(test.secondRowLockTimes),
			)

			// Check for error.
			err := fiatTransactionRowLockAndBalanceCheck(context.TODO(), mockQuerier, test.srcAccount, test.dstAccount)
			require.Error(t, err, "failed to get error.")
			require.Contains(t, err.Error(), test.expectedErrMsg, "error messages mismatched.")
		})
	}
}

func TestTransactions_FiatInternalTransfer(t *testing.T) { //nolint:maintidx
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test Fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the Fiat journal entries.
	resetTestFiatJournal(t, clientID1, clientID2)

	var (
		expectedTotalClientID1 = decimal.NewFromFloat(517306.27)
		expectedTotalClientID2 = decimal.NewFromFloat(190859.81)
		balanceClientID1       = decimal.NewFromFloat(521459.77)
		balanceClientID2       = decimal.NewFromFloat(192103.68)
		txTimestamp            = pgtype.Timestamptz{}
	)

	require.NoError(t, txTimestamp.Scan(time.Now().UTC()), "failed to create current timestamp.")

	// Configure context for test suite.
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)

	defer cancel()

	// Update base balances in accounts to test from.
	_, err := connection.Query.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: clientID1,
		Currency: CurrencyUSD,
		Amount:   balanceClientID1,
		LastTxTs: txTimestamp,
	})
	require.NoError(t, err, "failed to set base balance for Client1 in USD")

	_, err = connection.Query.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
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
				journalEntry, err := connection.Query.fiatGetJournalTransaction(ctx, &fiatGetJournalTransactionParams{
					ClientID: test.source.ClientID,
					TxID:     dstResult.TxID,
				})
				require.NoError(t, err, "failed to retrieve journal entries for transaction.")
				require.Len(t, journalEntry, 1, "incorrect row count retrieved for source.")

				journalEntry, err = connection.Query.fiatGetJournalTransaction(ctx, &fiatGetJournalTransactionParams{
					ClientID: test.destination.ClientID,
					TxID:     dstResult.TxID,
				})
				require.NoError(t, err, "failed to retrieve journal entries for transaction.")
				require.Len(t, journalEntry, 1, "incorrect row count retrieved for destination.")
			})
		}()
	}

	// Wait (tie-threads).
	wg.Wait()

	t.Run("Checking end totals", func(t *testing.T) {
		client1, err := connection.Query.fiatGetAccount(ctx, &fiatGetAccountParams{
			ClientID: clientID1,
			Currency: CurrencyUSD,
		})
		require.NoError(t, err, "failed to fiat account.")
		require.Equal(t, expectedTotalClientID1, client1.Balance, "client 1's balance mismatched.")

		client2, err := connection.Query.fiatGetAccount(ctx, &fiatGetAccountParams{
			ClientID: clientID2,
			Currency: CurrencyCAD,
		})
		require.NoError(t, err, "failed to fiat account.")
		require.Equal(t, expectedTotalClientID2, client2.Balance, "client 2's balance mismatched.")
	})
}

func TestTransactions_FiatInternalTransfer_Mock(t *testing.T) {
	t.Parallel()

	txDetails := FiatTransactionDetails{}
	rowLockDecimal := decimal.Decimal{}
	journalEntryRow := fiatInternalTransferJournalEntryRow{}
	balanceUpdateRow := fiatUpdateAccountBalanceRow{}

	testCases := []struct {
		name           string
		expectedErrMsg string
		rowLockError   error
		journalReturn  *fiatInternalTransferJournalEntryRow
		journalError   error
		journalTimes   int
		creditReturn   *fiatUpdateAccountBalanceRow
		creditError    error
		creditTimes    int
		debitReturn    *fiatUpdateAccountBalanceRow
		debitError     error
		debitTimes     int
	}{
		{
			name:           "Row lock and balance failure.",
			expectedErrMsg: "row lock failure",
			rowLockError:   errors.New("row lock failure"),
			journalReturn:  &journalEntryRow,
			journalError:   nil,
			journalTimes:   0,
			creditReturn:   &balanceUpdateRow,
			creditError:    nil,
			creditTimes:    0,
			debitReturn:    &balanceUpdateRow,
			debitError:     nil,
			debitTimes:     0,
		}, {
			name:           "Journal entry failure.",
			expectedErrMsg: "journal entry failure",
			rowLockError:   nil,
			journalReturn:  &journalEntryRow,
			journalError:   errors.New("journal entry failure"),
			journalTimes:   1,
			creditReturn:   &balanceUpdateRow,
			creditError:    nil,
			creditTimes:    0,
			debitReturn:    &balanceUpdateRow,
			debitError:     nil,
			debitTimes:     0,
		}, {
			name:           "Balance credit failure.",
			expectedErrMsg: "balance credit failure",
			rowLockError:   nil,
			journalReturn:  &journalEntryRow,
			journalError:   nil,
			journalTimes:   1,
			creditReturn:   &balanceUpdateRow,
			creditError:    errors.New("balance credit failure"),
			creditTimes:    1,
			debitReturn:    &balanceUpdateRow,
			debitError:     nil,
			debitTimes:     0,
		}, {
			name:           "Balance debit failure.",
			expectedErrMsg: "balance debit failure",
			rowLockError:   nil,
			journalReturn:  &journalEntryRow,
			journalError:   nil,
			journalTimes:   1,
			creditReturn:   &balanceUpdateRow,
			creditError:    nil,
			creditTimes:    1,
			debitReturn:    &balanceUpdateRow,
			debitError:     errors.New("balance debit failure"),
			debitTimes:     1,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockQuerier := NewMockQuerier(mockCtrl)

			// Configure mock expectations.
			gomock.InOrder(
				mockQuerier.EXPECT().
					fiatRowLockAccount(gomock.Any(), gomock.Any()).
					Return(rowLockDecimal, test.rowLockError).
					AnyTimes(),

				mockQuerier.EXPECT().
					fiatInternalTransferJournalEntry(gomock.Any(), gomock.Any()).
					Return(*test.journalReturn, test.journalError).
					Times(test.journalTimes),

				mockQuerier.EXPECT().
					fiatUpdateAccountBalance(gomock.Any(), gomock.Any()).
					Return(*test.creditReturn, test.creditError).
					Times(test.creditTimes),

				mockQuerier.EXPECT().
					fiatUpdateAccountBalance(gomock.Any(), gomock.Any()).
					Return(*test.debitReturn, test.debitError).
					Times(test.debitTimes),
			)

			// Check for error.
			_, _, err := fiatInternalTransfer(context.TODO(), connection.logger, mockQuerier, &txDetails, &txDetails)
			require.Error(t, err, "failed to get error.")
			require.Contains(t, err.Error(), test.expectedErrMsg, "error messages mismatched.")
		})
	}
}
