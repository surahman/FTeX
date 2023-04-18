package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
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
//	@Success		200		{object}	models.JWTAuthResponse	"a valid JWT token for the new account"
//	@Failure		400		{object}	models.HTTPError		"error message with any available details in payload"
//	@Failure		409		{object}	models.HTTPError		"error message with any available details in payload"
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

		if clientID, err = db.CreateUser(&user); err != nil {
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

		context.JSON(http.StatusOK, authToken)
	}
}
