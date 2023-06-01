package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
	"github.com/surahman/FTeX/pkg/utilities"
	"github.com/surahman/FTeX/pkg/validator"
	"go.uber.org/zap"
)

// OpenCrypto will handle an HTTP request to open a Cryptocurrency account.
//
//	@Summary		Open a Cryptocurrency account.
//	@Description	Creates a Cryptocurrency account for a specified ticker, to be provided as the currency in the request, for a user by creating a row in the Crypto Accounts table.
//	@Tags			crypto cryptocurrency currency open
//	@Id				openCrypto
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		models.HTTPOpenCurrencyAccountRequest	true	"cryptocurrency ticker for new account"
//	@Success		201		{object}	models.HTTPSuccess						"a message to confirm the creation of an account"
//	@Failure		400		{object}	models.HTTPError						"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError						"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError						"error message with any available details in payload"
//	@Router			/crypto/open [post]
func OpenCrypto(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID uuid.UUID
			err      error
			request  models.HTTPOpenCurrencyAccountRequest
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

		if err = db.CryptoCreateAccount(clientID, request.Currency); err != nil {
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
			models.HTTPSuccess{Message: "account created", Payload: []string{clientID.String(), request.Currency}})
	}
}

// OfferCrypto will handle an HTTP request to get a purchase or sale offer for a Fiat or Cryptocurrency.
//
//	@Summary		Purchase or sell a Cryptocurrency and using a Fiat currency.
//	@Description	Purchase or sell a Fiat currency using a Cryptocurrency. The amount must be a positive number with at most two or eight decimal places for Fiat and Cryptocurrencies respectively. Both currency accounts must be opened beforehand.
//	@Tags			fiat crypto cryptocurrency currency sell sale offer
//	@Id				sellOfferCrypto
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		models.HTTPCryptoOfferRequest	true	"the Cryptocurrency ticker, Fiat currency code, and amount to be converted in the source currency"
//	@Success		200		{object}	models.HTTPSuccess				"a message to confirm the purchase rate for a Fiat or Cryptocurrency"
//	@Failure		400		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError				"error message with any available details in payload"
//	@Router			/crypto/offer [post]
func OfferCrypto(
	logger *logger.Logger,
	auth auth.Auth,
	cache redis.Redis,
	quotes quotes.Quotes,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID      uuid.UUID
			err           error
			request       models.HTTPCryptoOfferRequest
			offer         models.HTTPExchangeOfferResponse
			status        int
			statusMessage string
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
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		offer, status, statusMessage, err = utilities.HTTPPrepareCryptoOffer(auth, cache, logger, quotes,
			clientID, request.SourceCurrency, request.DestinationCurrency, request.SourceAmount, *request.IsPurchase)
		if err != nil {
			httpErr := &models.HTTPError{Message: statusMessage}
			if statusMessage == constants.GetInvalidRequest() {
				httpErr.Payload = err.Error()
			}

			ginCtx.AbortWithStatusJSON(status, httpErr)

			return
		}

		offer.ClientID = clientID

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "crypto rate offer", Payload: offer})
	}
}

// ExchangeCrypto will handle an HTTP request to execute and complete a Cryptocurrency purchase/sale offer. The default
// action is to purchase a Cryptocurrency using a Fiat.
//
//	@Summary		Transfer funds between Fiat and Crypto accounts using a valid Offer ID.
//	@Description	Purchase or sell a Cryptocurrency to/from a Fiat currency accounts. The Offer ID must be valid and have expired.
//	@Tags			crypto fiat currency cryptocurrency exchange convert offer transfer execute
//	@Id				exchangeCrypto
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			offerID	body		models.HTTPTransferRequest	true	"the two currency codes and amount to be converted"
//	@Success		200		{object}	models.HTTPSuccess			"a message to confirm the conversion of funds"
//	@Failure		400		{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		408		{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError			"error message with any available details in payload"
//	@Router			/crypto/exchange/ [post]
func ExchangeCrypto(
	logger *logger.Logger,
	auth auth.Auth,
	cache redis.Redis,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			err      error
			clientID uuid.UUID
			request  models.HTTPTransferRequest
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
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		receipt, status, httpErrMsg, err := utilities.HTTPExchangeCrypto(auth, cache, db, logger, clientID, request.OfferID)
		if err != nil {
			ginCtx.AbortWithStatusJSON(status, &models.HTTPError{Message: httpErrMsg})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "funds exchange transfer successful", Payload: receipt})
	}
}

// BalanceCurrencyCrypto will handle an HTTP request to retrieve a balance for a specific Cryptocurrency.
//
//	@Summary		Retrieve balance for a specific Cryptocurrency.
//	@Description	Retrieves the balance for a specific Cryptocurrency. The currency ticker must be supplied as a query parameter.
//	@Tags			crypto cryptocurrency currency balance
//	@Id				balanceCurrencyCrypto
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			ticker	path		string				true	"the Cryptocurrency ticker to retrieve the balance for"
//	@Success		200		{object}	models.HTTPSuccess	"the details for a specific currency account"
//	@Failure		400		{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		404		{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError	"error message with any available details in payload"
//	@Router			/crypto/info/balance/{ticker} [get]
func BalanceCurrencyCrypto(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			accDetails postgres.CryptoAccount
			clientID   uuid.UUID
			ticker     = ginCtx.Param("ticker")
			err        error
		)

		// Extract and validate the currency.
		if len(ticker) < 1 || len(ticker) > 6 {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid currency", Payload: ticker})

			return
		}

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		if accDetails, err = db.CryptoBalanceCurrency(clientID, ticker); err != nil {
			var balanceErr *postgres.Error
			if !errors.As(err, &balanceErr) {
				logger.Info("failed to unpack Crypto account balance currency error", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			ginCtx.AbortWithStatusJSON(balanceErr.Code, models.HTTPError{Message: balanceErr.Message})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "account balance", Payload: accDetails})
	}
}

// TxDetailsCrypto will handle an HTTP request to retrieve information for a specific Crypto transaction.
//
//	@Summary		Retrieve transaction details for a specific transactionID.
//	@Description	Retrieves the transaction details for a specific transactionID. The transaction ID must be supplied as a query parameter.
//	@Tags			crypto cryptocurrency transactionID transaction details
//	@Id				txDetailsCrypto
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			transactionID	path		string				true	"the transaction ID to retrieve the details for"
//	@Success		200				{object}	models.HTTPSuccess	"the transaction details for a specific transaction ID"
//	@Failure		400				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		403				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		404				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		500				{object}	models.HTTPError	"error message with any available details in payload"
//	@Router			/crypto/info/transaction/{transactionID} [get]
func TxDetailsCrypto(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID      uuid.UUID
			transactionID = ginCtx.Param("transactionID")
			err           error
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		// Extract and validate the transactionID.
		journalEntries, status, errMsg, err := utilities.HTTPTxDetails(db, logger, clientID, transactionID)
		if err != nil {
			ginCtx.AbortWithStatusJSON(status, models.HTTPError{Message: errMsg})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "transaction details", Payload: journalEntries})
	}
}
