package utilities

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/quotes"
	"github.com/surahman/FTeX/pkg/redis"
)

func TestUtilities_HTTPGetCachedOffer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		expectErrMsg  string
		expectStatus  int
		expectErr     require.ErrorAssertionFunc
		redisGetErr   error
		redisGetData  models.HTTPExchangeOfferResponse
		redisGetTimes int
		redisDelErr   error
		redisDelTimes int
	}{
		{
			name:          "get unknown error",
			expectErrMsg:  "retry",
			expectStatus:  http.StatusInternalServerError,
			expectErr:     require.Error,
			redisGetErr:   errors.New("unknown error"),
			redisGetData:  models.HTTPExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   nil,
			redisDelTimes: 0,
		}, {
			name:          "get unknown package error",
			expectErrMsg:  "retry",
			expectStatus:  http.StatusInternalServerError,
			expectErr:     require.Error,
			redisGetErr:   redis.ErrCacheUnknown,
			redisGetData:  models.HTTPExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   nil,
			redisDelTimes: 0,
		}, {
			name:          "get package error",
			expectErrMsg:  "expired",
			expectStatus:  http.StatusRequestTimeout,
			expectErr:     require.Error,
			redisGetErr:   redis.ErrCacheMiss,
			redisGetData:  models.HTTPExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   nil,
			redisDelTimes: 0,
		}, {
			name:          "del unknown error",
			expectErrMsg:  "retry",
			expectStatus:  http.StatusInternalServerError,
			expectErr:     require.Error,
			redisGetErr:   nil,
			redisGetData:  models.HTTPExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   errors.New("unknown error"),
			redisDelTimes: 1,
		}, {
			name:          "del unknown package error",
			expectErrMsg:  "retry",
			expectStatus:  http.StatusInternalServerError,
			expectErr:     require.Error,
			redisGetErr:   nil,
			redisGetData:  models.HTTPExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   redis.ErrCacheUnknown,
			redisDelTimes: 1,
		}, {
			name:          "del cache miss",
			expectErrMsg:  "",
			expectStatus:  http.StatusOK,
			expectErr:     require.NoError,
			redisGetErr:   nil,
			redisGetData:  models.HTTPExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   redis.ErrCacheMiss,
			redisDelTimes: 1,
		}, {
			name:          "valid",
			expectErrMsg:  "",
			expectStatus:  http.StatusOK,
			expectErr:     require.NoError,
			redisGetErr:   nil,
			redisGetData:  models.HTTPExchangeOfferResponse{},
			redisGetTimes: 1,
			redisDelErr:   nil,
			redisDelTimes: 1,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockCache := mocks.NewMockRedis(mockCtrl)

			gomock.InOrder(
				mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(test.redisGetErr).
					SetArg(1, test.redisGetData).
					Times(test.redisGetTimes),

				mockCache.EXPECT().Del(gomock.Any()).
					Return(test.redisDelErr).
					Times(test.redisDelTimes),
			)

			_, status, msg, err := HTTPGetCachedOffer(mockCache, zapLogger, "SOME-OFFER-ID")
			test.expectErr(t, err, "error expectation failed.")
			require.Equal(t, test.expectStatus, status, "expected and actual status codes did not match.")
			require.Contains(t, msg, test.expectErrMsg, "expected error message did not match.")
		})
	}
}

func TestUtilities_HTTPFiatBalancePaginatedRequest(t *testing.T) {
	t.Parallel()

	encAED, err := testAuth.EncryptToString([]byte("AED"))
	require.NoError(t, err, "failed to encrypt AED currency.")

	encUSD, err := testAuth.EncryptToString([]byte("USD"))
	require.NoError(t, err, "failed to encrypt USD currency.")

	encEUR, err := testAuth.EncryptToString([]byte("EUR"))
	require.NoError(t, err, "failed to encrypt EUR currency.")

	testCases := []struct {
		name           string
		currencyStr    string
		limitStr       string
		expectCurrency postgres.Currency
		expectLimit    int32
	}{
		{
			name:           "empty currency",
			currencyStr:    "",
			limitStr:       "5",
			expectCurrency: postgres.CurrencyAED,
			expectLimit:    5,
		}, {
			name:           "AED",
			currencyStr:    encAED,
			limitStr:       "5",
			expectCurrency: postgres.CurrencyAED,
			expectLimit:    5,
		}, {
			name:           "USD",
			currencyStr:    encUSD,
			limitStr:       "5",
			expectCurrency: postgres.CurrencyUSD,
			expectLimit:    5,
		}, {
			name:           "EUR",
			currencyStr:    encEUR,
			limitStr:       "5",
			expectCurrency: postgres.CurrencyEUR,
			expectLimit:    5,
		}, {
			name:           "base bound limit",
			currencyStr:    encEUR,
			limitStr:       "0",
			expectCurrency: postgres.CurrencyEUR,
			expectLimit:    10,
		}, {
			name:           "above base bound limit",
			currencyStr:    encEUR,
			limitStr:       "999",
			expectCurrency: postgres.CurrencyEUR,
			expectLimit:    999,
		}, {
			name:           "empty request",
			currencyStr:    "",
			limitStr:       "",
			expectCurrency: postgres.CurrencyAED,
			expectLimit:    10,
		}, {
			name:           "empty currency",
			currencyStr:    "",
			limitStr:       "999",
			expectCurrency: postgres.CurrencyAED,
			expectLimit:    999,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actualCurr, actualLimit, err := HTTPFiatBalancePaginatedRequest(testAuth, test.currencyStr, test.limitStr)
			require.NoError(t, err, "error returned from query unpacking")
			require.Equal(t, test.expectCurrency, actualCurr, "currencies mismatched.")
			require.Equal(t, test.expectLimit, actualLimit, "request limit size mismatched.")
		})
	}
}

