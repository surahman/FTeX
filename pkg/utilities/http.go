package utilities

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

const retryMessage = "please retry your request later"

// HTTPGetCachedOffer will retrieve and then evict an offer from the Redis cache.
func HTTPGetCachedOffer(cache redis.Redis, logger *logger.Logger, offerID string) (
	models.HTTPExchangeOfferResponse, int, string, error) {
	var (
		err   error
		offer models.HTTPExchangeOfferResponse
	)

	// Retrieve the offer from Redis.
	if err = cache.Get(offerID, &offer); err != nil {
		var redisErr *redis.Error

		// If we have a valid Redis package error AND the error is that the key is not found.
		if errors.As(err, &redisErr) && redisErr.Is(redis.ErrCacheMiss) {
			return offer, http.StatusRequestTimeout, "Fiat exchange rate offer has expired", fmt.Errorf("%w", err)
		}

		logger.Warn("unknown error occurred whilst retrieving Fiat Offer from Redis", zap.Error(err))

		return offer, http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
	}

	// Remove the offer from Redis.
	if err = cache.Del(offerID); err != nil {
		var redisErr *redis.Error

		// Not a Redis custom error OR not a cache miss for the key (has already expired and could not be deleted).
		if !errors.As(err, &redisErr) || !redisErr.Is(redis.ErrCacheMiss) {
			logger.Warn("unknown error occurred whilst retrieving Fiat Offer from Redis", zap.Error(err))

			return offer, http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
		}
	}

	return offer, http.StatusOK, "", nil
}

// HTTPFiatBalancePaginatedRequest will convert the encrypted URL query parameter for the currency and the record limit
// and covert them to a currency and integer record limit. The currencyStr is the encrypted pageCursor passed in.
func HTTPFiatBalancePaginatedRequest(auth auth.Auth, currencyStr, limitStr string) (postgres.Currency, int32, error) {
	var (
		currency  postgres.Currency
		decrypted []byte
		err       error
		limit     int64
	)

	// Decrypt currency string and covert to currency struct.
	decrypted = []byte("AED")

	if len(currencyStr) > 0 {
		if decrypted, err = auth.DecryptFromString(currencyStr); err != nil {
			return currency, -1, fmt.Errorf("failed to decrypt next currency")
		}
	}

	if err = currency.Scan(string(decrypted)); err != nil {
		return currency, -1, fmt.Errorf("failed to parse currency")
	}

	// Convert record limit to int and set base bound for bad input.
	if len(limitStr) > 0 {
		if limit, err = strconv.ParseInt(limitStr, 10, 32); err != nil {
			return currency, -1, fmt.Errorf("failed to parse record limit")
		}
	}

	if limit < 1 {
		limit = 10
	}

	return currency, int32(limit), nil
}

// HTTPFiatTransactionInfoPaginatedRequest will generate the month bounds and record limits using supplied query
// parameters.
func HTTPFiatTransactionInfoPaginatedRequest(auth auth.Auth, monthStr, yearStr, timezoneStr string, pageSize int32) (
	pgtype.Timestamptz, pgtype.Timestamptz, string, error) {
	var (
		startYear      int64
		startMonth     int64
		endYear        int64
		endMonth       int64
		startTime      time.Time
		endTime        time.Time
		pageCursor     string
		periodStartStr string
		periodEndStr   string
		periodStart    pgtype.Timestamptz
		periodEnd      pgtype.Timestamptz
		err            error
	)

	// Generate timestamps.

	// Configure empty timezone to Zulu/UTC.
	if len(timezoneStr) == 0 {
		timezoneStr = "+00:00"
	}

	// Extract year and month.
	if startYear, err = strconv.ParseInt(yearStr, 10, 32); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("invalid year")
	}

	if startMonth, err = strconv.ParseInt(monthStr, 10, 32); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("invalid month")
	}

	// Setup end year and month.
	endYear = startYear
	endMonth = startMonth + 1

	if endMonth == 13 { //nolint:gomnd
		endMonth = 1
		endYear++
	}

	// Prepare Postgres timestamps.
	periodStartStr = fmt.Sprintf(constants.GetMonthFormatString(), startYear, startMonth, timezoneStr)
	if startTime, err = time.Parse(time.RFC3339, periodStartStr); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("start date parse failure %w", err)
	}

	if err = periodStart.Scan(startTime); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("invalid start date %w", err)
	}

	periodEndStr = fmt.Sprintf(constants.GetMonthFormatString(), endYear, endMonth, timezoneStr)
	if endTime, err = time.Parse(time.RFC3339, periodEndStr); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("end date parse failure %w", err)
	}

	if err = periodEnd.Scan(endTime); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("end start date %w", err)
	}

	// Prepare page cursor.
	if pageCursor, err = HTTPFiatTransactionGeneratePageCursor(auth, periodStartStr, periodEndStr, pageSize); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("failed to encrypt page cursor %w", err)
	}

	return periodStart, periodEnd, pageCursor, nil
}

