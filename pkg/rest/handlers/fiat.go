package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/constants"
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
//	@Param			user	body		models.HTTPOpenCurrencyAccount	true	"currency code for new account"
//	@Success		201		{object}	models.HTTPSuccess				"a message to confirm the creation of an account"
//	@Failure		400		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError				"error message with any available details in payload"
//	@Router			/fiat/open [post]
func OpenFiat(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID      uuid.UUID
			currency      postgres.Currency
			err           error
			originalToken = ginCtx.GetHeader(authHeaderKey)
			request       models.HTTPOpenCurrencyAccount
		)

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: "validation", Payload: err})

			return
		}

		// Extract and validate the currency.
		if err = currency.Scan(request.Currency); err != nil || !currency.Valid() {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid currency", Payload: request.Currency})

			return
		}

		if clientID, _, err = auth.ValidateJWT(originalToken); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		if err = db.FiatCreateAccount(clientID, currency); err != nil {
			var createErr *postgres.Error
			if !errors.As(err, &createErr) {
				logger.Info("failed to unpack open Fiat account error", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			ginCtx.AbortWithStatusJSON(createErr.Code, models.HTTPError{Message: createErr.Message})

			return
		}

		ginCtx.JSON(http.StatusCreated,
			models.HTTPSuccess{Message: "account created", Payload: []string{clientID.String(), request.Currency}})
	}
}

// DepositFiat will handle an HTTP request to deposit funds into a Fiat account.
//
//	@Summary		Deposit funds into a Fiat account.
//	@Description	Deposit funds into a Fiat account in a specific currency for a user. The amount must be a positive number with at most two decimal places.
//	@Tags			fiat currency deposit
//	@Id				depositFiat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			user	body		models.HTTPDepositCurrency	true	"currency code and amount to be deposited"
//	@Success		201		{object}	models.HTTPSuccess			"a message to confirm the creation of an account"
//	@Failure		400		{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError			"error message with any available details in payload"
//	@Router			/fiat/deposit [post]
func DepositFiat(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID        uuid.UUID
			currency        postgres.Currency
			err             error
			originalToken   = ginCtx.GetHeader(authHeaderKey)
			request         models.HTTPDepositCurrency
			transferReceipt *postgres.FiatAccountTransferResult
		)

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: "validation", Payload: err})

			return
		}

		// Extract and validate the currency.
		if err = currency.Scan(request.Currency); err != nil || !currency.Valid() {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid currency", Payload: request.Currency})

			return
		}

		// Check for correct decimal places.
		if !request.Amount.Equal(request.Amount.Truncate(constants.GetDecimalPlacesFiat())) {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "too many decimal places", Payload: request.Amount.String()})

			return
		}

		if clientID, _, err = auth.ValidateJWT(originalToken); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		if transferReceipt, err = db.FiatExternalTransfer(context.Background(),
			&postgres.FiatTransactionDetails{
				ClientID: clientID,
				Currency: currency,
				Amount:   request.Amount}); err != nil {
			var createErr *postgres.Error
			if !errors.As(err, &createErr) {
				logger.Info("failed to unpack deposit Fiat account error", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			ginCtx.AbortWithStatusJSON(createErr.Code, models.HTTPError{Message: createErr.Message})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "funds successfully transferred", Payload: transferReceipt})
	}
}