func TestUtilities_EncryptDecryptTransactionPageCursor(t *testing.T) {
	t.Parallel()

	startStr := fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "-04:00")
	endStr := fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "-04:00")

	startTS, err := time.Parse(time.RFC3339, startStr)
	require.NoError(t, err, "failed to parse expected start time.")

	pgStart := pgtype.Timestamptz{}
	require.NoError(t, pgStart.Scan(startTS), "failed to scan start timestamp to pg.")

	// Check end timestamp.
	endTS, err := time.Parse(time.RFC3339, endStr)
	require.NoError(t, err, "failed to parse expected start time.")

	pgEnd := pgtype.Timestamptz{}
	require.NoError(t, pgEnd.Scan(endTS), "failed to scan start timestamp to pg.")

	t.Run("successful encryption-decryption", func(t *testing.T) {
		t.Parallel()
		encrypted, err := HTTPFiatTransactionGeneratePageCursor(testAuth, startStr, endStr, 10)
		require.NoError(t, err, "failed to encrypt cursor.")
		require.True(t, len(encrypted) > 0, "empty encrypted cursor returned.")

		actualStart, actualStartStr, actualEnd, actualEndStr, actualOffset, err :=
			HTTPFiatTransactionUnpackPageCursor(testAuth, encrypted)
		require.NoError(t, err, "failed to decrypt page cursor.")
		require.Equal(t, pgStart, actualStart, "start period mismatch.")
		require.Equal(t, startStr, actualStartStr, "start period string mismatch.")
		require.Equal(t, pgEnd, actualEnd, "end period mismatch.")
		require.Equal(t, endStr, actualEndStr, "end period string mismatch.")
		require.Equal(t, int32(10), actualOffset, "offset mismatch.")
	})

	t.Run("missing offset", func(t *testing.T) {
		t.Parallel()
		input, err := testAuth.EncryptToString([]byte("start,end"))
		require.NoError(t, err, "failed to encrypt missing offset.")
		_, _, _, _, _, err = HTTPFiatTransactionUnpackPageCursor(testAuth, input)
		require.Error(t, err, "decrypted invalid page cursor.")
	})

	t.Run("invalid offset", func(t *testing.T) {
		t.Parallel()
		input, err := testAuth.EncryptToString([]byte("start,end,invalid-offset"))
		require.NoError(t, err, "failed to encrypt invalid offset.")
		_, _, _, _, _, err = HTTPFiatTransactionUnpackPageCursor(testAuth, input)
		require.Error(t, err, "decrypted invalid page cursor.")
	})
}

