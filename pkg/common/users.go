package common

import (
	"errors"
	"fmt"
	"net/http"
	"time"

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

// HTTPLoginUser will complete a login request for a user.
func HTTPLoginUser(auth auth.Auth, db postgres.Postgres, logger *logger.Logger,
	loginRequest *modelsPostgres.UserLoginCredentials) (*models.JWTAuthResponse, string, int, any, error) {
	var (
		err            error
		authToken      *models.JWTAuthResponse
		clientID       uuid.UUID
		hashedPassword string
	)

	if err = validator.ValidateStruct(loginRequest); err != nil {
		return nil, constants.ValidationString(), http.StatusBadRequest, fmt.Errorf("%w", err), fmt.Errorf("%w", err)
	}

	if clientID, hashedPassword, err = db.UserCredentials(loginRequest.Username); err != nil {
		return nil, "invalid credentials", http.StatusForbidden, nil, fmt.Errorf("%w", err)
	}

	if err = auth.CheckPassword(hashedPassword, loginRequest.Password); err != nil {
		return nil, "invalid username or password", http.StatusForbidden, nil, fmt.Errorf("%w", err)
	}

	if authToken, err = auth.GenerateJWT(clientID); err != nil {
		logger.Error("failure generating JWT during login", zap.Error(err))

		return nil, err.Error(), http.StatusInternalServerError, nil, fmt.Errorf("%w", err)
	}

	return authToken, "", 0, nil, nil
}

// HTTPRefreshLogin validates a JWT token and issues a fresh token.
func HTTPRefreshLogin(auth auth.Auth, db postgres.Postgres, logger *logger.Logger,
	clientID uuid.UUID, expiresAt int64) (*models.JWTAuthResponse, string, int, error) {
	var (
		err         error
		freshToken  *models.JWTAuthResponse
		accountInfo modelsPostgres.User
	)

	if accountInfo, err = db.UserGetInfo(clientID); err != nil {
		logger.Warn("failed to read user record for a valid JWT",
			zap.String("username", accountInfo.Username), zap.Error(err))

		return nil, constants.RetryMessageString(), http.StatusInternalServerError, fmt.Errorf("%w", err)
	}

	if accountInfo.IsDeleted {
		logger.Warn("attempt to refresh a JWT for a deleted user", zap.String("clientID", accountInfo.Username))

		return nil, "invalid token", http.StatusForbidden, fmt.Errorf("%w", err)
	}

	// Do not refresh tokens that are outside the refresh threshold. Tokens could expire during the execution of
	// this handler, but expired ones would be rejected during token validation. Thus, it is not necessary to
	// re-check expiration.
	if expiresAt-time.Now().Unix() > auth.RefreshThreshold() {
		return nil, fmt.Sprintf("JWT is still valid for more than %d seconds", auth.RefreshThreshold()),
			http.StatusNotExtended, fmt.Errorf("%w", err)
	}

	if freshToken, err = auth.GenerateJWT(clientID); err != nil {
		logger.Error("failure generating JWT during token refresh", zap.Error(err))

		return nil, err.Error(), http.StatusInternalServerError, fmt.Errorf("%w", err)
	}

	return freshToken, "", 0, nil
}

// HTTPDeleteUser validates a JWT token and issues a fresh token.
func HTTPDeleteUser(auth auth.Auth, db postgres.Postgres, logger *logger.Logger, clientID uuid.UUID,
	deleteRequest *models.HTTPDeleteUserRequest) (string, int, any, error) {
	var (
		err         error
		userAccount modelsPostgres.User
	)

	if err = validator.ValidateStruct(deleteRequest); err != nil {
		return constants.ValidationString(), http.StatusBadRequest, fmt.Errorf("%w", err), fmt.Errorf("%w", err)
	}

	// Get user account information to validate against.
	if userAccount, err = db.UserGetInfo(clientID); err != nil {
		logger.Warn("failed to read user record during an account deletion request",
			zap.String("clientID", clientID.String()), zap.Error(err))

		return constants.RetryMessageString(), http.StatusInternalServerError, nil, fmt.Errorf("%w", err)
	}

	// Validate if the user account is already deleted.
	if userAccount.Username != deleteRequest.Username {
		return "invalid deletion request", http.StatusForbidden, nil, fmt.Errorf("%w", err)
	}

	// Check the confirmation message.
	if fmt.Sprintf(constants.DeleteUserAccountConfirmation(), userAccount.Username) != deleteRequest.Confirmation {
		msg := "incorrect or incomplete deletion request confirmation"

		return msg, http.StatusBadRequest, nil, errors.New(msg)
	}

	// Check to make sure the account is not already deleted.
	if userAccount.IsDeleted {
		logger.Warn("attempt to delete an already deleted user account", zap.String("username", userAccount.Username))

		msg := "user account is already deleted"

		return msg, http.StatusForbidden, nil, errors.New(msg)
	}

	if err = auth.CheckPassword(userAccount.Password, deleteRequest.Password); err != nil {
		msg := "invalid user credentials"

		return msg, http.StatusForbidden, nil, errors.New(msg)
	}

	// Mark the account as deleted.
	if err = db.UserDelete(clientID); err != nil {
		logger.Warn("failed to mark a user record as deleted", zap.String("username", userAccount.Username), zap.Error(err))

		return constants.RetryMessageString(), http.StatusInternalServerError, nil, errors.New(constants.RetryMessageString())
	}

	return "", 0, nil, nil
}
