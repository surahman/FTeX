package rest

import (
	"fmt"

	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
)

// getTestUsers will generate a number of test users for testing.
func getTestUsers() map[string]*modelsPostgres.UserAccount {
	users := make(map[string]*modelsPostgres.UserAccount)
	username := "username%d"
	password := "user-password-%d"
	firstname := "firstname-%d"
	lastname := "lastname-%d"
	email := "user%d@email-address.com"

	for idx := 1; idx < 5; idx++ {
		uname := fmt.Sprintf(username, idx)
		users[uname] = &modelsPostgres.UserAccount{
			UserLoginCredentials: modelsPostgres.UserLoginCredentials{
				Username: username,
				Password: password,
			},
			FirstName: firstname,
			LastName:  lastname,
			Email:     email,
		}
	}

	return users
}
