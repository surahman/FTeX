package postgres

import (
	"testing"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	models "github.com/surahman/FTeX/pkg/models/postgres"
)

func TestQueries_RegisterUser(t *testing.T) {
	// Integration test check.
	if testing.Short() {
		t.Skip()
	}

	// Insert initial set of test users.
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
	clientID, err := connection.CreateUser(&testUser)
	require.NoError(t, err, "initial user insertion failed.")
	require.False(t, clientID.IsNil(), "returned client id was invalid")

	// Create user collision.
	clientID, err = connection.CreateUser(&testUser)
	require.Error(t, err, "inserted duplicate user.")
	require.True(t, clientID.IsNil(), "return duplicate client id was valid")
}
