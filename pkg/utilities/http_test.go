package utilities

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/mocks"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
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

			code, err := HTTPFiatTxParseQueryParams(testAuth, zapLogger, test.params)
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
