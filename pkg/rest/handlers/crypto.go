package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
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
//	@Param			offerID	body		models.HTTPTransferRequest	true	"the two currency code and amount to be converted"
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
			err          error
			clientID     uuid.UUID
			request      models.HTTPTransferRequest
			offer        models.HTTPExchangeOfferResponse
			receipt      models.HTTPCryptoTransferResponse
			offerID      string
			cryptoTicker string
			fiatTicker   string
			cryptoAmount decimal.Decimal
			fiatAmount   decimal.Decimal
			precision    = constants.GetDecimalPlacesCrypto()
			transferFunc = db.CryptoPurchase
			fiatCurrency []postgres.Currency
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

		// Extract Offer ID from request.
		{
			var rawOfferID []byte
			if rawOfferID, err = auth.DecryptFromString(request.OfferID); err != nil {
				logger.Warn("failed to decrypt Offer ID for Crypto transfer request", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					&models.HTTPError{Message: "please retry your request later"})

				return
			}

			offerID = string(rawOfferID)
		}

		// Retrieve the offer from Redis. Once retrieved, the entry must be removed from the cache to block re-use of
		// the offer. If a database update fails below this point the user will need to re-request an offer.
		{
			var (
				status int
				msg    string
			)
			if offer, status, msg, err = utilities.HTTPGetCachedOffer(cache, logger, offerID); err != nil {
				ginCtx.AbortWithStatusJSON(status, &models.HTTPError{Message: msg})

				return
			}
		}

		// Verify that offer is a Crypto offer.
		if !(offer.IsCryptoSale || offer.IsCryptoPurchase) {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				&models.HTTPError{Message: "invalid Cryptocurrency exchange offer"})

			return
		}

		// Verify that the client IDs match.
		if clientID != offer.ClientID {
			logger.Warn("clientID mismatch with the Crypto Offer stored in Redis",
				zap.Strings("Requester & Offer Client IDs", []string{clientID.String(), offer.ClientID.String()}))
			ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
				&models.HTTPError{Message: "please retry your request later"})

			return
		}

		// Configure transaction parameters. Default action should be to purchase a Cryptocurrency using Fiat.
		fiatTicker = offer.SourceAcc
		fiatAmount = offer.DebitAmount
		cryptoTicker = offer.DestinationAcc
		cryptoAmount = offer.Amount

		if offer.IsCryptoSale {
			cryptoTicker = offer.SourceAcc
			cryptoAmount = offer.DebitAmount
			fiatTicker = offer.DestinationAcc
			fiatAmount = offer.Amount
			precision = constants.GetDecimalPlacesCrypto()
			transferFunc = db.CryptoSell
		}

		// Get Fiat currency code.
		if fiatCurrency, err = utilities.HTTPValidateOfferRequest(
			offer.Amount, precision, fiatTicker); err != nil {
			logger.Warn("failed to extract Fiat currency from Crypto exchange offer",
				zap.Error(err))
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		// Execute transfer.
		if receipt.FiatTxReceipt, receipt.CryptoTxReceipt, err =
			transferFunc(clientID, fiatCurrency[0], fiatAmount, cryptoTicker, cryptoAmount); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusInternalServerError, &models.HTTPError{Message: err.Error()})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "funds exchange transfer successful", Payload: receipt})
	}
}
