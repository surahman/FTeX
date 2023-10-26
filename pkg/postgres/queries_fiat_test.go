package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestQueries_FiatCreateAccount(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test fiat accounts.
	clientID1, _ := resetTestFiatAccounts(t)

	err := connection.FiatCreateAccount(clientID1, "EUR")
	require.NoError(t, err, "failed to create Fiat account.")

	err = connection.FiatCreateAccount(clientID1, "EUR")
	require.Error(t, err, "created duplicate Fiat account.")
}

func TestQueries_FiatBalanceCurrency(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test fiat accounts.
	clientID1, _ := resetTestFiatAccounts(t)

	testCases := []struct {
		name      string
		currency  Currency
		expectErr require.ErrorAssertionFunc
	}{
		{
			name:      "AED valid",
			currency:  CurrencyAED,
			expectErr: require.NoError,
		}, {
			name:      "CAD valid",
			currency:  CurrencyCAD,
			expectErr: require.NoError,
		}, {
			name:      "USD valid",
			currency:  CurrencyUSD,
			expectErr: require.NoError,
		}, {
			name:      "EUR invalid",
			currency:  CurrencyEUR,
			expectErr: require.Error,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := connection.FiatBalance(clientID1, testCase.currency)
			testCase.expectErr(t, err, "error expectation failed.")
		})
	}
}

func TestQueries_FiatBalanceCurrencyPaginated(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test fiat accounts.
	clientID1, _ := resetTestFiatAccounts(t)

	testCases := []struct {
		name         string
		baseCurrency Currency
		limit        int32
		expectLen    int
	}{
		{
			name:         "AED All",
			baseCurrency: CurrencyAED,
			limit:        3,
			expectLen:    3,
		}, {
			name:         "AED One",
			baseCurrency: CurrencyAED,
			limit:        1,
			expectLen:    1,
		}, {
			name:         "AED Two",
			baseCurrency: CurrencyAED,
			limit:        2,
			expectLen:    2,
		}, {
			name:         "CAD All",
			baseCurrency: CurrencyCAD,
			limit:        3,
			expectLen:    2,
		}, {
			name:         "CAD All",
			baseCurrency: CurrencyCAD,
			limit:        1,
			expectLen:    1,
		}, {
			name:         "USD All",
			baseCurrency: CurrencyUSD,
			limit:        3,
			expectLen:    1,
		}, {
			name:         "USD One",
			baseCurrency: CurrencyUSD,
			limit:        1,
			expectLen:    1,
		}, {
			name:         "EUR invalid but okay",
			baseCurrency: CurrencyEUR,
			limit:        3,
			expectLen:    1,
		}, {
			name:         "ZWD invalid and not okay",
			baseCurrency: CurrencyZWD,
			limit:        3,
			expectLen:    0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			balances, err := connection.FiatBalancePaginated(clientID1, testCase.baseCurrency, testCase.limit)
			require.NoError(t, err, "failed to retrieve results.")
			require.Len(t, balances, testCase.expectLen, "incorrect number of records returned.")
		})
	}
}

func TestFiat_FiatTransactionsCurrencyPaginated(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert an initial set of test fiat accounts.
	clientID1, clientID2 := resetTestFiatAccounts(t)

	// Reset the test
	resetTestFiatJournal(t, clientID1, clientID2)

	// Context setup for no hold-and-wait.
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)

	defer cancel()

	// Insert some more fiat journal entries for good measure.
	{
		parameters := getTestFiatJournal(clientID1, clientID2)
		for _, item := range parameters {
			parameter := item
			for idx := 0; idx < 3; idx++ {
				_, err := connection.Query.fiatExternalTransferJournalEntry(ctx, &parameter)
				require.NoError(t, err, "error expectation failed.")
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
	testCases := []struct {
		name         string
		expectedCont int
		clientID     uuid.UUID
		currency     Currency
		startTime    pgtype.Timestamptz
		endTime      pgtype.Timestamptz
		offset       int32
		limit        int32
	}{
		{
			name:         "ClientID1 USD: Before-After",
			expectedCont: 4,
			clientID:     clientID1,
			currency:     "USD",
			offset:       0,
			limit:        4,
			startTime:    minuteBehind,
			endTime:      minuteAhead,
		}, {
			name:         "ClientID1 USD: Before-After, 2 items page 1",
			expectedCont: 2,
			clientID:     clientID1,
			currency:     "USD",
			offset:       0,
			limit:        2,
			startTime:    minuteBehind,
			endTime:      minuteAhead,
		}, {
			name:         "ClientID1 USD: Before-After, 2 items page 2",
			expectedCont: 2,
			clientID:     clientID1,
			currency:     "USD",
			offset:       2,
			limit:        4,
			startTime:    minuteBehind,
			endTime:      minuteAhead,
		}, {
			name:         "ClientID1 USD: Before-After, 3 items page 2",
			expectedCont: 3,
			clientID:     clientID1,
			currency:     "USD",
			offset:       1,
			limit:        4,
			startTime:    minuteBehind,
			endTime:      minuteAhead,
		}, {
			name:         "ClientID1 USD: Before",
			expectedCont: 0,
			clientID:     clientID1,
			currency:     "USD",
			offset:       0,
			limit:        4,
			startTime:    hourBehind,
			endTime:      minuteBehind,
		}, {
			name:         "ClientID1 USD: After",
			expectedCont: 0,
			clientID:     clientID1,
			currency:     "USD",
			offset:       0,
			limit:        4,
			startTime:    minuteAhead,
			endTime:      hourAhead,
		}, {
			name:         "ClientID2 - AED: Before-After",
			expectedCont: 4,
			clientID:     clientID2,
			currency:     "AED",
			offset:       0,
			limit:        4,
			startTime:    minuteBehind,
			endTime:      minuteAhead,
		}, {
			name:         "ClientID2 - PKR: Before-After",
			expectedCont: 0,
			clientID:     clientID2,
			currency:     "PKR",
			offset:       0,
			limit:        4,
			startTime:    minuteBehind,
			endTime:      minuteAhead,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Retrieving %s", testCase.name), func(t *testing.T) {
			rows, err := connection.FiatTransactionsPaginated(testCase.clientID, testCase.currency,
				testCase.limit, testCase.offset, testCase.startTime, testCase.endTime)
			require.NoError(t, err, "error expectation failed.")
			require.Len(t, rows, testCase.expectedCont, "expected row count mismatch.")
		})
	}
}
