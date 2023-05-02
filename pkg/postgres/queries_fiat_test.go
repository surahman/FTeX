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
		expectNil require.ValueAssertionFunc
	}{
		{
			name:      "AED valid",
			currency:  CurrencyAED,
			expectErr: require.NoError,
			expectNil: require.NotNil,
		}, {
			name:      "CAD valid",
			currency:  CurrencyCAD,
			expectErr: require.NoError,
			expectNil: require.NotNil,
		}, {
			name:      "USD valid",
			currency:  CurrencyUSD,
			expectErr: require.NoError,
			expectNil: require.NotNil,
		}, {
			name:      "UER invalid",
			currency:  CurrencyEUR,
			expectErr: require.Error,
			expectNil: require.Nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			balance, err := connection.FiatBalanceCurrency(clientID1, testCase.currency)
			testCase.expectErr(t, err, "error expectation failed.")
			testCase.expectNil(t, balance, "nil balance expectation failed.")
		})
	}
}
