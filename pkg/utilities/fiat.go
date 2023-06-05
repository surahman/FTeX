package utilities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/surahman/FTeX/pkg/auth"
	"github.com/surahman/FTeX/pkg/constants"
	"github.com/surahman/FTeX/pkg/logger"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/validator"
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

// HTTPFiatDeposit deposits a valid amount into a Fiat account.
func HTTPFiatDeposit(db postgres.Postgres, logger *logger.Logger, clientID uuid.UUID,
	request *models.HTTPDepositCurrencyRequest) (*postgres.FiatAccountTransferResult, int, string, any, error) {
	var (
		pgCurrency      postgres.Currency
		err             error
		transferReceipt *postgres.FiatAccountTransferResult
	)

	if err = validator.ValidateStruct(request); err != nil {
		return nil, http.StatusBadRequest, "validation", err.Error(), fmt.Errorf("%w", err)
	}

	// Extract and validate the currency.
	if err = pgCurrency.Scan(request.Currency); err != nil || !pgCurrency.Valid() {
		return nil, http.StatusBadRequest, "invalid currency", request.Currency, fmt.Errorf("%w", err)
	}

	// Check for correct decimal places.
	if !request.Amount.Equal(request.Amount.Truncate(constants.GetDecimalPlacesFiat())) || request.Amount.IsNegative() {
		return nil, http.StatusBadRequest, "invalid amount", request.Amount, fmt.Errorf("%w", err)
	}

	if transferReceipt, err = db.FiatExternalTransfer(context.Background(),
		&postgres.FiatTransactionDetails{
			ClientID: clientID,
			Currency: pgCurrency,
			Amount:   request.Amount}); err != nil {
		var createErr *postgres.Error
		if !errors.As(err, &createErr) {
			logger.Info("failed to unpack deposit Fiat account error", zap.Error(err))

			return nil, http.StatusInternalServerError, retryMessage, nil, fmt.Errorf("%w", err)
		}

		return nil, createErr.Code, createErr.Message, nil, fmt.Errorf("%w", err)
	}

	return transferReceipt, 0, "", nil, nil
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