func TestUtilities_HTTPFiatTransactionInfoPaginatedRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		monthStr    string
		yearStr     string
		timezone    string
		expectStart string
		expectEnd   string
		expectErr   require.ErrorAssertionFunc
	}{
		{
			name:        "valid - Before UTC",
			monthStr:    "6",
			yearStr:     "2023",
			timezone:    "-04:00",
			expectStart: fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "-04:00"),
			expectEnd:   fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "-04:00"),
			expectErr:   require.NoError,
		}, {
			name:        "valid - no timezone",
			monthStr:    "6",
			yearStr:     "2023",
			timezone:    "",
			expectStart: fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "+00:00"),
			expectEnd:   fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "+00:00"),
			expectErr:   require.NoError,
		}, {
			name:        "invalid - 0 month",
			monthStr:    "0",
			yearStr:     "2023",
			timezone:    "+04:00",
			expectStart: fmt.Sprintf(constants.GetMonthFormatString(), 2023, 0, "+04:00"),
			expectEnd:   fmt.Sprintf(constants.GetMonthFormatString(), 2023, 0, "+04:00"),
			expectErr:   require.Error,
		}, {
			name:        "valid - no limit",
			monthStr:    "6",
			yearStr:     "2023",
			timezone:    "+04:00",
			expectStart: fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "+04:00"),
			expectEnd:   fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "+04:00"),
			expectErr:   require.NoError,
		}, {
			name:        "valid - no offset",
			monthStr:    "6",
			yearStr:     "2023",
			timezone:    "+04:00",
			expectStart: fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "+04:00"),
			expectEnd:   fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "+04:00"),
			expectErr:   require.NoError,
		}, {
			name:        "valid - June",
			monthStr:    "6",
			yearStr:     "2023",
			timezone:    "+04:00",
			expectStart: fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "+04:00"),
			expectEnd:   fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "+04:00"),
			expectErr:   require.NoError,
		}, {
			name:        "valid - December",
			monthStr:    "12",
			yearStr:     "2023",
			timezone:    "+04:00",
			expectStart: fmt.Sprintf(constants.GetMonthFormatString(), 2023, 12, "+04:00"),
			expectEnd:   fmt.Sprintf(constants.GetMonthFormatString(), 2024, 1, "+04:00"),
			expectErr:   require.NoError,
		}, {
			name:        "valid - January",
			monthStr:    "1",
			yearStr:     "2023",
			timezone:    "+04:00",
			expectStart: fmt.Sprintf(constants.GetMonthFormatString(), 2023, 1, "+04:00"),
			expectEnd:   fmt.Sprintf(constants.GetMonthFormatString(), 2023, 2, "+04:00"),
			expectErr:   require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			startTS, endTS, pageCursor, err := HTTPFiatTransactionInfoPaginatedRequest(
				testAuth, test.monthStr, test.yearStr, test.timezone, 3)
			test.expectErr(t, err, "failed error expectation")

			if err != nil {
				return
			}

			require.True(t, len(pageCursor) > 100, "empty page cursor returned.")

			// Check start timestamp.
			expectedStartTS, err := time.Parse(time.RFC3339, test.expectStart)
			require.NoError(t, err, "failed to parse expected start time.")

			expectedPgStart := pgtype.Timestamptz{}
			require.NoError(t, expectedPgStart.Scan(expectedStartTS), "failed to scan start timestamp to pg.")
			require.Equal(t, expectedPgStart, startTS, "start timestamp mismatch.")

			// Check end timestamp.
			expectedEndTS, err := time.Parse(time.RFC3339, test.expectEnd)
			require.NoError(t, err, "failed to parse expected start time.")

			expectedPgEnd := pgtype.Timestamptz{}
			require.NoError(t, expectedPgEnd.Scan(expectedEndTS), "failed to scan start timestamp to pg.")
			require.Equal(t, expectedPgEnd, endTS, "end timestamp mismatch.")
		})
	}
}

func TestUtilities_HTTPFiatPaginatedTxParams(t *testing.T) {
	t.Parallel()

	// Timestamps for start and end of period.
	startStr := fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "-04:00")
	endStr := fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "-04:00")

	startTS, err := time.Parse(time.RFC3339, startStr)
	require.NoError(t, err, "failed to parse expected start time.")

	pgStart := pgtype.Timestamptz{}
	require.NoError(t, pgStart.Scan(startTS), "failed to scan start timestamp to pg.")

	endTS, err := time.Parse(time.RFC3339, endStr)
	require.NoError(t, err, "failed to parse expected start time.")

	pgEnd := pgtype.Timestamptz{}
	require.NoError(t, pgEnd.Scan(endTS), "failed to scan start timestamp to pg.")

	// Encrypted page cursors.
	cursorOffset10, err := HTTPFiatTransactionGeneratePageCursor(testAuth, startStr, endStr, 10)
	require.NoError(t, err, "failed to create page cursor with offset 10.")

	cursorOffset33, err := HTTPFiatTransactionGeneratePageCursor(testAuth, startStr, endStr, 33)
	require.NoError(t, err, "failed to create page cursor with offset 33.")

	testCases := []struct {
		params         *HTTPFiatPaginatedTxParams
		name           string
		expectOffset   int32
		expectPageSize int32
		expectErrCode  int
		expectErrMsg   string
		expectErr      require.ErrorAssertionFunc
	}{
		{
			name: "invalid page size",
			params: &HTTPFiatPaginatedTxParams{
				PageSizeStr:   "bad-input",
				PageCursorStr: cursorOffset33,
			},
			expectOffset:   0,
			expectPageSize: 0,
			expectErrMsg:   "invalid page size",
			expectErrCode:  http.StatusBadRequest,
			expectErr:      require.Error,
		}, {
			name: "valid - page size na, offset 33 request",
			params: &HTTPFiatPaginatedTxParams{
				PageCursorStr: cursorOffset33,
			},
			expectOffset:   33,
			expectPageSize: 10,
			expectErrMsg:   "",
			expectErrCode:  0,
			expectErr:      require.NoError,
		}, {
			name: "valid - page size 7, offset 33 request",
			params: &HTTPFiatPaginatedTxParams{
				PageSizeStr:   "7",
				PageCursorStr: cursorOffset33,
			},
			expectOffset:   33,
			expectPageSize: 7,
			expectErrMsg:   "",
			expectErrCode:  0,
			expectErr:      require.NoError,
		}, {
			name: "valid - page size 3, offset 10 request",
			params: &HTTPFiatPaginatedTxParams{
				PageSizeStr:   "3",
				PageCursorStr: cursorOffset10,
			},
			expectOffset:   10,
			expectPageSize: 3,
			expectErrMsg:   "",
			expectErrCode:  0,
			expectErr:      require.NoError,
		}, {
			name: "valid - initial request",
			params: &HTTPFiatPaginatedTxParams{
				PageSizeStr:   "3",
				PageCursorStr: "",
				TimezoneStr:   "-04:00",
				MonthStr:      "6",
				YearStr:       "2023",
			},
			expectOffset:   0,
			expectPageSize: 3,
			expectErrMsg:   "",
			expectErrCode:  0,
			expectErr:      require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			code, err := HTTPTxParseQueryParams(testAuth, zapLogger, test.params)
			test.expectErr(t, err, "error expectation failed.")
			require.Equal(t, test.expectErrCode, code, "error codes mismatched.")

			if err != nil {
				return
			}

			require.Equal(t, test.expectOffset, test.params.Offset, "offset mismatched.")
			require.Equal(t, test.expectPageSize, test.params.PageSize, "page size mismatched.")
			require.Equal(t, pgStart, test.params.PeriodStart, "start timestamp mismatched.")
			require.Equal(t, pgEnd, test.params.PeriodEnd, "end timestamp mismatched.")
			require.True(t, len(test.params.NextPage) > 100, "next page cursor is incorrect.")
		})
	}
}

