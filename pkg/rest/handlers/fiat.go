package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/rs/xid"
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

// OpenFiat will handle an HTTP request to open a Fiat account.
//
//	@Summary		Open a Fiat account.
//	@Description	Creates a Fiat account for a specific currency for a user by creating a row in the Fiat Accounts table.
//	@Tags			fiat currency open
//	@Id				openFiat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		models.HTTPOpenCurrencyAccountRequest	true	"currency code for new account"
//	@Success		201		{object}	models.HTTPSuccess						"a message to confirm the creation of an account"
//	@Failure		400		{object}	models.HTTPError						"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError						"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError						"error message with any available details in payload"
//	@Router			/fiat/open [post]
func OpenFiat(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID uuid.UUID
			currency postgres.Currency
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

		// Extract and validate the currency.
		if err = currency.Scan(request.Currency); err != nil || !currency.Valid() {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid currency", Payload: request.Currency})

			return
		}

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
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
//	@Param			request	body		models.HTTPDepositCurrencyRequest	true	"currency code and amount to be deposited"
//	@Success		200		{object}	models.HTTPSuccess					"a message to confirm the deposit of funds"
//	@Failure		400		{object}	models.HTTPError					"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError					"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError					"error message with any available details in payload"
//	@Router			/fiat/deposit [post]
func DepositFiat(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID        uuid.UUID
			currency        postgres.Currency
			err             error
			request         models.HTTPDepositCurrencyRequest
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
		if !request.Amount.Equal(request.Amount.Truncate(constants.GetDecimalPlacesFiat())) || request.Amount.IsNegative() {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid amount", Payload: request.Amount.String()})

			return
		}

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
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

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "funds successfully transferred", Payload: *transferReceipt})
	}
}

// validateSourceDestinationAmount will validate the source and destination accounts as well as the source amount.
func validateSourceDestinationAmount(src, dst string, sourceAmount decimal.Decimal) (
	postgres.Currency, postgres.Currency, error) {
	var (
		err         error
		source      postgres.Currency
		destination postgres.Currency
	)

	// Extract and validate the currency.
	if err = source.Scan(src); err != nil || !source.Valid() {
		return source, destination, fmt.Errorf("invalid source currency %s", src)
	}

	if err = destination.Scan(dst); err != nil || !destination.Valid() {
		return source, destination, fmt.Errorf("invalid destination currency %s", dst)
	}

	// Check for correct decimal places.
	if !sourceAmount.Equal(sourceAmount.Truncate(constants.GetDecimalPlacesFiat())) || sourceAmount.IsNegative() {
		return source, destination, fmt.Errorf("invalid source amount %s", sourceAmount.String())
	}

	return source, destination, nil
}

// ExchangeOfferFiat will handle an HTTP request to get an exchange offer of funds between two Fiat currencies.
//
//	@Summary		Exchange quote for Fiat funds between two Fiat currencies.
//	@Description	Exchange quote for Fiat funds between two Fiat currencies. The amount must be a positive number with at most two decimal places and both currency accounts must be opened.
//	@Tags			fiat currency exchange convert offer transfer
//	@Id				exchangeOfferFiat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			request	body		models.HTTPFiatExchangeOfferRequest	true	"the two currency code and amount to be converted"
//	@Success		200		{object}	models.HTTPSuccess					"a message to confirm the conversion rate for a currency"
//	@Failure		400		{object}	models.HTTPError					"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError					"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError					"error message with any available details in payload"
//	@Router			/fiat/exchange/offer [post]
func ExchangeOfferFiat(
	logger *logger.Logger,
	auth auth.Auth,
	cache redis.Redis,
	quotes quotes.Quotes,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			err     error
			request models.HTTPFiatExchangeOfferRequest
			offer   models.HTTPFiatExchangeOfferResponse
			offerID = xid.New().String()
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
		if _, _, err = validateSourceDestinationAmount(
			request.SourceCurrency, request.DestinationCurrency, request.SourceAmount); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid request", Payload: err.Error()})

			return
		}

		if offer.ClientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		// Compile exchange rate offer.
		if offer.Rate, offer.Amount, err = quotes.FiatConversion(
			request.SourceCurrency, request.DestinationCurrency, request.SourceAmount, nil); err != nil {
			logger.Warn("failed to retrieve quote for Fiat currency conversion", zap.Error(err))
			ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
				&models.HTTPError{Message: "please retry your request later"})

			return
		}

		offer.SourceAcc = request.SourceCurrency
		offer.DestinationAcc = request.DestinationCurrency
		offer.DebitAmount = request.SourceAmount
		offer.Expires = time.Now().Add(constants.GetFiatOfferTTL()).Unix()

		// Encrypt offer ID before returning to client.
		if offer.OfferID, err = auth.EncryptToString([]byte(offerID)); err != nil {
			logger.Warn("failed to encrypt offer ID for Fiat conversion", zap.Error(err))
			ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
				&models.HTTPError{Message: "please retry your request later"})

			return
		}

		// Store the offer in Redis.
		if err = cache.Set(offerID, &offer, constants.GetFiatOfferTTL()); err != nil {
			logger.Warn("failed to store Fiat conversion offer in cache", zap.Error(err))
			ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
				&models.HTTPError{Message: "please retry your request later"})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "conversion rate offer", Payload: offer})
	}
}

