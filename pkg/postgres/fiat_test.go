package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFiat_CreateFiatAccount(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert initial set of test fiat accounts.
	clientID1, clientID2 := insertTestFiatAccounts(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Account collisions test.
	for key, testCase := range getTestFiatAccounts(clientID1, clientID2) {
		parameters := testCase

		t.Run(fmt.Sprintf("Inserting %s", key), func(t *testing.T) {
			for _, param := range parameters {
				accInfo := param
				rowCount, err := connection.Query.createFiatAccount(ctx, &accInfo)
				require.Error(t, err, "did not error whilst inserting duplicate fiat account.")
				require.Equal(t, int64(0), rowCount, "rows were added.")
			}
		})
	}
}