func TestUtilities_HTTPValidateOfferRequest(t *testing.T) {
	t.Parallel()

	amountValid, err := decimal.NewFromString("10101.11")
	require.NoError(t, err, "failed to parse valid amount.")

	amountInvalidNegative, err := decimal.NewFromString("-10101.11")
	require.NoError(t, err, "failed to parse invalid negative amount")

	amountInvalidDecimal, err := decimal.NewFromString("10101.111")
	require.NoError(t, err, "failed to parse invalid decimal amount")

	testCases := []struct {
		name         string
		expectErrMsg string
		currencies   []string
		amount       decimal.Decimal
		expectErr    require.ErrorAssertionFunc
	}{
		{
			name:         "valid",
			expectErrMsg: "",
			currencies:   []string{"USD", "CAD"},
			amount:       amountValid,
			expectErr:    require.NoError,
		}, {
			name:         "invalid source currency",
			expectErrMsg: "invalid Fiat currency",
			currencies:   []string{"INVALID", "CAD"},
			amount:       amountValid,
			expectErr:    require.Error,
		}, {
			name:         "invalid destination currency",
			expectErrMsg: "invalid Fiat currency",
			currencies:   []string{"USD", "INVALID"},
			amount:       amountValid,
			expectErr:    require.Error,
		}, {
			name:         "invalid negative amount",
			expectErrMsg: "source amount",
			currencies:   []string{"USD", "CAD"},
			amount:       amountInvalidNegative,
			expectErr:    require.Error,
		}, {
			name:         "invalid decimal amount",
			expectErrMsg: "source amount",
			currencies:   []string{"USD", "CAD"},
			amount:       amountInvalidDecimal,
			expectErr:    require.Error,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsedCurrencies, err := HTTPValidateOfferRequest(test.amount, constants.GetDecimalPlacesFiat(), test.currencies...)
			test.expectErr(t, err, "error expectation failed.")

			if err != nil {
				require.Contains(t, err.Error(), test.expectErrMsg, "error message is incorrect.")

				return
			}

			require.Equal(t, len(test.currencies), len(parsedCurrencies), "incorrect number of parsed currencies returned.")

			for idx, actualCurrency := range parsedCurrencies {
				var expectedCurrency postgres.Currency
				require.NoError(t, expectedCurrency.Scan(test.currencies[idx]), "failed to parse expected currency.")
				require.Equal(t, expectedCurrency, actualCurrency, "parse currency mismatched.")
			}
		})
	}
}

