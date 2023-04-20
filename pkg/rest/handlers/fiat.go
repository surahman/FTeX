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

// OpenFiat will handle an HTTP request to open a Fiat account.
//
//	@Summary		Open a Fiat account.
//	@Description	Creates a Fiat account for a specific currency for a user by creating a row in the Fiat Accounts table.
//	@Tags			fiat currency open
//	@Id				openFiat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			user	body		models.HTTPOpenCurrencyAccount	true	"Currency code for new account"
//	@Success		201		{object}	models.HTTPSuccess				"a message to confirm the creation of an account"
//	@Failure		400		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError				"error message with any available details in payload"
//	@Router			/fiat/open [post]
func OpenFiat(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(context *gin.Context) {
		var (
			clientID      uuid.UUID
			currency      postgres.Currency
			err           error
			originalToken = context.GetHeader(authHeaderKey)
			request       models.HTTPOpenCurrencyAccount
		)

		if err = context.ShouldBindJSON(&request); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: "validation", Payload: err})

			return
		}

		// Extract and validate the currency.
		if err = currency.Scan(request.Currency); err != nil || !currency.Valid() {
			context.AbortWithStatusJSON(http.StatusBadRequest,
				&models.HTTPError{Message: "invalid currency", Payload: request.Currency})

			return
		}

		if clientID, _, err = auth.ValidateJWT(originalToken); err != nil {
			context.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		if err = db.FiatCreateAccount(clientID, currency); err != nil {
			var createErr *postgres.Error
			if !errors.As(err, &createErr) {
				logger.Info("failed to unpack open Fiat account error", zap.Error(err))
				context.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			context.AbortWithStatusJSON(createErr.Code, models.HTTPError{Message: createErr.Message})

			return
		}

		context.JSON(http.StatusCreated,
			models.HTTPSuccess{Message: "account created", Payload: []string{clientID.String(), request.Currency}})
	}
}
