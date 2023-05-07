package utilities

import (
	"fmt"
	"strconv"
	"strings"
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

// generateTimestampRange will generate the start and end timestamps with timezone information.
func generateTimestampRange(monthStr, yearStr, timezoneStr string) (
	pgtype.Timestamptz, string, pgtype.Timestamptz, string, error) {
	var (
		startYear  int64
		startMonth int64
		endYear    int64
		endMonth   int64
		startTime  time.Time
		endTime    time.Time
		startTSStr string
		startPGTS  pgtype.Timestamptz
		endTSStr   string
		endPGTS    pgtype.Timestamptz
		err        error
	)

	// Configure empty timezone to Zulu/UTC.
	if len(timezoneStr) == 0 {
		timezoneStr = "+00:00"
	}

	// Extract year and month.
	if startYear, err = strconv.ParseInt(yearStr, 10, 32); err != nil {
		return startPGTS, startTSStr, endPGTS, endTSStr, fmt.Errorf("invalid year")
	}

	if startMonth, err = strconv.ParseInt(monthStr, 10, 32); err != nil {
		return startPGTS, startTSStr, endPGTS, endTSStr, fmt.Errorf("invalid month")
	}

	// Setup end year and month.
	endYear = startYear
	endMonth = startMonth + 1

	if endMonth == 13 { //nolint:gomnd
		endMonth = 1
		endYear++
	}

	// Prepare Postgres timestamps.
	startTSStr = fmt.Sprintf(constants.GetMonthFormatString(), startYear, startMonth, timezoneStr)
	if startTime, err = time.Parse(time.RFC3339, startTSStr); err != nil {
		return startPGTS, startTSStr, endPGTS, endTSStr, fmt.Errorf("start date parse failure %w", err)
	}

	if err = startPGTS.Scan(startTime); err != nil {
		return startPGTS, startTSStr, endPGTS, endTSStr, fmt.Errorf("invalid start date %w", err)
	}

	endTSStr = fmt.Sprintf(constants.GetMonthFormatString(), endYear, endMonth, timezoneStr)
	if endTime, err = time.Parse(time.RFC3339, endTSStr); err != nil {
		return startPGTS, startTSStr, endPGTS, endTSStr, fmt.Errorf("end date parse failure %w", err)
	}

	if err = endPGTS.Scan(endTime); err != nil {
		return startPGTS, startTSStr, endPGTS, endTSStr, fmt.Errorf("end start date %w", err)
	}

	return startPGTS, startTSStr, endPGTS, endTSStr, nil
}

// HTTPFiatTransactionInfoPaginatedRequest will generate the month bounds and record limits using supplied query
// parameters.
func HTTPFiatTransactionInfoPaginatedRequest(
	auth auth.Auth,
	monthStr,
	yearStr,
	timezoneStr string) (pgtype.Timestamptz, pgtype.Timestamptz, string, error) {
	var (
		pageCursor     string
		periodStartStr string
		periodEndStr   string
		periodStart    pgtype.Timestamptz
		periodEnd      pgtype.Timestamptz
		err            error
	)

	// Generate timestamps.
	if periodStart, periodStartStr, periodEnd, periodEndStr, err =
		generateTimestampRange(monthStr, yearStr, timezoneStr); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("failed to prepare time range %w", err)
	}

	// Prepare page cursor.
	if pageCursor, err = HTTPFiatTransactionGeneratePageCursor(auth, periodStartStr, periodEndStr, 0); err != nil {
		return periodStart, periodEnd, pageCursor, fmt.Errorf("failed to encrypt page cursor %w", err)
	}

	return periodStart, periodEnd, pageCursor, nil
}

// HTTPFiatTransactionGeneratePageCursor will generate the encrypted page cursor.
//
//nolint:wrapcheck
func HTTPFiatTransactionGeneratePageCursor(auth auth.Auth, periodStartStr, periodEndStr string, offset int64) (
	string, error) {
	return auth.EncryptToString([]byte(fmt.Sprintf("%s,%s,%d", periodStartStr, periodEndStr, offset)))
}

// HTTPFiatTransactionUnpackPageCursor will unpack an encrypted page cursor to its component parts.
func HTTPFiatTransactionUnpackPageCursor(auth auth.Auth, pageCursor string) (
	pgtype.Timestamptz, pgtype.Timestamptz, int32, error) {
	var (
		startPGTS pgtype.Timestamptz
		endPGTS   pgtype.Timestamptz
		buffer    []byte
		err       error
	)

	if buffer, err = auth.DecryptFromString(pageCursor); err != nil {
		return startPGTS, endPGTS, -1, fmt.Errorf("failed to decrypt page cursor %w", err)
	}

	components := strings.Split(string(buffer), ",")
	if len(components) != 3 { //nolint:gomnd
		return startPGTS, endPGTS, -1, fmt.Errorf("decrypted page curror is invalid")
	}

	offset, err := strconv.ParseInt(components[2], 10, 32)
	if err != nil {
		return startPGTS, endPGTS, -1, fmt.Errorf("failed to parse offset %w", err)
	}

	// Prepare Postgres timestamps.
	startTime, err := time.Parse(time.RFC3339, components[0])
	if err != nil {
		return startPGTS, endPGTS, -1, fmt.Errorf("start date parse failure %w", err)
	}

	if err = startPGTS.Scan(startTime); err != nil {
		return startPGTS, endPGTS, -1, fmt.Errorf("invalid start date %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, components[1])
	if err != nil {
		return startPGTS, endPGTS, -1, fmt.Errorf("end date parse failure %w", err)
	}

	if err = endPGTS.Scan(endTime); err != nil {
		return startPGTS, endPGTS, -1, fmt.Errorf("end start date %w", err)
	}

	return startPGTS, endPGTS, int32(offset), nil
}
