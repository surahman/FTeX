package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueries_FiatCreateAccount(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
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

	// Insert initial set of test fiat accounts.
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
			_, err := connection.FiatBalanceCurrency(clientID1, testCase.currency)
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

	// Insert initial set of test fiat accounts.
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
			balances, err := connection.FiatBalanceCurrencyPaginated(clientID1, testCase.baseCurrency, testCase.limit)
			require.NoError(t, err, "failed to retrieve results.")
			require.Equal(t, testCase.expectLen, len(balances), "incorrect number of records returned.")
		})
	}
}
