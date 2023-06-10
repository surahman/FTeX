package common

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	"go.uber.org/zap"
)

// HTTPCryptoOpen opens a Crypto account.
func HTTPCryptoOpen(db postgres.Postgres, logger *logger.Logger, clientID uuid.UUID, ticker string) (
	int, string, error) {
	var (
		err error
	)

	if err = db.CryptoCreateAccount(clientID, ticker); err != nil {
		var createErr *postgres.Error
		if !errors.As(err, &createErr) {
			logger.Info("failed to unpack open Crypto account error", zap.Error(err))

			return http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
		}

		return createErr.Code, createErr.Message, fmt.Errorf("%w", err)
	}

	return 0, "", nil
}

// HTTPCryptoBalance retrieves a balance for a specific Crypto account.
func HTTPCryptoBalance(db postgres.Postgres, logger *logger.Logger, clientID uuid.UUID, ticker string) (
	*postgres.CryptoAccount, int, string, any, error) {
	var (
		accDetails postgres.CryptoAccount
		err        error
	)

	// Extract and validate the currency.
	if len(ticker) < 1 || len(ticker) > 6 {
		return nil, http.StatusBadRequest, constants.GetInvalidCurrencyString(), ticker,
			errors.New(constants.GetInvalidCurrencyString())
	}

	if accDetails, err = db.CryptoBalance(clientID, ticker); err != nil {
		var balanceErr *postgres.Error
		if !errors.As(err, &balanceErr) {
			logger.Info("failed to unpack Crypto account balance currency error", zap.Error(err))

			return nil, http.StatusInternalServerError, retryMessage, nil, fmt.Errorf("%w", err)
		}

		return nil, balanceErr.Code, balanceErr.Message, nil, fmt.Errorf("%w", err)
	}

	return &accDetails, 0, "", nil, nil
}

// HTTPCryptoOffer will request the conversion rate, prepare the price quote, and store it in the Redis cache.
func HTTPCryptoOffer(auth auth.Auth, cache redis.Redis, logger *logger.Logger, quotes quotes.Quotes,
	clientID uuid.UUID, source, destination string, sourceAmount decimal.Decimal, isPurchase bool) (
	models.HTTPExchangeOfferResponse, int, string, error) {
	var (
		err          error
		offer        models.HTTPExchangeOfferResponse
		offerID      = xid.New().String()
		precision    = constants.GetDecimalPlacesFiat()
		fiatCurrency = source
	)

	// Configure precision and fiat tickers for Crypto sale.
	if !isPurchase {
		precision = constants.GetDecimalPlacesCrypto()
		fiatCurrency = destination
	}

	// Validate the Fiat currency and source amount.
	if _, err = HTTPValidateOfferRequest(sourceAmount, precision, fiatCurrency); err != nil {
		return offer, http.StatusBadRequest, constants.GetInvalidRequest(), fmt.Errorf("%w", err)
	}

	// Compile exchange rate offer.
	if offer.Rate, offer.Amount, err = quotes.CryptoConversion(
		source, destination, sourceAmount, isPurchase, nil); err != nil {
		logger.Warn("failed to retrieve quote for Cryptocurrency purchase/sale offer", zap.Error(err))

		return offer, http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
	}

	// Check to make sure there is a valid Cryptocurrency amount.
	if !offer.Amount.GreaterThan(decimal.NewFromFloat(0)) {
		msg := "cryptocurrency purchase/sale amount is too small"

		return offer, http.StatusBadRequest, msg, errors.New(msg)
	}

	offer.PriceQuote.ClientID = clientID
	offer.SourceAcc = source
	offer.DestinationAcc = destination
	offer.DebitAmount = sourceAmount
	offer.Expires = time.Now().Add(constants.GetFiatOfferTTL()).Unix()
	offer.IsCryptoPurchase = isPurchase
	offer.IsCryptoSale = !isPurchase

	// Encrypt offer ID before returning to client.
	if offer.OfferID, err = auth.EncryptToString([]byte(offerID)); err != nil {
		msg := "failed to encrypt offer ID for Cryptocurrency purchase/sale offer"
		logger.Warn(msg, zap.Error(err))

		return offer, http.StatusInternalServerError, retryMessage, errors.New(msg)
	}

	// Store the offer in Redis.
	if err = cache.Set(offerID, &offer, constants.GetFiatOfferTTL()); err != nil {
		msg := "failed to store Cryptocurrency purchase/sale offer in cache"
		logger.Warn(msg, zap.Error(err))

		return offer, http.StatusInternalServerError, retryMessage, errors.New(msg)
	}

	return offer, 0, "", nil
}

