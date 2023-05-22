package postgres

import (
	"testing"

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
