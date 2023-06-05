package utilities

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/postgres"
	"go.uber.org/zap"
)

// HTTPFiatOpen opens a Fiat account.
func HTTPFiatOpen(db postgres.Postgres, logger *logger.Logger, clientID uuid.UUID, currency string) (
	int, string, error) {
	var (
		pgCurrency postgres.Currency
		err        error
	)

	// Extract and validate the currency.
	if err = pgCurrency.Scan(currency); err != nil || !pgCurrency.Valid() {
		return http.StatusBadRequest, "invalid currency", fmt.Errorf("%w", err)
	}

	if err = db.FiatCreateAccount(clientID, pgCurrency); err != nil {
		var createErr *postgres.Error
		if !errors.As(err, &createErr) {
			logger.Info("failed to unpack open Fiat account error", zap.Error(err))

			return http.StatusInternalServerError, retryMessage, fmt.Errorf("%w", err)
		}

		return createErr.Code, createErr.Message, fmt.Errorf("%w", err)
	}

	return 0, "", nil
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