func TestUtilities_HTTPPrepareCryptoOffer(t *testing.T) { //nolint: maintidx
	t.Parallel()

	var (
		sourceAmount = decimal.NewFromFloat(23123.12)
		quotesRate   = decimal.NewFromFloat(23100)
	)

	testCases := []struct {
		name             string
		source           string
		destination      string
		expectErrMsg     string
		httpMessage      string
		httpStatus       int
		isPurchase       bool
		quotesAmount     decimal.Decimal
		quotesTimes      int
		quotesErr        error
		authEncryptTimes int
		authEncryptErr   error
		redisTimes       int
		redisErr         error
		expectErr        require.ErrorAssertionFunc
	}{
		{
			name:             "validate offer - purchase",
			source:           "INVALID",
			destination:      "BTC",
			expectErrMsg:     "INVALID",
			httpMessage:      constants.GetInvalidRequest(),
			httpStatus:       http.StatusBadRequest,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      0,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "crypto conversion - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "quote failure",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        errors.New("quote failure"),
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "zero amount - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "too small",
			httpMessage:      "too small",
			httpStatus:       http.StatusBadRequest,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(0),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "encryption failure - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "failed to encrypt",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   errors.New("encryption failure"),
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "cache failure - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "failed to store",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         errors.New("cache failure"),
			expectErr:        require.Error,
		}, {
			name:             "valid - purchase",
			source:           "USD",
			destination:      "BTC",
			expectErrMsg:     "",
			httpMessage:      "",
			httpStatus:       0,
			isPurchase:       true,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         nil,
			expectErr:        require.NoError,
		}, {
			name:             "validate offer - sale",
			source:           "BTC",
			destination:      "INVALID",
			expectErrMsg:     "INVALID",
			httpMessage:      constants.GetInvalidRequest(),
			httpStatus:       http.StatusBadRequest,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      0,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "crypto conversion - sale",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "quote failure",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        errors.New("quote failure"),
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "zero amount - sale",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "too small",
			httpMessage:      "too small",
			httpStatus:       http.StatusBadRequest,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(0),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 0,
			authEncryptErr:   nil,
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "encryption failure - sale",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "failed to encrypt",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   errors.New("encryption failure"),
			redisTimes:       0,
			redisErr:         nil,
			expectErr:        require.Error,
		}, {
			name:             "cache failure - sale",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "failed to store",
			httpMessage:      retryMessage,
			httpStatus:       http.StatusInternalServerError,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         errors.New("cache failure"),
			expectErr:        require.Error,
		}, {
			name:             "valid - sell",
			source:           "BTC",
			destination:      "USD",
			expectErrMsg:     "",
			httpMessage:      "",
			httpStatus:       0,
			isPurchase:       false,
			quotesAmount:     decimal.NewFromFloat(1.23),
			quotesTimes:      1,
			quotesErr:        nil,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisTimes:       1,
			redisErr:         nil,
			expectErr:        require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockCache := mocks.NewMockRedis(mockCtrl)
			mockQuotes := quotes.NewMockQuotes(mockCtrl)

			gomock.InOrder(
				mockQuotes.EXPECT().CryptoConversion(
					test.source, test.destination, sourceAmount, test.isPurchase, nil).
					Return(quotesRate, test.quotesAmount, test.quotesErr).
					Times(test.quotesTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("OFFER-ID", test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockCache.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(test.redisErr).
					Times(test.redisTimes),
			)

			offer, status, msg, err := HTTPPrepareCryptoOffer(mockAuth, mockCache, zapLogger, mockQuotes,
				uuid.UUID{}, test.source, test.destination, sourceAmount, test.isPurchase)
			test.expectErr(t, err, "error expectation failed.")

			if err != nil {
				require.Contains(t, err.Error(), test.expectErrMsg, "error message is incorrect.")
				require.Contains(t, msg, test.httpMessage, "http error message mismatched.")
				require.Equal(t, test.httpStatus, status, "http status mismatched.")

				return
			}

			require.Equal(t, test.source, offer.SourceAcc, "source account mismatch.")
			require.Equal(t, test.destination, offer.DestinationAcc, "destination account mismatch.")
			require.Equal(t, sourceAmount, offer.DebitAmount, "debit amount mismatch.")
			require.Equal(t, quotesRate, offer.Rate, "offer rate mismatch.")
			require.Equal(t, test.quotesAmount, offer.Amount, "offer amount mismatch.")
		})
	}
}

func TestUtilities_HTTPExchangeCrypto(t *testing.T) { //nolint: maintidx
	t.Parallel()

	validClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate a valid uuid.")

	invalidClientID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate a valid uuid.")

	cryptoAmount := decimal.NewFromFloat(1234.56)
	fiatAmount := decimal.NewFromFloat(78910.11)

	validFiat := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "BTC",
			DestinationAcc: "USD",
			Rate:           decimal.Decimal{},
			Amount:         decimal.Decimal{},
		},
		DebitAmount:      decimal.Decimal{},
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: false,
		IsCryptoSale:     false,
	}

	invalidSale := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "USD",
			DestinationAcc: "BTC", // should be Fiat currency ticker.
			Rate:           decimal.Decimal{},
			Amount:         fiatAmount,
		},
		DebitAmount:      cryptoAmount,
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: false,
		IsCryptoSale:     true,
	}

	validSale := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "BTC",
			DestinationAcc: "USD",
			Rate:           decimal.Decimal{},
			Amount:         fiatAmount,
		},
		DebitAmount:      cryptoAmount,
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: false,
		IsCryptoSale:     true,
	}

	validPurchase := models.HTTPExchangeOfferResponse{
		PriceQuote: models.PriceQuote{
			ClientID:       validClientID,
			SourceAcc:      "USD",
			DestinationAcc: "BTC",
			Rate:           decimal.Decimal{},
			Amount:         cryptoAmount,
		},
		DebitAmount:      fiatAmount,
		OfferID:          "OFFER-ID",
		Expires:          0,
		IsCryptoPurchase: true,
		IsCryptoSale:     false,
	}

	testCases := []struct {
		name             string
		expectErrMsg     string
		clientID         uuid.UUID
		httpStatus       int
		authEncryptTimes int
		authEncryptErr   error
		redisGetData     models.HTTPExchangeOfferResponse
		redisGetTimes    int
		redisGetErr      error
		redisDelTimes    int
		redisDelErr      error
		purchaseTimes    int
		purchaseErr      error
		sellTimes        int
		sellErr          error
		expectErr        require.ErrorAssertionFunc
	}{
		{
			name:             "valid - purchase",
			clientID:         validClientID,
			expectErrMsg:     "",
			httpStatus:       0,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validPurchase,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    1,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.NoError,
		}, {
			name:             "decrypt failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "retry",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   errors.New("decrypt failure"),
			redisGetData:     validSale,
			redisGetTimes:    0,
			redisGetErr:      nil,
			redisDelTimes:    0,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "cache get failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "retry",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      errors.New("cache get failure"),
			redisDelTimes:    0,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "cache del failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "retry",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      errors.New("cache del failure"),
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "cache del failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "invalid",
			httpStatus:       http.StatusBadRequest,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validFiat,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "clientID mismatch - sell",
			clientID:         invalidClientID,
			expectErrMsg:     "retry",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "validation failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "Fiat",
			httpStatus:       http.StatusBadRequest,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     invalidSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        0,
			sellErr:          nil,
			expectErr:        require.Error,
		}, {
			name:             "transaction failure - sell",
			clientID:         validClientID,
			expectErrMsg:     "sell failure",
			httpStatus:       http.StatusInternalServerError,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        1,
			sellErr:          errors.New("sell failure"),
			expectErr:        require.Error,
		}, {
			name:             "valid - sell",
			clientID:         validClientID,
			expectErrMsg:     "",
			httpStatus:       0,
			authEncryptTimes: 1,
			authEncryptErr:   nil,
			redisGetData:     validSale,
			redisGetTimes:    1,
			redisGetErr:      nil,
			redisDelTimes:    1,
			redisDelErr:      nil,
			purchaseTimes:    0,
			purchaseErr:      nil,
			sellTimes:        1,
			sellErr:          nil,
			expectErr:        require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockCache := mocks.NewMockRedis(mockCtrl)
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().DecryptFromString(gomock.Any()).
					Return([]byte("OFFER-ID"), test.authEncryptErr).
					Times(test.authEncryptTimes),

				mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(test.redisGetErr).
					SetArg(1, test.redisGetData).
					Times(test.redisGetTimes),

				mockCache.EXPECT().Del(gomock.Any()).
					Return(test.redisDelErr).
					Times(test.redisDelTimes),

				mockPostgres.EXPECT().CryptoPurchase(
					gomock.Any(), postgres.CurrencyUSD, fiatAmount, "BTC", cryptoAmount).
					Return(&postgres.FiatJournal{}, &postgres.CryptoJournal{}, test.purchaseErr).
					Times(test.purchaseTimes),

				mockPostgres.EXPECT().CryptoSell(
					gomock.Any(), postgres.CurrencyUSD, fiatAmount, "BTC", cryptoAmount).
					Return(&postgres.FiatJournal{}, &postgres.CryptoJournal{}, test.sellErr).
					Times(test.sellTimes),
			)

			_, status, errMsg, err :=
				HTTPExchangeCrypto(mockAuth, mockCache, mockPostgres, zapLogger, test.clientID, "offer-id")
			test.expectErr(t, err, "error expectation failed.")

			require.Equal(t, test.httpStatus, status, "http status code mismatched.")
			require.Contains(t, errMsg, test.expectErrMsg, "http error message mismatched.")
		})
	}
}

