package utilities

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/postgres"
)

// HTTPFiatBalancePaginatedRequest will convert the encrypted URL query parameter for the currency and the record limit
// and covert them to a currency and integer record limit.
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
func HTTPFiatTransactionInfoPaginatedRequest(monthStr, yearStr, timezoneStr, limitStr, offsetStr string) (
	pgtype.Timestamptz, pgtype.Timestamptz, int32, int32, error) {
	var (
		limit       int64
		offset      int64
		startYear   int64
		startMonth  int64
		endYear     int64
		endMonth    int64
		startTime   time.Time
		endTime     time.Time
		periodStart pgtype.Timestamptz
		periodEnd   pgtype.Timestamptz
		err         error
	)

	// Configure empty timezone to Zulu/UTC.
	if len(timezoneStr) == 0 {
		timezoneStr = "+00:00"
	}

	// Extract year and month.
	if startYear, err = strconv.ParseInt(yearStr, 10, 32); err != nil {
		return periodStart, periodEnd, -1, -1, fmt.Errorf("invalid year")
	}

	if startMonth, err = strconv.ParseInt(monthStr, 10, 32); err != nil {
		return periodStart, periodEnd, -1, -1, fmt.Errorf("invalid month")
	}

	// Setup end year and month.
	endYear = startYear
	endMonth = startMonth + 1

	if endMonth == 13 { //nolint:gomnd
		endMonth = 1
		endYear++
	}

	// Prepare Postgres timestamps.
	if startTime, err = time.Parse(time.RFC3339,
		fmt.Sprintf(constants.GetMonthFormatString(), startYear, startMonth, timezoneStr)); err != nil {
		return periodStart, periodEnd, -1, -1, fmt.Errorf("start date parse failure %w", err)
	}

	if err = periodStart.Scan(startTime); err != nil {
		return periodStart, periodEnd, -1, -1, fmt.Errorf("invalid start date %w", err)
	}

	if endTime, err = time.Parse(time.RFC3339,
		fmt.Sprintf(constants.GetMonthFormatString(), endYear, endMonth, timezoneStr)); err != nil {
		return periodStart, periodEnd, -1, -1, fmt.Errorf("end date parse failure %w", err)
	}

	if err = periodEnd.Scan(endTime); err != nil {
		return periodStart, periodEnd, -1, -1, fmt.Errorf("end start date %w", err)
	}

	// Convert record limit to int and set base bound for bad input.
	if len(limitStr) > 0 {
		if limit, err = strconv.ParseInt(limitStr, 10, 32); err != nil {
			return periodStart, periodEnd, -1, -1, fmt.Errorf("failed to parse record limit %w", err)
		}
	}

	if limit < 1 {
		limit = 10
	}

	// Extract offset.
	if len(offsetStr) > 0 {
		if offset, err = strconv.ParseInt(offsetStr, 10, 32); err != nil {
			return periodStart, periodEnd, -1, -1, fmt.Errorf("failed to parse offset limit %w", err)
		}
	}

	return periodStart, periodEnd, int32(limit), int32(offset), nil
}

// HTTPFiatTransactionGeneratePageCursor will generate the encrypted page cursor.
func HTTPFiatTransactionGeneratePageCursor() {

}
