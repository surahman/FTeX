package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/common"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	modelsPostgres "github.com/surahman/FTeX/pkg/models/postgres"
	"github.com/surahman/FTeX/pkg/postgres"
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
	return func(ginCtx *gin.Context) {
		var (
			authToken  *models.JWTAuthResponse
			err        error
			user       modelsPostgres.UserAccount
			httpMsg    string
			httpStatus int
			payload    any
		)

		if err = ginCtx.ShouldBindJSON(&user); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		if authToken, httpMsg, httpStatus, payload, err = common.HTTPRegisterUser(auth, db, logger, &user); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, &models.HTTPError{Message: httpMsg, Payload: payload})

			return
		}

		ginCtx.JSON(http.StatusCreated, authToken)
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
	return func(ginCtx *gin.Context) {
		var (
			err          error
			authToken    *models.JWTAuthResponse
			loginRequest modelsPostgres.UserLoginCredentials
			httpMsg      string
			httpStatus   int
			payload      any
		)

		if err = ginCtx.ShouldBindJSON(&loginRequest); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		if authToken, httpMsg, httpStatus, payload, err = common.HTTPLoginUser(auth, db, logger, &loginRequest); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, &models.HTTPError{Message: httpMsg, Payload: payload})

			return
		}

		ginCtx.JSON(http.StatusOK, authToken)
	}
}

// LoginRefresh validates a JWT token and issues a fresh token.
//
//	@Summary		Refresh a user's JWT by extending its expiration time.
//	@Description	Refreshes a user's JWT by validating it and then issuing a fresh JWT with an extended validity time. JWT must be expiring in under 60 seconds.
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
	return func(ginCtx *gin.Context) {
		var (
			err        error
			clientID   uuid.UUID
			expiresAt  int64
			freshToken *models.JWTAuthResponse
			httpMsg    string
			httpStatus int
		)

		if clientID, expiresAt, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		if freshToken, httpMsg, httpStatus, err = common.HTTPRefreshLogin(auth, db, logger, clientID, expiresAt); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, &models.HTTPError{Message: httpMsg})

			return
		}

		ginCtx.JSON(http.StatusOK, freshToken)
	}
}

// DeleteUser will mark a user as deleted in the database.
//
//	@Summary		Deletes a user. The user must supply their credentials as well as a confirmation message.
//	@Description	Deletes a user stored in the database by marking it as deleted. The user must supply their login credentials as well as complete the following confirmation message:
//	@Description	"I understand the consequences, delete my user account USERNAME HERE"
//	@Tags			user users delete security
//	@Id				deleteUser
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		models.HTTPDeleteUserRequest	true	"The request payload for deleting an account"
//	@Success		204		{object}	models.HTTPSuccess				"message with a confirmation of a deleted user account"
//	@Failure		400		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError				"error message with any available details in payload"
//	@Router			/user/delete [delete]
func DeleteUser(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID      uuid.UUID
			deleteRequest models.HTTPDeleteUserRequest
			err           error
			httpMsg       string
			httpStatus    int
			payload       any
		)

		// Validate the JWT and extract the clientID. Compare the clientID against the deletion request login
		// credentials.
		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		// Get the deletion request from the message body and validate it.
		if err = ginCtx.ShouldBindJSON(&deleteRequest); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		if httpMsg, httpStatus, payload, err = common.HTTPDeleteUser(auth, db, logger, clientID, &deleteRequest); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, &models.HTTPError{Message: httpMsg, Payload: payload})

			return
		}

		ginCtx.JSON(http.StatusNoContent, models.HTTPSuccess{Message: "account successfully deleted"})
	}
}