func TestUtilities_HTTPTxDetailsCrypto(t *testing.T) {
	t.Parallel()

	validTxID, err := uuid.NewV4()
	require.NoError(t, err, "failed to generate uuid.")

	cryptoJournal := []postgres.CryptoJournal{{}, {}}
	fiatJournal := []postgres.FiatJournal{{}, {}}

	testCases := []struct {
		name          string
		txID          string
		expectErrMsg  string
		httpStatus    int
		cryptoJournal []postgres.CryptoJournal
		fiatJournal   []postgres.FiatJournal
		cryptoTxErr   error
		cryptoTxTimes int
		fiatTxErr     error
		fiatTxTimes   int
		expectErr     require.ErrorAssertionFunc
	}{
		{
			name:          "invalid transaction id",
			txID:          "invalid-transaction-id",
			expectErrMsg:  "invalid transaction ID",
			httpStatus:    http.StatusBadRequest,
			cryptoJournal: cryptoJournal,
			fiatJournal:   fiatJournal,
			fiatTxErr:     nil,
			fiatTxTimes:   0,
			cryptoTxErr:   nil,
			cryptoTxTimes: 0,
			expectErr:     require.Error,
		}, {
			name:          "unknown fiat db error",
			txID:          validTxID.String(),
			expectErrMsg:  "please retry",
			httpStatus:    http.StatusInternalServerError,
			fiatJournal:   fiatJournal,
			fiatTxErr:     errors.New("unknown db error"),
			fiatTxTimes:   1,
			cryptoJournal: cryptoJournal,
			cryptoTxErr:   nil,
			cryptoTxTimes: 0,
			expectErr:     require.Error,
		}, {
			name:          "unknown crypto db error",
			txID:          validTxID.String(),
			expectErrMsg:  "please retry",
			httpStatus:    http.StatusInternalServerError,
			fiatJournal:   fiatJournal,
			fiatTxErr:     nil,
			fiatTxTimes:   1,
			cryptoJournal: cryptoJournal,
			cryptoTxErr:   errors.New("unknown db error"),
			cryptoTxTimes: 1,
			expectErr:     require.Error,
		}, {
			name:          "known fiat db error",
			txID:          validTxID.String(),
			expectErrMsg:  "could not complete",
			httpStatus:    http.StatusInternalServerError,
			fiatJournal:   fiatJournal,
			fiatTxErr:     postgres.ErrTransactFiat,
			fiatTxTimes:   1,
			cryptoJournal: cryptoJournal,
			cryptoTxErr:   nil,
			cryptoTxTimes: 0,
			expectErr:     require.Error,
		}, {
			name:          "known crypto db error",
			txID:          validTxID.String(),
			expectErrMsg:  "could not complete",
			httpStatus:    http.StatusInternalServerError,
			fiatJournal:   fiatJournal,
			fiatTxErr:     nil,
			fiatTxTimes:   1,
			cryptoJournal: cryptoJournal,
			cryptoTxErr:   postgres.ErrTransactCrypto,
			cryptoTxTimes: 1,
			expectErr:     require.Error,
		}, {
			name:          "empty result set",
			txID:          validTxID.String(),
			expectErrMsg:  "id not found",
			httpStatus:    http.StatusNotFound,
			fiatJournal:   []postgres.FiatJournal{},
			fiatTxErr:     nil,
			fiatTxTimes:   1,
			cryptoJournal: []postgres.CryptoJournal{},
			cryptoTxErr:   nil,
			cryptoTxTimes: 1,
			expectErr:     require.Error,
		}, {
			name:          "valid",
			txID:          validTxID.String(),
			expectErrMsg:  "",
			httpStatus:    0,
			fiatJournal:   fiatJournal,
			fiatTxErr:     nil,
			fiatTxTimes:   1,
			cryptoJournal: cryptoJournal,
			cryptoTxErr:   nil,
			cryptoTxTimes: 1,
			expectErr:     require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockPostgres.EXPECT().FiatTxDetailsCurrency(gomock.Any(), gomock.Any()).
					Return(test.fiatJournal, test.fiatTxErr).
					Times(test.fiatTxTimes),

				mockPostgres.EXPECT().CryptoTxDetailsCurrency(gomock.Any(), gomock.Any()).
					Return(test.cryptoJournal, test.cryptoTxErr).
					Times(test.cryptoTxTimes),
			)

			_, status, errMsg, err := HTTPTxDetails(mockPostgres, zapLogger, uuid.UUID{}, test.txID)
			test.expectErr(t, err, "error expectation failed.")

			require.Equal(t, test.httpStatus, status, "http status code mismatched.")
			require.Contains(t, errMsg, test.expectErrMsg, "http error message mismatched.")
		})
	}
}