// HTTPFiatTransactionGeneratePageCursor will generate the encrypted page cursor.
//
//nolint:wrapcheck
func HTTPFiatTransactionGeneratePageCursor(auth auth.Auth, periodStartStr, periodEndStr string, offset int32) (
	string, error) {
	return auth.EncryptToString([]byte(fmt.Sprintf("%s,%s,%d", periodStartStr, periodEndStr, offset)))
}

// HTTPFiatTransactionUnpackPageCursor will unpack an encrypted page cursor to its component parts.
func HTTPFiatTransactionUnpackPageCursor(auth auth.Auth, pageCursor string) (
	pgtype.Timestamptz, string, pgtype.Timestamptz, string, int32, error) {
	var (
		startPGTS pgtype.Timestamptz
		endPGTS   pgtype.Timestamptz
		buffer    []byte
		err       error
	)

	if buffer, err = auth.DecryptFromString(pageCursor); err != nil {
		return startPGTS, "", endPGTS, "", -1, fmt.Errorf("failed to decrypt page cursor %w", err)
	}

	components := strings.Split(string(buffer), ",")
	if len(components) != 3 { //nolint:gomnd
		return startPGTS, "", endPGTS, "", -1, fmt.Errorf("decrypted page curror is invalid")
	}

	offset, err := strconv.ParseInt(components[2], 10, 32)
	if err != nil {
		return startPGTS, "", endPGTS, "", -1, fmt.Errorf("failed to parse offset %w", err)
	}

	// Prepare Postgres timestamps.
	startTime, err := time.Parse(time.RFC3339, components[0])
	if err != nil {
		return startPGTS, "", endPGTS, "", -1, fmt.Errorf("start date parse failure %w", err)
	}

	if err = startPGTS.Scan(startTime); err != nil {
		return startPGTS, "", endPGTS, "", -1, fmt.Errorf("invalid start date %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, components[1])
	if err != nil {
		return startPGTS, "", endPGTS, "", -1, fmt.Errorf("end date parse failure %w", err)
	}

	if err = endPGTS.Scan(endTime); err != nil {
		return startPGTS, "", endPGTS, "", -1, fmt.Errorf("end start date %w", err)
	}

	return startPGTS, components[0], endPGTS, components[1], int32(offset), nil
}

// HTTPFiatPaginatedTxParams contains the HTTP request as well as the database query parameters.
type HTTPFiatPaginatedTxParams struct {
	// HTTP request input parameters.
	PageSizeStr   string
	PageCursorStr string
	TimezoneStr   string
	MonthStr      string
	YearStr       string

	// Postgres query parameters.
	Offset      int32
	PageSize    int32
	NextPage    string
	PeriodStart pgtype.Timestamptz
	PeriodEnd   pgtype.Timestamptz
}

// HTTPFiatTxParseQueryParams will parse the HTTP request input parameters in database query parameters for the
// paginated Fiat transactions requests.
func HTTPFiatTxParseQueryParams(auth auth.Auth, logger *logger.Logger, params *HTTPFiatPaginatedTxParams) (int, error) {
	var (
		err            error
		periodStartStr string
		periodEndStr   string
	)

	// Prepare page size.
	if len(params.PageSizeStr) > 0 {
		var pageSize int64

		if pageSize, err = strconv.ParseInt(params.PageSizeStr, 10, 32); err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid page size")
		}

		params.PageSize = int32(pageSize)
	}

	if params.PageSize < 1 {
		params.PageSize = 10
	}

	// Decrypt values from page cursor, if present. Otherwise, prepare values using query strings.
	if len(params.PageCursorStr) > 0 {
		if params.PeriodStart, periodStartStr, params.PeriodEnd, periodEndStr, params.Offset, err =
			HTTPFiatTransactionUnpackPageCursor(auth, params.PageCursorStr); err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid next page")
		}

		// Prepare next page cursor. Adjust offset to move along to next record set.
		if params.NextPage, err = HTTPFiatTransactionGeneratePageCursor(
			auth, periodStartStr, periodEndStr, params.Offset+params.PageSize); err != nil {
			logger.Info("failed to encrypt Fiat paginated transactions next page cursor", zap.Error(err))

			return http.StatusInternalServerError, fmt.Errorf("please retry your request later")
		}
	} else {
		if params.PeriodStart, params.PeriodEnd, params.NextPage, err =
			HTTPFiatTransactionInfoPaginatedRequest(auth,
				params.MonthStr, params.YearStr, params.TimezoneStr, params.PageSize); err != nil {
			logger.Info("failed to prepare time periods for paginated Fiat transaction details", zap.Error(err))

			return http.StatusInternalServerError, fmt.Errorf("please retry your request later")
		}
	}

	return 0, nil
}