// HTTPExchangeCrypto will complete a Cryptocurrency exchange.
func HTTPExchangeCrypto(auth auth.Auth, cache redis.Redis, db postgres.Postgres, logger *logger.Logger,
	clientID uuid.UUID, offerID string) (models.HTTPCryptoTransferResponse, int, string, error) {
	var (
		err          error
		offer        models.HTTPExchangeOfferResponse
		receipt      models.HTTPCryptoTransferResponse
		cryptoTicker string
		fiatTicker   string
		cryptoAmount decimal.Decimal
		fiatAmount   decimal.Decimal
		precision    = constants.GetDecimalPlacesCrypto()
		transferFunc = db.CryptoPurchase
		fiatCurrency []postgres.Currency
	)

	// Extract Offer ID from request.
	{
		var rawOfferID []byte
		if rawOfferID, err = auth.DecryptFromString(offerID); err != nil {
			logger.Warn("failed to decrypt Offer ID for Crypto transfer request", zap.Error(err))

			return receipt, http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
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
		if offer, status, msg, err = HTTPGetCachedOffer(cache, logger, offerID); err != nil {
			return receipt, status, msg, fmt.Errorf("%w", err)
		}
	}

	// Verify that offer is a Crypto offer.
	if !(offer.IsCryptoSale || offer.IsCryptoPurchase) {
		msg := "invalid Cryptocurrency exchange offer"

		return receipt, http.StatusBadRequest, msg, errors.New(msg)
	}

	// Verify that the client IDs match.
	if clientID != offer.ClientID {
		msg := "clientID mismatch with the Crypto Offer stored in Redis"
		logger.Warn(msg,
			zap.Strings("Requester & Offer Client IDs", []string{clientID.String(), offer.ClientID.String()}))

		return receipt, http.StatusInternalServerError, retryMessage, errors.New(msg)
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
	if fiatCurrency, err = HTTPValidateOfferRequest(
		offer.Amount, precision, fiatTicker); err != nil {
		msg := "failed to extract Fiat currency from Crypto exchange offer"
		logger.Warn(msg, zap.Error(err))

		return receipt, http.StatusBadRequest, msg, fmt.Errorf("%w", err)
	}

	// Execute transfer.
	if receipt.FiatTxReceipt, receipt.CryptoTxReceipt, err =
		transferFunc(clientID, fiatCurrency[0], fiatAmount, cryptoTicker, cryptoAmount); err != nil {
		return receipt, http.StatusInternalServerError, err.Error(), fmt.Errorf("%w", err)
	}

	return receipt, 0, "", nil
}

// cryptoBalancePaginatedRequest will convert the encrypted URL query parameter for the ticker and the record
// limit and covert them to a string and integer record limit. The tickerStr is the encrypted pageCursor passed in.
func cryptoBalancePaginatedRequest(auth auth.Auth, tickerStr, limitStr string) (string, int32, error) {
	var (
		ticker    string
		decrypted []byte
		err       error
		limit     int64
	)

	// Decrypt currency ticker string.
	decrypted = []byte("BTC")

	if len(tickerStr) > 0 {
		if decrypted, err = auth.DecryptFromString(tickerStr); err != nil {
			return ticker, -1, fmt.Errorf("failed to decrypt next ticker")
		}
	}

	ticker = string(decrypted)

	// Convert record limit to int and set base bound for bad input.
	if len(limitStr) > 0 {
		if limit, err = strconv.ParseInt(limitStr, 10, 32); err != nil {
			return ticker, -1, fmt.Errorf("failed to parse record limit")
		}
	}

	if limit < 1 {
		limit = 10
	}

	return ticker, int32(limit), nil
}

// HTTPCryptoBalancePaginated retrieves a page of data from the Cryptocurrency account balances and prepares a link to
// the next page of data.
func HTTPCryptoBalancePaginated(auth auth.Auth, db postgres.Postgres, logger *logger.Logger, clientID uuid.UUID,
	pageCursor, pageSizeStr string, isREST bool) (models.HTTPCryptoDetailsPaginated, int, string, error) {
	var (
		ticker        string
		err           error
		pageSize      int32
		nextPage      string
		cryptoDetails models.HTTPCryptoDetailsPaginated
	)

	// Extract and assemble the page cursor and page size.
	if ticker, pageSize, err = cryptoBalancePaginatedRequest(auth, pageCursor, pageSizeStr); err != nil {
		return cryptoDetails, http.StatusBadRequest, "invalid page cursor or page size", fmt.Errorf("%w", err)
	}

	if cryptoDetails.AccountBalances, err = db.CryptoBalancesPaginated(clientID, ticker, pageSize+1); err != nil {
		var balanceErr *postgres.Error
		if !errors.As(err, &balanceErr) {
			logger.Info("failed to unpack Fiat account balance currency error", zap.Error(err))

			return cryptoDetails, http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
		}

		return cryptoDetails, balanceErr.Code, balanceErr.Message, fmt.Errorf("%w", err)
	}

	// Generate the next page link by pulling the last item returned if the page size is N + 1 of the requested.
	lastRecordIdx := int(pageSize)
	if len(cryptoDetails.AccountBalances) == lastRecordIdx+1 {
		// Generate next page link.
		if nextPage, err =
			auth.EncryptToString([]byte(cryptoDetails.AccountBalances[pageSize].Ticker)); err != nil {
			logger.Error("failed to encrypt Fiat currency for use as cursor", zap.Error(err))

			return cryptoDetails, http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
		}

		// Remove last element.
		cryptoDetails.AccountBalances = cryptoDetails.AccountBalances[:pageSize]

		// Generate naked next page link for REST.
		if isREST {
			cryptoDetails.Links.NextPage = fmt.Sprintf(constants.GetNextPageRESTFormatString(), nextPage, pageSize)
		} else {
			cryptoDetails.Links.PageCursor = nextPage
		}
	}

	return cryptoDetails, 0, "", nil
}

// HTTPCryptoTXPaginated will retrieve a page of Cryptocurrency transactions based on a time range.
func HTTPCryptoTXPaginated(auth auth.Auth, db postgres.Postgres, logger *logger.Logger, params *HTTPPaginatedTxParams,
	clientID uuid.UUID, ticker string) (models.HTTPCryptoTransactionsPaginated, int, string, error) {
	var (
		err          error
		httpCode     int
		monthStrLen  = len(params.MonthStr)
		transactions models.HTTPCryptoTransactionsPaginated
	)

	// Check for required parameters.
	if len(params.PageCursorStr) == 0 && (monthStrLen < 1 || monthStrLen > 2 || len(params.YearStr) != 4) {
		msg := "missing required parameters"

		return transactions, http.StatusBadRequest, msg, errors.New(msg)
	}

	// Decrypt values from page cursor, if present. Otherwise, prepare values using query strings.
	if httpCode, err = HTTPTxParseQueryParams(auth, logger, params); err != nil {
		return transactions, httpCode, err.Error(), fmt.Errorf("%w", err)
	}

	// Retrieve transaction details page.
	if transactions.TransactionDetails, err = db.CryptoTransactionsPaginated(
		clientID, ticker, params.PageSize+1, params.Offset, params.PeriodStart, params.PeriodEnd); err != nil {
		var balanceErr *postgres.Error
		if !errors.As(err, &balanceErr) {
			logger.Info("failed to unpack transactions request error", zap.Error(err))

			return transactions, http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
		}

		return transactions, balanceErr.Code, balanceErr.Message, fmt.Errorf("%w", err)
	}

	if len(transactions.TransactionDetails) == 0 {
		msg := "no transactions"

		return transactions, http.StatusRequestedRangeNotSatisfiable, msg, errors.New(msg)
	}

	// Generate naked next page link. The params will have had the nextPage link generated in the prior methods called.
	transactions.Links.NextPage = fmt.Sprintf(constants.GetNextPageRESTFormatString(), params.NextPage, params.PageSize)

	// Check if there are further pages of data. If not, set the next link to be empty.
	if len(transactions.TransactionDetails) > int(params.PageSize) {
		transactions.TransactionDetails = transactions.TransactionDetails[:int(params.PageSize)]
	} else {
		transactions.Links.NextPage = ""
	}

	return transactions, 0, "", nil
}
