package postgres

import "testing"

func TestFiat_CreateFiatAccount(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert initial set of test fiat accounts.
	insertTestFiatAccounts(t)
}