func TestUtilities_HTTPCryptoBalancePaginatedRequest(t *testing.T) {
	t.Parallel()

	encBTC, err := testAuth.EncryptToString([]byte("BTC"))
	require.NoError(t, err, "failed to encrypt BTC currency.")

	encETH, err := testAuth.EncryptToString([]byte("ETH"))
	require.NoError(t, err, "failed to encrypt ETH currency.")

	encXRP, err := testAuth.EncryptToString([]byte("XRP"))
	require.NoError(t, err, "failed to encrypt XRP currency.")

	testCases := []struct {
		name            string
		encryptedTicker string
		limitStr        string
		expectTicker    string
		expectLimit     int32
	}{
		{
			name:            "empty currency",
			encryptedTicker: "",
			limitStr:        "5",
			expectTicker:    "BTC",
			expectLimit:     5,
		}, {
			name:            "BTC",
			encryptedTicker: encBTC,
			limitStr:        "5",
			expectTicker:    "BTC",
			expectLimit:     5,
		}, {
			name:            "ETH",
			encryptedTicker: encETH,
			limitStr:        "5",
			expectTicker:    "ETH",
			expectLimit:     5,
		}, {
			name:            "XRP",
			encryptedTicker: encXRP,
			limitStr:        "5",
			expectTicker:    "XRP",
			expectLimit:     5,
		}, {
			name:            "base bound limit",
			encryptedTicker: encXRP,
			limitStr:        "0",
			expectTicker:    "XRP",
			expectLimit:     10,
		}, {
			name:            "above base bound limit",
			encryptedTicker: encXRP,
			limitStr:        "999",
			expectTicker:    "XRP",
			expectLimit:     999,
		}, {
			name:            "empty request",
			encryptedTicker: "",
			limitStr:        "",
			expectTicker:    "BTC",
			expectLimit:     10,
		}, {
			name:            "empty currency",
			encryptedTicker: "",
			limitStr:        "999",
			expectTicker:    "BTC",
			expectLimit:     999,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actualTicker, actualLimit, err := HTTPCryptoBalancePaginatedRequest(testAuth, test.encryptedTicker, test.limitStr)
			require.NoError(t, err, "error returned from query unpacking")
			require.Equal(t, test.expectTicker, actualTicker, "tickers mismatched.")
			require.Equal(t, test.expectLimit, actualLimit, "request limit size mismatched.")
		})
	}
}