// getCachedOffer will retrieve and then evict an offer from the Redis cache.
func getCachedOffer(cache redis.Redis, logger *logger.Logger, offerID string) (
	models.HTTPFiatExchangeOfferResponse, int, string, error) {
	var (
		err   error
		offer models.HTTPFiatExchangeOfferResponse
	)

	// Retrieve the offer from Redis.
	if err = cache.Get(offerID, &offer); err != nil {
		var redisErr *redis.Error

		// If we have a valid Redis package error AND the error is that the key is not found.
		if errors.As(err, &redisErr) && redisErr.Is(redis.ErrCacheMiss) {
			return offer, http.StatusRequestTimeout, "Fiat exchange rate offer has expired", fmt.Errorf("%w", err)
		}

		logger.Warn("unknown error occurred whilst retrieving Fiat Offer from Redis", zap.Error(err))

		return offer, http.StatusInternalServerError, "please retry your request later", fmt.Errorf("%w", err)
	}

	// Remove the offer from Redis.
	if err = cache.Del(offerID); err != nil {
		var redisErr *redis.Error

		// Not a Redis custom error OR not a cache miss for the key (has already expired and could not be deleted).
		if !errors.As(err, &redisErr) || !redisErr.Is(redis.ErrCacheMiss) {
			logger.Warn("unknown error occurred whilst retrieving Fiat Offer from Redis", zap.Error(err))

			return offer, http.StatusInternalServerError, "please retry your request later", fmt.Errorf("%w", err)
		}
	}

	return offer, http.StatusOK, "", nil
}

