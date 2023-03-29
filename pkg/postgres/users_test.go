package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert initial set of test users.
	insertTestUsers(t)

	// Username and account id collisions.
	for key, testCase := range getTestUsers() {
		user := testCase

		t.Run(fmt.Sprintf("Test case %s", key), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

			defer cancel()

			clientID, err := connection.Query.createUser(ctx, &user)
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

	clientID, err := connection.Query.createUser(ctx, &userPass)
	require.NoError(t, err, "user account with non-duplicate key fields should be created.")
	require.True(t, clientID.Valid, "failed to retrieve client id from response")
}

func TestPostgres_DeleteUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert initial set of test users.
	insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Non-existent user.
	result, err := connection.Query.deleteUser(ctx, "non-existent-user")
	require.NoError(t, err, "failed to execute delete for non-existent user.")
	require.Equal(t, int64(0), result.RowsAffected(), "deleted a non-existent user.")

	// Remove all inserted users.
	for key, testCase := range getTestUsers() {
		t.Run(fmt.Sprintf("Deleting User: %s", key), func(t *testing.T) {
			result, err := connection.Query.deleteUser(ctx, testCase.Username)
			require.NoError(t, err, "failed to execute delete on user.")
			require.Equal(t, int64(1), result.RowsAffected(), "failed to execute delete on user.")
		})
	}
}

func TestGetClientIdUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert initial set of test users.
	insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Non-existent user.
	result, err := connection.Query.getClientIdUser(ctx, "non-existent-user")
	require.Error(t, err, "got client id for non-existent user.")
	require.False(t, result.Valid, "client id for a non-existent user is valid.")

	// Get Client IDs for all inserted users.
	for key, testCase := range getTestUsers() {
		t.Run(fmt.Sprintf("Getting Client ID: %s", key), func(t *testing.T) {
			result, err = connection.Query.getClientIdUser(ctx, testCase.Username)
			require.NoError(t, err, "failed to get client id for user.")
			require.True(t, result.Valid, "invalid client id for user.")
		})
	}
}

func TestGetCredentialsUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert initial set of test users.
	insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Non-existent user.
	result, err := connection.Query.getCredentialsUser(ctx, "non-existent-user")
	require.Error(t, err, "got credentials for non-existent user.")
	require.False(t, result.ClientID.Valid, "client id for a non-existent user is valid.")
	require.Equal(t, 0, len(result.Password), "got password for a non-existent user.")

	// Get Client IDs for all inserted users.
	for key, testCase := range getTestUsers() {
		t.Run(fmt.Sprintf("Getting credentials: %s", key), func(t *testing.T) {
			result, err = connection.Query.getCredentialsUser(ctx, testCase.Username)
			require.NoError(t, err, "failed to get client id for user.")
			require.True(t, result.ClientID.Valid, "invalid client id for user.")
			require.Equal(t, testCase.Password, result.Password, "mismatched password for user.")
		})
	}
}

func TestGetInfoUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert initial set of test users.
	insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Non-existent user.
	result, err := connection.Query.getInfoUser(ctx, "non-existent-user")
	require.Error(t, err, "got credentials for non-existent user.")
	require.False(t, result.ClientID.Valid, "client id for a non-existent user is valid.")
	require.False(t, result.IsDeleted, "deleted flag for a non-existent user is set.")
	require.Equal(t, 0, len(result.FirstName), "got first name for a non-existent user.")
	require.Equal(t, 0, len(result.LastName), "got last name for a non-existent user.")
	require.Equal(t, 0, len(result.Email), "got email address for a non-existent user.")

	// Get Client IDs for all inserted users.
	for key, testCase := range getTestUsers() {
		t.Run(fmt.Sprintf("Getting user information: %s", key), func(t *testing.T) {
			result, err = connection.Query.getInfoUser(ctx, testCase.Username)
			require.NoError(t, err, "failed to get client id for user.")
			require.True(t, result.ClientID.Valid, "invalid client id for user.")
			require.False(t, result.IsDeleted, "deleted flag for user is set.")
			require.Equal(t, testCase.FirstName, result.FirstName, "first name mismatch.")
			require.Equal(t, testCase.LastName, result.LastName, "last name mismatch.")
			require.Equal(t, testCase.Email, result.Email, "email address mismatch.")
		})
	}
}
