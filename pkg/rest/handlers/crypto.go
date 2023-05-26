package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/validator"
	"go.uber.org/zap"
)

// OpenCrypto will handle an HTTP request to open a Cryptocurrency account.
//
//	@Summary		Open a Cryptocurrency account.
//	@Description	Creates a Cryptocurrency account for a specified ticker for a user by creating a row in the Crypto Accounts table.
//	@Tags			crypto cryptocurrency currency open
//	@Id				openCrypto
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		models.HTTPOpenCryptoAccountRequest	true	"cryptocurrency ticker for new account"
//	@Success		201		{object}	models.HTTPSuccess					"a message to confirm the creation of an account"
//	@Failure		400		{object}	models.HTTPError					"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError					"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError					"error message with any available details in payload"
//	@Router			/crypto/open [post]
func OpenCrypto(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID uuid.UUID
			err      error
			request  models.HTTPOpenCryptoAccountRequest
		)

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: "validation", Payload: err})

			return
		}

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		if err = db.CryptoCreateAccount(clientID, request.Ticker); err != nil {
			var createErr *postgres.Error
			if !errors.As(err, &createErr) {
				logger.Info("failed to unpack open Crypto account error", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			ginCtx.AbortWithStatusJSON(createErr.Code, models.HTTPError{Message: createErr.Message})

			return
		}

		ginCtx.JSON(http.StatusCreated,
			models.HTTPSuccess{Message: "account created", Payload: []string{clientID.String(), request.Ticker}})
	}
}