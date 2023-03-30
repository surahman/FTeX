package postgres

import (
	"context"
	"testing"
)

func TestTransactions_FiatExternalTransfer(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert test users.
	insertTestUsers(t)

	// Insert initial set of test fiat accounts.
	clientID1, _ := resetTestFiatAccounts(t)

	xferDetails := FiatAccountDetails{
		ClientID: clientID1,
		Currency: "USD",
	}

	result, err := connection.FiatExternalTransfer(context.Background(), &xferDetails, 1443.9786)
	_, _ = result, err

	result, err = connection.FiatExternalTransfer(context.Background(), &xferDetails, 1443.9786)
	_, _ = result, err
}
