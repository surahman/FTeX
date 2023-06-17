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
//
//nolint:dupl
func OpenFiat(logger *logger.Logger, auth auth.Auth, db postgres.Postgres, authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID    uuid.UUID
			err         error
			request     models.HTTPOpenCurrencyAccountRequest
			httpStatus  int
			httpMessage string
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

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

		if httpStatus, httpMessage, err = common.HTTPFiatOpen(db, logger, clientID, request.Currency); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage, Payload: request.Currency})

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
			err             error
			httpMessage     string
			httpStatus      int
			payload         any
			request         models.HTTPDepositCurrencyRequest
			transferReceipt *postgres.FiatAccountTransferResult
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if transferReceipt, httpStatus, httpMessage, payload, err =
			common.HTTPFiatDeposit(db, logger, clientID, &request); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage, Payload: payload})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "funds successfully transferred", Payload: *transferReceipt})
	}
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
//	@Param			request	body		models.HTTPExchangeOfferRequest	true	"the two currency code and amount to be converted"
//	@Success		200		{object}	models.HTTPSuccess				"a message to confirm the conversion rate for a currency"
//	@Failure		400		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError				"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError				"error message with any available details in payload"
//	@Router			/fiat/exchange/offer [post]
func ExchangeOfferFiat(
	logger *logger.Logger,
	auth auth.Auth,
	cache redis.Redis,
	quotes quotes.Quotes,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			clientID    uuid.UUID
			err         error
			httpStatus  int
			httpMessage string
			payload     any
			request     models.HTTPExchangeOfferRequest
			offer       *models.HTTPExchangeOfferResponse
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if offer, httpStatus, httpMessage, payload, err =
			common.HTTPFiatOffer(auth, cache, logger, quotes, clientID, &request); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage, Payload: payload})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "conversion rate offer", Payload: offer})
	}
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
//	@Param			offerID	body		models.HTTPTransferRequest	true	"the two currency codes and amount to be converted"
//	@Success		200		{object}	models.HTTPSuccess			"a message to confirm the conversion of funds"
//	@Failure		400		{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		408		{object}	models.HTTPError			"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError			"error message with any available details in payload"
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
			receipt     *models.HTTPFiatTransferResponse
			request     models.HTTPTransferRequest
			httpStatus  int
			httpMessage string
			payload     any
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, &models.HTTPError{Message: err.Error()})

			return
		}

		if err = ginCtx.ShouldBindJSON(&request); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusBadRequest, models.HTTPError{Message: err.Error()})

			return
		}

		if receipt, httpStatus, httpMessage, payload, err =
			common.HTTPFiatTransfer(auth, cache, db, logger, clientID, &request); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage, Payload: payload})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "funds exchange transfer successful", Payload: receipt})
	}
}

// BalanceFiat will handle an HTTP request to retrieve a balance for a specific Fiat currency.
//
//	@Summary		Retrieve balance for a specific Fiat currency.
//	@Description	Retrieves the balance for a specific Fiat currency. The currency ticker must be supplied as a query parameter.
//	@Tags			fiat currency balance
//	@Id				balanceCurrencyFiat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			ticker	path		string				true	"the currency ticker to retrieve the balance for"
//	@Success		200		{object}	models.HTTPSuccess	"the details for a specific currency account"
//	@Failure		400		{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		403		{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		404		{object}	models.HTTPError	"error message with any available details in payload"
//	@Failure		500		{object}	models.HTTPError	"error message with any available details in payload"
//	@Router			/fiat/info/balance/{ticker} [get]
func BalanceFiat( //nolint:dupl
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			accDetails  *postgres.FiatAccount
			clientID    uuid.UUID
			err         error
			httpStatus  int
			httpMessage string
			payload     any
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		if accDetails, httpStatus, httpMessage, payload, err =
			common.HTTPFiatBalance(db, logger, clientID, ginCtx.Param("ticker")); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage, Payload: payload})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "account balance", Payload: accDetails})
	}
}

// TxDetailsFiat will handle an HTTP request to retrieve information on a specific transaction.
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
func TxDetailsFiat(
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

// BalanceFiatPaginated will handle an HTTP request to retrieve a balance for all currency accounts held by a single
// client.
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
func BalanceFiatPaginated(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			accDetails  *models.HTTPFiatDetailsPaginated
			clientID    uuid.UUID
			err         error
			httpStatus  int
			httpMessage string
		)

		if clientID, _, err = auth.ValidateJWT(ginCtx.GetHeader(authHeaderKey)); err != nil {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, models.HTTPError{Message: err.Error()})

			return
		}

		if accDetails, httpStatus, httpMessage, err = common.HTTPFiatBalancePaginated(auth, db, logger,
			clientID, ginCtx.Query("pageCursor"), ginCtx.Query("pageSize"), true); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "account balances", Payload: accDetails})
	}
}

// TxDetailsFiatPaginated will handle an HTTP request to retrieve all transaction details for a currency account held by
// a single client for a given month.
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
func TxDetailsFiatPaginated(
	logger *logger.Logger,
	auth auth.Auth,
	db postgres.Postgres,
	authHeaderKey string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		var (
			journalEntries *models.HTTPFiatTransactionsPaginated
			clientID       uuid.UUID
			err            error
			httpStatus     int
			httpMessage    string
			payload        any

			params = common.HTTPPaginatedTxParams{
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

		if journalEntries, httpStatus, httpMessage, payload, err = common.HTTPFiatTransactionsPaginated(auth, db,
			logger, clientID, ginCtx.Param("currencyCode"), &params, true); err != nil {
			ginCtx.AbortWithStatusJSON(httpStatus, models.HTTPError{Message: httpMessage, Payload: payload})

			return
		}

		ginCtx.JSON(http.StatusOK, models.HTTPSuccess{Message: "account transactions", Payload: journalEntries})
	}
}
