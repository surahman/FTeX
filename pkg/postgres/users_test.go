package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateUsers(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert initial set of test users.
	insertTestUsers(t)

	// Username and account id collisions.
	for key, user := range getTestUsers() {
		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

			defer cancel()

			clientID, err := connection.db.Query.createUser(ctx, user)
			require.Error(t, err, "user account creation collision did not result in an error.")
			require.False(t, clientID.Valid, "incorrectly retrieved client id from response")
		})
	}

	// New user with different username and account but duplicated fields.
	userPass := createUserParams{
		Username:  "user-5",
		Password:  "user-pwd-1",
		FirstName: "firstname-1",
		LastName:  "lastname-1",
		Email:     "user1@email-address.com",
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	clientID, err := connection.db.Query.createUser(ctx, userPass)
	require.NoError(t, err, "user account with non-duplicate key fields should be created.")
	require.True(t, clientID.Valid, "failed to retrieve client id from response")
}