// ExchangeTransferFiat will handle an HTTP request to execute and complete a Fiat currency exchange offer.
//
//	@Summary		Transfer Fiat funds between two Fiat currencies using a valid Offer ID.
//	@Description	Transfer Fiat funds between two Fiat currencies. The Offer ID must be valid and have expired.
//	@Tags			fiat currency exchange convert offer transfer execute
//	@Id				exchangeTransferFiat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			offerID	body		models.HTTPFiatTransferRequest	true	"the two currency code and amount to be converted"
//	@Success		200		{object}	models.HTTPSuccess				"a message to confirm the conversion of funds"
//	@Failure		400		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		408		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError				"error message with any available details in payload"
//	@Router			/fiat/exchange/transfer [post]
func ExchangeTransferFiat(
	logger *logger.Logger,
	auth auth.Auth,
	cache redis.Redis,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			err         error
			clientID    uuid.UUID
			request     models.HTTPFiatTransferRequest
			offer       models.HTTPFiatExchangeOfferResponse
			receipt     models.HTTPFiatTransferResponse
			offerID     string
			srcCurrency postgres.Currency
			dstCurrency postgres.Currency
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
				logger.Warn("failed to decrypt Offer ID for Fiat transfer request", zap.Error(err))
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
			if offer, status, msg, err = getCachedOffer(cache, logger, offerID); err != nil {
				ginCtx.AbortWithStatusJSON(status, &models.HTTPError{Message: msg})

				return
			}
		}

		// Verify that the client IDs match.
		if clientID != offer.ClientID {
			logger.Warn("clientID mismatch with Fiat Offer stored in Redis",
				zap.Strings("Requester & Offer Client IDs", []string{clientID.String(), offer.ClientID.String()}))
			ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
				&models.HTTPError{Message: "please retry your request later"})

			return
		}

		// Get currency codes.
		if srcCurrency, dstCurrency, err = validateSourceDestinationAmount(
			offer.SourceAcc, offer.DestinationAcc, offer.Amount); err != nil {
			logger.Warn("failed to extract source and destination currencies from Fiat exchange offer",
				zap.Error(err))
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, &models.HTTPError{Message: err.Error()})

			return
		}

		// Execute exchange.
		srcTxDetails := &postgres.FiatTransactionDetails{
			ClientID: offer.ClientID,
			Currency: srcCurrency,
			Amount:   offer.DebitAmount,
		}
		dstTxDetails := &postgres.FiatTransactionDetails{
			ClientID: offer.ClientID,
			Currency: dstCurrency,
			Amount:   offer.Amount,
		}

		if receipt.SrcTxReceipt, receipt.DstTxReceipt, err = db.
			FiatInternalTransfer(context.Background(), srcTxDetails, dstTxDetails); err != nil {
			logger.Warn("failed to complete internal Fiat transfer", zap.Error(err))
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				&models.HTTPError{Message: "please check you have both currency accounts and enough funds."})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "funds exchange transfer successful", Payload: receipt})
	}
}

// BalanceCurrencyFiat will handle an HTTP request to retrieve a balance for a specific currency.
//
//	@Summary		Retrieve balance for a specific currency.
//	@Description	Retrieves the balance for a specific currency. The currency code must be supplied as a query parameter.
//	@Tags			fiat currency balance
//	@Id				balanceCurrencyFiat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			currencyCode	path		string				true	"the currency code to retrieve the balance for"
//	@Success		200				{object}	models.HTTPSuccess	"the details for a specific currency account"
//	@Failure		400				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		403				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		404				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		500				{object}	models.HTTPError	"error message with any available details in payload"
//	@Router			/fiat/info/balance/{currencyCode} [get]
func BalanceCurrencyFiat(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			accDetails postgres.FiatAccount
			clientID   uuid.UUID
			currency   postgres.Currency
			err        error
		)

		// Extract and validate the currency.
		if err = currency.Scan(ginCtx.Param("currencyCode")); err != nil || !currency.Valid() {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid currency", Payload: ginCtx.Param("currencyCode")})

			return
		}

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		if accDetails, err = db.FiatBalanceCurrency(clientID, currency); err != nil {
			var balanceErr *postgres.Error
			if !errors.As(err, &balanceErr) {
				logger.Info("failed to unpack Fiat account balance currency error", zap.Error(err))
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

// TxDetailsCurrencyFiat will handle an HTTP request to retrieve information on a specific transaction.
//
//	@Summary		Retrieve transaction details for a specific transactionID.
//	@Description	Retrieves the transaction details for a specific transactionID. The transaction ID must be supplied as a query parameter.
//	@Tags			fiat transactionID transaction details
//	@Id				txDetailsCurrencyFiat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			transactionID	path		string				true	"the transaction ID to retrieve the details for"
//	@Success		200				{object}	models.HTTPSuccess	"the transaction details for a specific transaction ID"
//	@Failure		400				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		403				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		404				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		500				{object}	models.HTTPError	"error message with any available details in payload"
//	@Router			/fiat/info/transaction/{transactionID} [get]
func TxDetailsCurrencyFiat(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			journalEntries []postgres.FiatJournal
			clientID       uuid.UUID
			transactionID  uuid.UUID
			err            error
		)

		// Extract and validate the transactionID.
		if transactionID, err = uuid.FromString(ginCtx.Param("transactionID")); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid transaction ID", Payload: ginCtx.Param("transactionID")})

			return
		}

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		if journalEntries, err = db.FiatTxDetailsCurrency(clientID, transactionID); err != nil {
			var balanceErr *postgres.Error
			if !errors.As(err, &balanceErr) {
				logger.Info("failed to unpack Fiat account balance transactionID error", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			ginCtx.AbortWithStatusJSON(balanceErr.Code, models.HTTPError{Message: balanceErr.Message})

			return
		}

		if len(journalEntries) == 0 {
			ginCtx.AbortWithStatusJSON(http.StatusNotFound, models.HTTPError{Message: "transaction id not found"})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "transaction details", Payload: journalEntries})
	}
}

