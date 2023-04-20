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