func TestUtilities_HTTPCryptoTxPaginated(t *testing.T) {
	t.Parallel()

	var (
		pageCursor  = "some-page-cursor"
		fourRecords = []postgres.CryptoAccount{{}, {}, {}, {}}
	)

	testCases := []struct {
		name               string
		pageSize           string
		expectNextPageSize string
		expectErrMsg       string
		httpStatus         int
		expectedRecordsLen int
		expectedPageSize   int32
		decryptStringErr   error
		decryptStringTimes int
		balanceData        []postgres.CryptoAccount
		balanceErr         error
		balanceTimes       int
		encryptStringErr   error
		encryptStingTimes  int
		expectErr          require.ErrorAssertionFunc
		expectNextPage     require.BoolAssertionFunc
	}{
		{
			name:               "cursor bad page failure",
			pageSize:           "bad-page-size",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       "page size",
			httpStatus:         http.StatusBadRequest,
			expectedRecordsLen: 0,
			expectedPageSize:   3,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         nil,
			balanceTimes:       0,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.Error,
			expectNextPage:     require.True,
		}, {
			name:               "cursor decryption failure",
			pageSize:           "3",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       "invalid page cursor",
			httpStatus:         http.StatusBadRequest,
			expectedRecordsLen: 0,
			expectedPageSize:   3,
			decryptStringErr:   errors.New("decrypt failure"),
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         nil,
			balanceTimes:       0,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.Error,
			expectNextPage:     require.True,
		}, {
			name:               "db failure - known error",
			pageSize:           "3",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       "not found",
			httpStatus:         http.StatusNotFound,
			expectedRecordsLen: 0,
			expectedPageSize:   3,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         postgres.ErrNotFound,
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.Error,
			expectNextPage:     require.True,
		}, {
			name:               "db failure - unknown error",
			pageSize:           "3",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       retryMessage,
			httpStatus:         http.StatusInternalServerError,
			expectedRecordsLen: 0,
			expectedPageSize:   3,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         errors.New("unknown db failure"),
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.Error,
			expectNextPage:     require.True,
		}, {
			name:               "valid - default page size",
			pageSize:           "0",
			expectNextPageSize: "",
			expectErrMsg:       "",
			httpStatus:         0,
			expectedRecordsLen: 4,
			expectedPageSize:   10,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         nil,
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.NoError,
			expectNextPage:     require.False,
		}, {
			name:               "valid - no next page",
			pageSize:           "4",
			expectNextPageSize: "",
			expectErrMsg:       "",
			httpStatus:         0,
			expectedRecordsLen: 3,
			expectedPageSize:   4,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        []postgres.CryptoAccount{{}, {}, {}},
			balanceErr:         nil,
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  0,
			expectErr:          require.NoError,
			expectNextPage:     require.False,
		}, {
			name:               "valid - has next page",
			pageSize:           "3",
			expectNextPageSize: "pageSize=3",
			expectErrMsg:       "",
			httpStatus:         0,
			expectedRecordsLen: 3,
			expectedPageSize:   3,
			decryptStringErr:   nil,
			decryptStringTimes: 1,
			balanceData:        fourRecords,
			balanceErr:         nil,
			balanceTimes:       1,
			encryptStringErr:   nil,
			encryptStingTimes:  1,
			expectErr:          require.NoError,
			expectNextPage:     require.True,
		},
	}

	for _, testCase := range testCases {
		test := testCase
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			// Mock configurations.
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockAuth := mocks.NewMockAuth(mockCtrl)
			mockPostgres := mocks.NewMockPostgres(mockCtrl)

			gomock.InOrder(
				mockAuth.EXPECT().DecryptFromString(pageCursor).
					Return([]byte("decrypted-string"), test.decryptStringErr).
					Times(test.decryptStringTimes),

				mockPostgres.EXPECT().CryptoBalancePaginated(gomock.Any(), gomock.Any(), test.expectedPageSize+1).
					Return(test.balanceData, test.balanceErr).
					Times(test.balanceTimes),

				mockAuth.EXPECT().EncryptToString(gomock.Any()).
					Return("next-page-cursor", test.encryptStringErr).
					Times(test.encryptStingTimes),
			)

			actualDetails, status, errMsg, err := HTTPCryptoTxPaginated(mockAuth, mockPostgres, zapLogger,
				uuid.UUID{}, pageCursor, test.pageSize)
			test.expectErr(t, err, "error expectation failed.")

			require.Equal(t, test.httpStatus, status, "http status code mismatched.")
			require.Contains(t, errMsg, test.expectErrMsg, "http error message mismatched.")

			if err != nil {
				return
			}

			require.Equal(t, 0, len(actualDetails.Links.PageCursor), "page cursor set.")
			require.Contains(t, actualDetails.Links.NextPage, test.expectNextPageSize, "expected page size mismatch.")
			test.expectNextPage(t, len(actualDetails.Links.NextPage) > 0, "next page link not set.")
			require.Equal(t, test.expectedRecordsLen, len(actualDetails.AccountBalances),
				"number of returned records mismatched")
		})
	}
}
