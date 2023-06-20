package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/common"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
	"github.com/surahman/FTeX/pkg/validator"
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
//
//nolint:dupl
func OpenCrypto(logger *logger.Logger, auth auth.Auth, db postgres.Postgres) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID    uuid.UUID
			err         error
			request     models.HTTPOpenCurrencyAccountRequest
			httpStatus  int
			httpMessage string
		)

		if clientID, _, err = auth.TokenInfoFromGinCtx(ginCtx); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "malformed authentication token"})

			return
		}

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: constants.ValidationString(), Payload: err})

			return
		}

		if httpStatus, httpMessage, err = common.HTTPCryptoOpen(db, logger, clientID, request.Currency); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage, Payload: request.Currency})

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
func OfferCrypto(logger *logger.Logger, auth auth.Auth, cache redis.Redis, quotes quotes.Quotes) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID      uuid.UUID
			err           error
			request       models.HTTPCryptoOfferRequest
			offer         models.HTTPExchangeOfferResponse
			status        int
			statusMessage string
		)

		if clientID, _, err = auth.TokenInfoFromGinCtx(ginCtx); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "malformed authentication token"})

			return
		}

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: constants.ValidationString(), Payload: err})

			return
		}

		offer, status, statusMessage, err = common.HTTPCryptoOffer(auth, cache, logger, quotes,
			clientID, request.SourceCurrency, request.DestinationCurrency, request.SourceAmount, *request.IsPurchase)
		if err != nil {
			httpErr := &models.HTTPError{Message: statusMessage}
			if statusMessage == constants.InvalidRequestString() {
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
func ExchangeCrypto(logger *logger.Logger, auth auth.Auth, cache redis.Redis, db postgres.Postgres) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			err      error
			clientID uuid.UUID
			request  models.HTTPTransferRequest
		)

		if clientID, _, err = auth.TokenInfoFromGinCtx(ginCtx); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "malformed authentication token"})

			return
		}

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if err = validator.ValidateStruct(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest,
				models.HTTPError{Message: constants.ValidationString(), Payload: err})

			return
		}

		receipt, status, httpErrMsg, err := common.HTTPExchangeCrypto(auth, cache, db, logger, clientID, request.OfferID)
		if err != nil {
			ginCtx.AbortWithStatusJSON(status, &models.HTTPError{Message: httpErrMsg})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "funds exchange transfer successful", Payload: receipt})
	}
}

// BalanceCrypto will handle an HTTP request to retrieve a balance for a specific Cryptocurrency.
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
//
//nolint:dupl
func BalanceCrypto(logger *logger.Logger, auth auth.Auth, db postgres.Postgres) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			accDetails  *postgres.CryptoAccount
			clientID    uuid.UUID
			err         error
			httpStatus  int
			httpMessage string
			payload     any
		)

		if clientID, _, err = auth.TokenInfoFromGinCtx(ginCtx); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: "malformed authentication token"})

			return
		}

		if accDetails, httpStatus, httpMessage, payload, err =
			common.HTTPCryptoBalance(db, logger, clientID, ginCtx.Param("ticker")); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage, Payload: payload})

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
		journalEntries, status, errMsg, err := common.HTTPTxDetails(db, logger, clientID, transactionID)
		if err != nil {
			ginCtx.AbortWithStatusJSON(status, models.HTTPError{Message: errMsg})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "transaction details", Payload: journalEntries})
	}
}

// BalanceCryptoPaginated will handle an HTTP request to retrieve a balance for all Cryptocurrency accounts held by a
// single client.
//
// If a user requests N records, N+1 records will be requested. This is used to calculate if any further records are
// available for retrieval. The page cursor will be the encrypted N+1'th record to retrieve in the subsequent call.
//
//	@Summary		Retrieve all the Cryptocurrency balances for a specific client.
//	@Description	Retrieves all the Cryptocurrency balances for a specific client. The initial request will only contain (optionally) the page size. Subsequent requests will require a cursors to the next page that will be returned in a previous call to the endpoint. The user may choose to change the page size in any sequence of calls.
//	@Tags			crypto cryptocurrency currency balance
//	@Id				balanceCurrencyCryptoPaginated
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
//	@Router			/crypto/info/balance [get]
func BalanceCryptoPaginated(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			accDetails  models.HTTPCryptoDetailsPaginated
			httpStatus  int
			httpMessage string
			clientID    uuid.UUID
			err         error
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		accDetails, httpStatus, httpMessage, err = common.HTTPCryptoBalancePaginated(auth, db, logger,
			clientID, ginCtx.Query("pageCursor"), ginCtx.Query("pageSize"), true)
		if err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "account balances", Payload: accDetails})
	}
}

// TxDetailsCryptoPaginated will handle an HTTP request to retrieve all transaction details for a currency account
// held by a single client for a given month.
//
// If a user requests N records, N+1 records will be requested. This is used to calculate if any further records are
// available for retrieval. The page cursor will be the encrypted date range for the month as well as the offset.
//
//	@Summary		Retrieve all the transactions for a currency account for a specific client during a specified month.
//	@Description	Retrieves all the transaction details for currency a specific client during the specified month. The initial request will contain (optionally) the page size and, month, year, and timezone (option, defaults to UTC). Subsequent requests will require a cursors to the next page that will be returned in the previous call to the endpoint. The user may choose to change the page size in any sequence of calls.
//	@Tags			crypto cryptocurrency currency transaction
//	@Id				txDetailsCryptoPaginated
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			ticker		path		string				true	"the currency ticker to retrieve the transaction details for."
//	@Param			pageCursor	query		string				false	"The page cursor into the query results records."
//	@Param			timezone	query		string				false	"The timezone for the month in question."
//	@Param			month		query		int					false	"The month for which transaction records are being requested."
//	@Param			year		query		int					false	"The year for the month for which transaction records are being requested."
//	@Param			pageSize	query		int					false	"The number of records to retrieve on this page."
//	@Success		200			{object}	models.HTTPSuccess	"a message to confirm the conversion of funds"
//	@Failure		400			{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		403			{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		404			{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		416			{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		500			{object}	models.HTTPError	"error message with any available details in payload"
//	@Router			/crypto/info/transaction/all/{ticker}/ [get]
func TxDetailsCryptoPaginated(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID       uuid.UUID
			ticker         = ginCtx.Param("ticker")
			err            error
			journalEntries models.HTTPCryptoTransactionsPaginated
			httpStatus     int
			httpMessage    string
			params         = common.HTTPPaginatedTxParams{
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

		if journalEntries, httpStatus, httpMessage, err =
			common.HTTPCryptoTransactionsPaginated(auth, db, logger, &params, clientID, ticker, false); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "account transactions", Payload: journalEntries})
	}
}
