package rest

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

// RegisterUser will handle an HTTP request to create a user.
//
//	@Summary		Register a user.
//	@Description	Creates a user account by inserting credentials into the database. A hashed password is stored.
//	@Tags			user users register security
//	@Id				registerUser
//	@Accept			json
//	@Produce		json
//	@Param			user	body		models.UserAccount		true	"Username, password, first and last name, email address of user"
//	@Success		201		{object}	models.JWTAuthResponse	"a valid JWT token for the new account"
//	@Failure		400		{object}	models.HTTPError		"error message with any available details in payload"
//	@Failure		404		{object}	models.HTTPError		"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError		"error message with any available details in payload"
//	@Router			/user/register [post]
func RegisterUser(logger *logger.Logger, auth auth.Auth, db postgres.Postgres) gin.HandlerFunc {
	return func(context *gin.Context) {
		var (
			authToken *models.JWTAuthResponse
			clientID  uuid.UUID
			err       error
			user      modelsPostgres.UserAccount
		)

		if err = context.ShouldBindJSON(&user); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&user); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: "validation", Payload: err})

			return
		}

		if user.Password, err = auth.HashPassword(user.Password); err != nil {
			logger.Error("failure hashing password", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &models.HTTPError{Message: err.Error()})

			return
		}

		if clientID, err = db.UserRegister(&user); err != nil {
			var registerErr *postgres.Error
			if !errors.As(err, &registerErr) {
				logger.Warn("failed to extract create user account error", zap.Error(err))
				context.AbortWithStatusJSON(http.StatusInternalServerError, "account creation failed, please try again later")

				return
			}

			context.AbortWithStatusJSON(registerErr.Code, &models.HTTPError{Message: err.Error()})

			return
		}

		if authToken, err = auth.GenerateJWT(clientID); err != nil {
			logger.Error("failure generating JWT during account creation", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &models.HTTPError{Message: err.Error()})

			return
		}

		context.JSON(http.StatusCreated, authToken)
	}
}

// LoginUser validates login credentials and generates a JWT.
//
//	@Summary		Login a user.
//	@Description	Logs in a user by validating credentials and returning a JWT.
//	@Tags			user users login security
//	@Id				loginUser
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		models.UserLoginCredentials	true	"Username and password to login with"
//	@Success		200			{object}	models.JWTAuthResponse		"a valid JWT token for the new account"
//	@Failure		400			{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		409			{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		500			{object}	models.HTTPError			"error message with any available details in payload"
//	@Router			/user/login [post]
func LoginUser(logger *logger.Logger, auth auth.Auth, db postgres.Postgres) gin.HandlerFunc {
	return func(context *gin.Context) {
		var (
			err            error
			authToken      *models.JWTAuthResponse
			loginRequest   modelsPostgres.UserLoginCredentials
			clientID       uuid.UUID
			hashedPassword string
		)

		if err = context.ShouldBindJSON(&loginRequest); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&loginRequest); err != nil {
			context.JSON(http.StatusBadRequest, &models.HTTPError{Message: "validation", Payload: err})

			return
		}

		if clientID, hashedPassword, err = db.UserCredentials(loginRequest.Username); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "invalid username or password"})

			return
		}

		if err = auth.CheckPassword(hashedPassword, loginRequest.Password); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "invalid username or password"})

			return
		}

		if authToken, err = auth.GenerateJWT(clientID); err != nil {
			logger.Error("failure generating JWT during login", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &models.HTTPError{Message: err.Error()})

			return
		}

		context.JSON(http.StatusOK, authToken)
	}
}

