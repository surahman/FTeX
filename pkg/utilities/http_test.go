package utilities

import (
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/postgres"
)

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

func TestUtilities_HTTPFiatTransactionInfoPaginatedRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		monthStr     string
		yearStr      string
		timezone     string
		limit        string
		offset       string
		expectStart  string
		expectEnd    string
		expectErr    require.ErrorAssertionFunc
		expectLimit  int32
		expectOffset int32
	}{
		{
			name:         "valid - Before UTC",
			monthStr:     "6",
			yearStr:      "2023",
			timezone:     "-04:00",
			limit:        "4",
			offset:       "9",
			expectStart:  fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "-04:00"),
			expectEnd:    fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "-04:00"),
			expectLimit:  4,
			expectOffset: 9,
			expectErr:    require.NoError,
		}, {
			name:         "valid - no timezone",
			monthStr:     "6",
			yearStr:      "2023",
			timezone:     "",
			limit:        "4",
			offset:       "9",
			expectStart:  fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "+00:00"),
			expectEnd:    fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "+00:00"),
			expectLimit:  4,
			expectOffset: 9,
			expectErr:    require.NoError,
		}, {
			name:         "invalid - 0 month",
			monthStr:     "0",
			yearStr:      "2023",
			timezone:     "+04:00",
			limit:        "4",
			offset:       "9",
			expectStart:  fmt.Sprintf(constants.GetMonthFormatString(), 2023, 0, "+04:00"),
			expectEnd:    fmt.Sprintf(constants.GetMonthFormatString(), 2023, 0, "+04:00"),
			expectLimit:  4,
			expectOffset: 9,
			expectErr:    require.Error,
		}, {
			name:         "valid - no limit",
			monthStr:     "6",
			yearStr:      "2023",
			timezone:     "+04:00",
			limit:        "",
			offset:       "9",
			expectStart:  fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "+04:00"),
			expectEnd:    fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "+04:00"),
			expectLimit:  10,
			expectOffset: 9,
			expectErr:    require.NoError,
		}, {
			name:         "valid - no offset",
			monthStr:     "6",
			yearStr:      "2023",
			timezone:     "+04:00",
			limit:        "10",
			offset:       "",
			expectStart:  fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "+04:00"),
			expectEnd:    fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "+04:00"),
			expectLimit:  10,
			expectOffset: 0,
			expectErr:    require.NoError,
		}, {
			name:         "valid - June",
			monthStr:     "6",
			yearStr:      "2023",
			timezone:     "+04:00",
			limit:        "4",
			offset:       "9",
			expectStart:  fmt.Sprintf(constants.GetMonthFormatString(), 2023, 6, "+04:00"),
			expectEnd:    fmt.Sprintf(constants.GetMonthFormatString(), 2023, 7, "+04:00"),
			expectLimit:  4,
			expectOffset: 9,
			expectErr:    require.NoError,
		}, {
			name:         "valid - December",
			monthStr:     "12",
			yearStr:      "2023",
			timezone:     "+04:00",
			limit:        "4",
			offset:       "9",
			expectStart:  fmt.Sprintf(constants.GetMonthFormatString(), 2023, 12, "+04:00"),
			expectEnd:    fmt.Sprintf(constants.GetMonthFormatString(), 2024, 1, "+04:00"),
			expectLimit:  4,
			expectOffset: 9,
			expectErr:    require.NoError,
		}, {
			name:         "valid - January",
			monthStr:     "1",
			yearStr:      "2023",
			timezone:     "+04:00",
			limit:        "4",
			offset:       "9",
			expectStart:  fmt.Sprintf(constants.GetMonthFormatString(), 2023, 1, "+04:00"),
			expectEnd:    fmt.Sprintf(constants.GetMonthFormatString(), 2023, 2, "+04:00"),
			expectLimit:  4,
			expectOffset: 9,
			expectErr:    require.NoError,
		},
	}

	for _, testCase := range testCases {
		test := testCase

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			startTS, endTS, limit, offset, pageCursor, err := HTTPFiatTransactionInfoPaginatedRequest(
				testAuth, test.monthStr, test.yearStr, test.timezone, test.limit, test.offset)
			test.expectErr(t, err, "failed error expectation")

			if err != nil {
				return
			}

			require.Equal(t, test.expectLimit, limit, "limit mismatch.")
			require.Equal(t, test.expectOffset, offset, "offset mismatch.")
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
