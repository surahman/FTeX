package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		return
	}

	// Insert an initial set of test users.
	insertTestUsers(t)

	// Username and account id collisions.
	for key, testCase := range getTestUsers() {
		user := testCase

		t.Run("Test case "+key, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

			defer cancel()

			clientID, err := connection.Query.userCreate(ctx, &user)
			require.Error(t, err, "user account creation collision did not result in an error.")
			require.True(t, clientID.IsNil(), "incorrectly retrieved client id from response")
		})
	}

	// New user with different username and account but duplicated fields.
	userPass := userCreateParams{
		Username:  "user-5",
		Password:  "user-pwd-1",
		FirstName: "firstname-1",
		LastName:  "lastname-1",
		Email:     "user1@email-address.com",
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	clientID, err := connection.Query.userCreate(ctx, &userPass)
	require.NoError(t, err, "user account with non-duplicate key fields should be created.")
	require.False(t, clientID.IsNil(), "failed to retrieve client id from response")
}

func TestPostgres_DeleteUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert an initial set of test users.
	clientIDs := insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Non-existent user.
	invalidID, err := uuid.NewV1()
	require.NoError(t, err, "failed to generate invalid client id.")
	rowsAffected, err := connection.Query.userDelete(ctx, invalidID)
	require.NoError(t, err, "failed to execute delete for non-existent user.")
	require.Zero(t, rowsAffected, "deleted a non-existent user.")

	// Remove all inserted users.
	for _, clientID := range clientIDs {
		t.Run("Deleting User: "+clientID.String(), func(t *testing.T) {
			rowsAffected, err = connection.Query.userDelete(ctx, clientID)
			require.NoError(t, err, "failed to execute delete on user.")
			require.Equal(t, int64(1), rowsAffected, "failed to execute delete on user.")
		})
	}
}

func TestGetClientIdUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert an initial set of test users.
	insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Non-existent user.
	result, err := connection.Query.userGetClientId(ctx, "non-existent-user")
	require.Error(t, err, "got client id for non-existent user.")
	require.True(t, result.IsNil(), "client id for a non-existent user is valid.")

	// Get Client IDs for all inserted users.
	for key, testCase := range getTestUsers() {
		t.Run("Getting Client ID: "+key, func(t *testing.T) {
			result, err = connection.Query.userGetClientId(ctx, testCase.Username)
			require.NoError(t, err, "failed to get client id for user.")
			require.False(t, result.IsNil(), "invalid client id for user.")
		})
	}
}

func TestGetCredentialsUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert an initial set of test users.
	insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Non-existent user.
	result, err := connection.Query.userGetCredentials(ctx, "non-existent-user")
	require.Error(t, err, "got credentials for non-existent user.")
	require.True(t, result.ClientID.IsNil(), "client id for a non-existent user is valid.")
	require.Empty(t, result.Password, "got password for a non-existent user.")

	// Get Client IDs for all inserted users.
	for key, testCase := range getTestUsers() {
		t.Run("Getting credentials: "+key, func(t *testing.T) {
			result, err = connection.Query.userGetCredentials(ctx, testCase.Username)
			require.NoError(t, err, "failed to get client id for user.")
			require.False(t, result.ClientID.IsNil(), "invalid client id for user.")
			require.Equal(t, testCase.Password, result.Password, "mismatched password for user.")
		})
	}
}

func TestGetInfoUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert an initial set of test users.
	clientIDs := insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	invalidID, err := uuid.NewV1()
	require.NoError(t, err, "failed to generate invalid id.")

	// Non-existent user.
	result, err := connection.Query.userGetInfo(ctx, invalidID)
	require.Error(t, err, "got credentials for non-existent user.")
	require.Empty(t, result.Username, "got username for a non-existent user.")
	require.True(t, result.ClientID.IsNil(), "client id for a non-existent user is valid.")
	require.Empty(t, result.FirstName, "got first name for a non-existent user.")
	require.Empty(t, result.LastName, "got last name for a non-existent user.")
	require.Empty(t, result.Email, "got email address for a non-existent user.")
	require.False(t, result.IsDeleted, "deleted flag for a non-existent user is set.")

	// Get Client IDs for all inserted users.
	testUsers := getTestUsers()

	for _, clientID := range clientIDs {
		t.Run("Getting user information: "+clientID.String(), func(t *testing.T) {
			result, err = connection.Query.userGetInfo(ctx, clientID)
			require.NoError(t, err, "failed to get client id for user.")

			expected := testUsers[result.Username]

			require.False(t, result.ClientID.IsNil(), "invalid client id for user.")
			require.False(t, result.IsDeleted, "deleted flag for user is set.")
			require.Equal(t, expected.FirstName, result.FirstName, "first name mismatch.")
			require.Equal(t, expected.LastName, result.LastName, "last name mismatch.")
			require.Equal(t, expected.Email, result.Email, "email address mismatch.")
		})
	}
}

func TestPostgres_IsDeletedUser(t *testing.T) {
	// Skip integration tests for short test runs.
	if testing.Short() {
		t.Skip()
	}

	// Insert an initial set of test users.
	clientIDs := insertTestUsers(t)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)

	defer cancel()

	// Non-existent user.
	invalidID, err := uuid.NewV1()
	require.NoError(t, err, "failed to generate invalid client id.")
	isDeleted, err := connection.Query.userIsDeleted(ctx, invalidID)
	require.Error(t, err, "failed to execute delete for non-existent user.")
	require.False(t, isDeleted, "deleted status set on non-existent user.")

	// Remove all inserted users.
	for _, clientID := range clientIDs {
		t.Run("Checking deleted status of user: "+clientID.String(), func(t *testing.T) {
			// Before deletion.
			isDeleted, err = connection.Query.userIsDeleted(ctx, clientID)
			require.NoError(t, err, "failed to retrieve a/c active status user.")
			require.False(t, isDeleted, "incorrect a/c active status for user.")

			// After deletion.
			rowsAffected, err := connection.Query.userDelete(ctx, clientID)
			require.NoError(t, err, "failed to execute delete on user.")
			require.Equal(t, int64(1), rowsAffected, "failed to execute delete on user.")

			isDeleted, err = connection.Query.userIsDeleted(ctx, clientID)
			require.NoError(t, err, "failed to retrieve a/c inactive status user.")
			require.True(t, isDeleted, "incorrect a/c inactive status for user.")
		})
	}
}