// BalanceCurrencyFiatPaginated will handle an HTTP request to retrieve a balance for all currency accounts held by a
// single client.
//
// If a user request N records, N+1 records will be requested. This is used to calculate if any further records are
// available to for retrieval. The page cursor will be the encrypted N+1'th record to retrieve in the subsequent call.
//
//	@Summary		Retrieve all the currency balances for a specific client.
//	@Description	Retrieves all the currency balances for a specific client. The initial request will only contain (optionally) the page size. Subsequent requests will require a cursors to the next page that will be returned in a previous call to the endpoint. The user may choose to change the page size in any sequence of calls.
//	@Tags			fiat currency balance
//	@Id				balanceCurrencyFiatPaginated
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			pageCursor	query		string				false	"The page cursor into the query results records."
//	@Param			pageSize	query		int					false	"The number of records to retrieve on this page."
//	@Success		200			{object}	models.HTTPSuccess	"a message to with a page of account balances for the client's accounts"
//	@Failure		400			{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		403			{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		404			{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		500			{object}	models.HTTPError	"error message with any available details in payload"
//	@Router			/fiat/info/balance [get]
func BalanceCurrencyFiatPaginated(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			accDetails []postgres.FiatAccount
			clientID   uuid.UUID
			currency   postgres.Currency
			err        error
			nextPage   string
			pageSize   int32
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		// Extract and assemble the page cursor and page size.
		if currency, pageSize, err = utilities.HTTPFiatBalancePaginatedRequest(
			auth,
			ginCtx.Query("pageCursor"),
			ginCtx.Query("pageSize")); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: "invalid page cursor or page size"})

			return
		}

		if accDetails, err = db.FiatBalanceCurrencyPaginated(clientID, currency, pageSize+1); err != nil {
			var balanceErr *postgres.Error
			if !errors.As(err, &balanceErr) {
				logger.Info("failed to unpack Fiat account balance currency error", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			ginCtx.AbortWithStatusJSON(balanceErr.Code, models.HTTPError{Message: balanceErr.Message})

			return
		}

		// Generate the next page link by pulling the last item returned if the page size is N + 1 of the requested.
		lastRecordIdx := int(pageSize)
		if len(accDetails) == lastRecordIdx+1 {
			// Generate next page link.
			if nextPage, err = auth.EncryptToString([]byte(accDetails[pageSize].Currency)); err != nil {
				logger.Error("failed to encrypt Fiat currency for use as cursor", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			// Remove last element.
			accDetails = accDetails[:pageSize]

			// Generate naked next page link.
			nextPage = fmt.Sprintf(constants.GetNextPageRESTFormatString(), nextPage, pageSize)
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "account balances",
			Payload: models.HTTPFiatDetailsPaginated{
				AccountBalances: accDetails,
				Links: models.HTTPLinks{
					NextPage: nextPage,
				},
			}})
	}
}

// TxDetailsCurrencyFiatPaginated will handle an HTTP request to retrieve all transaction details for a currency account
// held by a single client for a given month.
//
// If a user request N records, N+1 records will be requested. This is used to calculate if any further records are
// available to for retrieval. The page cursor will be the encrypted date range for the month as well as the offset.
//
//	@Summary		Retrieve all the transactions for a currency account for a specific client during a specified month.
//	@Description	Retrieves all the transaction details for currency a specific client during the specified month. The initial request will contain (optionally) the page size and, month, year, and timezone (option, defaults to UTC). Subsequent requests will require a cursors to the next page that will be returned in the previous call to the endpoint. The user may choose to change the page size in any sequence of calls.
//	@Tags			fiat currency transaction
//	@Id				txDetailsCurrencyFiatPaginated
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			currencyCode	path		string				true	"the currency code to retrieve the transaction details for."
//	@Param			pageCursor		query		string				false	"The page cursor into the query results records."
//	@Param			timezone		query		string				false	"The timezone for the month in question."
//	@Param			month			query		int					false	"The month for which transaction records are being requested."
//	@Param			year			query		int					false	"The year for the month for which transaction records are being requested."
//	@Param			pageSize		query		int					false	"The number of records to retrieve on this page."
//	@Success		200				{object}	models.HTTPSuccess	"a message to confirm the conversion of funds"
//	@Failure		400				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		403				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		404				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		416				{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		500				{object}	models.HTTPError	"error message with any available details in payload"
//	@Router			/fiat/info/transaction/all/{currencyCode}/ [get]
func TxDetailsCurrencyFiatPaginated(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			journalEntries []postgres.FiatJournal
			clientID       uuid.UUID
			currency       postgres.Currency
			err            error

			params = utilities.HTTPFiatPaginatedTxParams{
				PageSizeStr:   ginCtx.Query("pageSize"),
				PageCursorStr: ginCtx.Query("pageCursor"),
				TimezoneStr:   ginCtx.Query("timezone"),
				MonthStr:      ginCtx.Query("month"),
				YearStr:       ginCtx.Query("year"),
			}
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		// Extract and validate the currency.
		if err = currency.Scan(ginCtx.Param("currencyCode")); err != nil || !currency.Valid() {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: "invalid currency", Payload: ginCtx.Param("currencyCode")})

			return
		}

		// Check for required parameters.
		if len(params.PageCursorStr) == 0 && (len(params.MonthStr) == 0 || len(params.YearStr) == 0) {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: "missing required parameters"})

			return
		}

		// Decrypt values from page cursor, if present. Otherwise, prepare values using query strings.
		httpCode, err := utilities.HTTPFiatTxParseQueryParams(auth, logger, &params)
		if err != nil {
			ginCtx.AbortWithStatusJSON(httpCode, models.HTTPError{Message: err.Error()})

			return
		}

		// Retrieve transaction details page.
		if journalEntries, err = db.FiatTransactionsCurrencyPaginated(
			clientID, currency, params.PageSize+1, params.Offset, params.PeriodStart, params.PeriodEnd); err != nil {
			var balanceErr *postgres.Error
			if !errors.As(err, &balanceErr) {
				logger.Info("failed to unpack Fiat transactions request error", zap.Error(err))
				ginCtx.AbortWithStatusJSON(http.StatusInternalServerError,
					models.HTTPError{Message: "please retry your request later"})

				return
			}

			ginCtx.AbortWithStatusJSON(balanceErr.Code, models.HTTPError{Message: balanceErr.Message})

			return
		}

		if len(journalEntries) == 0 {
			ginCtx.AbortWithStatusJSON(http.StatusRequestedRangeNotSatisfiable,
				models.HTTPError{Message: "no transactions"})

			return
		}

		// Generate naked next page link.
		params.NextPage = fmt.Sprintf(constants.GetNextPageRESTFormatString(), params.NextPage, params.PageSize)

		// Check if there are further pages of data. If not, set the next link to be empty.
		if len(journalEntries) > int(params.PageSize) {
			journalEntries = journalEntries[:int(params.PageSize)]
		} else {
			params.NextPage = ""
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "account transactions",
			Payload: models.HTTPFiatTransactionsPaginated{
				TransactionDetails: journalEntries,
				Links: models.HTTPLinks{
					NextPage: params.NextPage,
				},
			}})
	}
}