// HTTPValidateOfferRequest will validate an offer request by checking the amount and Fiat currencies are valid.
func HTTPValidateOfferRequest(debitAmount decimal.Decimal, precision int32, fiatCurrencies ...string) (
	[]postgres.Currency, error) {
	var (
		err              error
		parsedCurrencies = make([]postgres.Currency, len(fiatCurrencies))
	)

	// Validate the source Fiat currency.
	for idx, fiatCurrencyCode := range fiatCurrencies {
		if err = parsedCurrencies[idx].Scan(fiatCurrencyCode); err != nil || !parsedCurrencies[idx].Valid() {
			return parsedCurrencies, fmt.Errorf("invalid Fiat currency %s", fiatCurrencyCode)
		}
	}

	// Check for correct decimal places in source amount.
	if !debitAmount.Equal(debitAmount.Truncate(precision)) || debitAmount.IsNegative() {
		return parsedCurrencies, fmt.Errorf("invalid source amount %s", debitAmount.String())
	}

	return parsedCurrencies, nil
}

// HTTPPrepareCryptoOffer will request the conversion rate, prepare the price quote, and store it in the Redis cache.
func HTTPPrepareCryptoOffer(auth auth.Auth, cache redis.Redis, logger logger.Logger, quotes quotes.Quotes,
	source, destination string, sourceAmount decimal.Decimal, isPurchase bool) (
	models.HTTPExchangeOfferResponse, int, string, error) {
	var (
		err          error
		offer        models.HTTPExchangeOfferResponse
		offerID      = xid.New().String()
		precision    = constants.GetDecimalPlacesFiat()
		fiatCurrency = source
	)

	// Configure precision, fiat, and crypto tickers for Crypto sale.
	if !isPurchase {
		precision = constants.GetDecimalPlacesCrypto()
		fiatCurrency = destination
	}

	// Validate the Fiat currency and source amount.
	if _, err = HTTPValidateOfferRequest(sourceAmount, precision, fiatCurrency); err != nil {
		return offer, http.StatusBadRequest, "invalid request", fmt.Errorf("%w", err)
	}

	// Compile exchange rate offer.
	if offer.Rate, offer.Amount, err = quotes.CryptoConversion(
		source, destination, sourceAmount, isPurchase, nil); err != nil {
		logger.Warn("failed to retrieve quote for Cryptocurrency purchase", zap.Error(err))

		return offer, http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
	}

	// Check to make sure there is a valid Cryptocurrency amount.
	if offer.Amount.IsZero() {
		msg := "cryptocurrency purchase amount is too small"

		return offer, http.StatusBadRequest, msg, errors.New(msg)
	}

	offer.SourceAcc = source
	offer.DestinationAcc = destination
	offer.DebitAmount = sourceAmount
	offer.Expires = time.Now().Add(constants.GetFiatOfferTTL()).Unix()

	// Encrypt offer ID before returning to client.
	if offer.OfferID, err = auth.EncryptToString([]byte(offerID)); err != nil {
		msg := "failed to encrypt offer ID for Crypto purchase"
		logger.Warn(msg, zap.Error(err))

		return offer, http.StatusInternalServerError, retryMessage, errors.New(msg)
	}

	// Store the offer in Redis.
	if err = cache.Set(offerID, &offer, constants.GetFiatOfferTTL()); err != nil {
		msg := "failed to store Cryptocurrency purchase offer in cache"
		logger.Warn(msg, zap.Error(err))

		return offer, http.StatusInternalServerError, retryMessage, errors.New(msg)
	}

	return offer, 0, "", nil
}
