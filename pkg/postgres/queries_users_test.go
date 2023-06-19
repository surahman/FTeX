package postgres

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	models "github.com/surahman/FTeX/pkg/models/postgres"
)

func TestQueries_UserRegister(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert an initial set of test users.
	insertTestUsers(t)

	testUser := models.UserAccount{
		UserLoginCredentials: models.UserLoginCredentials{
			Username: xid.New().String(),
			Password: xid.New().String(),
		},
		FirstName: xid.New().String(),
		LastName:  xid.New().String(),
		Email:     xid.New().String(),
	}

	// Create new user.
	clientID, err := connection.UserRegister(&testUser)
	require.NoError(t, err, "initial user insertion failed.")
	require.False(t, clientID.IsNil(), "returned client id was invalid.")

	// Create user collision.
	clientID, err = connection.UserRegister(&testUser)
	require.Error(t, err, "inserted duplicate user.")
	require.True(t, clientID.IsNil(), "return duplicate client id was valid.")
}

func TestQueries_UserCredentials(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert an initial set of test users.
	insertTestUsers(t)

	const uname = "username1"

	// Active account.
	clientID, hashedPass, err := connection.UserCredentials(uname)
	require.NoError(t, err, "failed to retrieve user credentials.")
	require.False(t, clientID.IsNil(), "retrieved an invalid clientID.")
	require.True(t, len(hashedPass) > 0, "retrieved an invalid password.")

	// Deleted account.
	rowsAffected, err := connection.Query.userDelete(context.TODO(), clientID)
	require.NoError(t, err, "errored whilst trying to delete user.")
	require.Equal(t, rowsAffected, int64(1), "no users were deleted.")

	clientID, hashedPass, err = connection.UserCredentials(uname)
	require.Error(t, err, "retrieved deleted user credentials.")
	require.True(t, clientID.IsNil(), "retrieved an valid clientID for a deleted account.")
	require.True(t, len(hashedPass) == 0, "retrieved a password for a deleted account.")

	// Non-existent user.
	clientID, hashedPass, err = connection.UserCredentials("invalid-username")
	require.Error(t, err, "retrieved invalid users' credentials.")
	require.True(t, clientID.IsNil(), "retrieved an invalid users' clientID.")
	require.True(t, len(hashedPass) == 0, "retrieved a password for an invalid user.")
}

func TestQueries_UserGetInfo(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert an initial set of test users.
	insertTestUsers(t)

	// Get valid account information.
	const uname = "username1"
	expectedAccount := getTestUsers()[uname]

	clientID, err := connection.queries.userGetClientId(context.TODO(), uname)
	require.NoError(t, err, "failed to retrieve client id for username1.")
	actualAccount, err := connection.UserGetInfo(clientID)
	require.NoError(t, err, "failed to retrieve account info for username1.")
	require.Equal(t, expectedAccount.Username, actualAccount.Username, "username mismatch.")
	require.Equal(t, expectedAccount.Password, actualAccount.Password, "password mismatch.")
	require.Equal(t, expectedAccount.FirstName, actualAccount.FirstName, "firstname mismatch.")
	require.Equal(t, expectedAccount.LastName, actualAccount.LastName, "lastname mismatch.")
	require.Equal(t, expectedAccount.Email, actualAccount.Email, "email address mismatch.")
	require.False(t, actualAccount.IsDeleted, "should not be deleted.")

	// Invalid clientID.
	invalidID, err := uuid.NewV1()
	require.NoError(t, err, "failed to generate invalid id.")
	_, err = connection.UserGetInfo(invalidID)
	require.Error(t, err, "retrieved an invalid id.")
}