// LoginRefresh validates a JWT token and issues a fresh token.
//
//	@Summary		Refresh a user's JWT by extending its expiration time.
//	@Description	Refreshes a user's JWT by validating it and then issuing a fresh JWT with an extended validity time.
//	@Description	JWT must be expiring in under 60 seconds.
//	@Tags			user users login refresh security
//	@Id				loginRefresh
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	models.JWTAuthResponse	"A new valid JWT"
//	@Failure		403	{object}	models.HTTPError		"error message with any available details in payload"
//	@Failure		500	{object}	models.HTTPError		"error message with any available details in payload"
//	@Failure		510	{object}	models.HTTPError		"error message with any available details in payload"
//	@Router			/user/refresh [post]
func LoginRefresh(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(context *gin.Context) {
		var (
			err           error
			freshToken    *models.JWTAuthResponse
			clientID      uuid.UUID
			accountInfo   modelsPostgres.User
			expiresAt     int64
			originalToken = context.GetHeader(authHeaderKey)
		)

		if clientID, expiresAt, err = auth.ValidateJWT(originalToken); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		if accountInfo, err = db.UserGetInfo(clientID); err != nil {
			logger.Warn("failed to read user record for a valid JWT",
				zap.String("username", accountInfo.Username), zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError,
				&models.HTTPError{Message: "please retry your request later"})

			return
		}

		if accountInfo.IsDeleted {
			logger.Warn("attempt to refresh a JWT for a deleted user", zap.String("clientID", accountInfo.Username))
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "invalid token"})

			return
		}

		// Do not refresh tokens that are outside the refresh threshold. Tokens could expire during the execution of
		// this handler but expired ones would be rejected during token validation. Thus, it is not necessary to
		// re-check expiration.
		if expiresAt-time.Now().Unix() > auth.RefreshThreshold() {
			context.AbortWithStatusJSON(http.StatusNotExtended,
				&models.HTTPError{Message: fmt.Sprintf("JWT is still valid for more than %d seconds", auth.RefreshThreshold())})

			return
		}

		if freshToken, err = auth.GenerateJWT(clientID); err != nil {
			logger.Error("failure generating JWT during token refresh", zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError, &models.HTTPError{Message: err.Error()})

			return
		}

		context.JSON(http.StatusOK, freshToken)
	}
}

// DeleteUser will mark a user as deleted in the database.
//
//	@Summary		Deletes a user. The user must supply their credentials as well as a confirmation message.
//	@Description	Deletes a user stored in the database by marking it as deleted. The user must supply their login
//	@Description	credentials as well as complete the following confirmation message:
//	@Description	"I understand the consequences, delete my user account USERNAME HERE"
//	@Tags			user users delete security
//	@Id				deleteUser
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		models.HTTPDeleteUserRequest	true	"The request payload for deleting an account"
//	@Success		200		{object}	models.HTTPSuccess				"message with a confirmation of a deleted user account"
//	@Failure		400		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError				"error message with any available details in payload"
//	@Router			/user/delete [delete]
func DeleteUser(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(context *gin.Context) {
		var (
			clientID      uuid.UUID
			deleteRequest models.HTTPDeleteUserRequest
			err           error
			userAccount   modelsPostgres.User
			jwt           = context.GetHeader(authHeaderKey)
		)

		// Get the delete request from the message body and validate it.
		if err = context.ShouldBindJSON(&deleteRequest); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&deleteRequest); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: "validation", Payload: err})

			return
		}

		// Validate the JWT and extract the username. Compare the username against the deletion request login
		// credentials.
		if clientID, _, err = auth.ValidateJWT(jwt); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		// Get user account information to validate against.
		if userAccount, err = db.UserGetInfo(clientID); err != nil {
			context.AbortWithStatusJSON(http.StatusInternalServerError,
				&models.HTTPError{Message: "please retry your request later"})

			return
		}

		// Validate if the user account is already deleted.
		if userAccount.Username != deleteRequest.Username {
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "invalid deletion request"})

			return
		}

		// Check confirmation message.
		if fmt.Sprintf(constants.GetDeleteUserAccountConfirmation(), userAccount.Username) != deleteRequest.Confirmation {
			context.AbortWithStatusJSON(http.StatusBadRequest,
				&models.HTTPError{Message: "incorrect or incomplete deletion request confirmation"})

			return
		}

		// Check to make sure the account is not already deleted.
		if userAccount.IsDeleted {
			logger.Warn("attempt to delete an already deleted user account", zap.String("username", userAccount.Username))
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "user account is already deleted"})

			return
		}

		if err = auth.CheckPassword(userAccount.Password, deleteRequest.Password); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "invalid username or password"})

			return
		}

		// Mark account as deleted.
		if err = db.UserDelete(clientID); err != nil {
			logger.Warn("failed to mark a user record as deleted", zap.String("username", userAccount.Username), zap.Error(err))
			context.AbortWithStatusJSON(http.StatusInternalServerError,
				&models.HTTPError{Message: "please retry your request later"})

			return
		}

		context.JSON(http.StatusOK, models.HTTPSuccess{Message: "account successfully deleted"})
	}
}
