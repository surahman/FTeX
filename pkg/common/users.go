package common

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/validator"
	"go.uber.org/zap"
)

// HTTPGenerateTestUsers will generate a number of test users for testing.
func HTTPGenerateTestUsers() map[string]*modelsPostgres.UserAccount {
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

// HTTPRegisterUser will create a row in the database's users' table corresponding to a new user.
func HTTPRegisterUser(auth auth.Auth, db postgres.Postgres, logger *logger.Logger, user *modelsPostgres.UserAccount) (
	*models.JWTAuthResponse, string, int, any, error) {
	var (
		authToken *models.JWTAuthResponse
		clientID  uuid.UUID
		err       error
	)

	if err = validator.ValidateStruct(user); err != nil {
		return nil, constants.ValidationString(), http.StatusBadRequest, fmt.Errorf("%w", err), fmt.Errorf("%w", err)
	}

	if user.Password, err = auth.HashPassword(user.Password); err != nil {
		logger.Error("failure hashing password", zap.Error(err))

		return nil, constants.RetryMessageString(), http.StatusInternalServerError, nil, fmt.Errorf("%w", err)
	}

	if clientID, err = db.UserRegister(user); err != nil {
		var registerErr *postgres.Error
		if !errors.As(err, &registerErr) {
			logger.Warn("failed to extract create user account error", zap.Error(err))

			return nil, constants.RetryMessageString(), http.StatusInternalServerError, nil, fmt.Errorf("%w", err)
		}

		return nil, err.Error(), registerErr.Code, nil, fmt.Errorf("%w", err)
	}

	if authToken, err = auth.GenerateJWT(clientID); err != nil {
		logger.Error("failure generating JWT during account creation", zap.Error(err))

		return nil, err.Error(), http.StatusInternalServerError, nil, fmt.Errorf("%w", err)
	}

	return authToken, "", 0, nil, nil
}
