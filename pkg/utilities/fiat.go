package utilities

import (
	"context"
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
		return http.StatusBadRequest, constants.GetInvalidCurrencyString(), fmt.Errorf("%w", err)
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
		return nil, http.StatusBadRequest, constants.GetValidationString(), err.Error(), fmt.Errorf("%w", err)
	}

	// Extract and validate the currency.
	if err = pgCurrency.Scan(request.Currency); err != nil || !pgCurrency.Valid() {
		return nil, http.StatusBadRequest, constants.GetInvalidCurrencyString(), request.Currency, fmt.Errorf("%w", err)
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

// HTTPFiatOffer retrieves an exchange rate offer from a quote provider and stores it in the Redis session cache.
func HTTPFiatOffer(auth auth.Auth, cache redis.Redis, logger *logger.Logger, quotes quotes.Quotes, clientID uuid.UUID,
	request *models.HTTPExchangeOfferRequest) (*models.HTTPExchangeOfferResponse, int, string, any, error) {
	var (
		err     error
		offer   models.HTTPExchangeOfferResponse
		offerID = xid.New().String()
	)

	if err = validator.ValidateStruct(request); err != nil {
		return nil, http.StatusBadRequest, constants.GetValidationString(), err.Error(), fmt.Errorf("%w", err)
	}

	// Extract and validate the currency.
	if _, err = HTTPValidateOfferRequest(request.SourceAmount, constants.GetDecimalPlacesFiat(),
		request.SourceCurrency, request.DestinationCurrency); err != nil {
		return nil, http.StatusBadRequest, constants.GetInvalidRequest(), err.Error(), fmt.Errorf("%w", err)
	}

	// Compile exchange rate offer.
	if offer.Rate, offer.Amount, err = quotes.FiatConversion(
		request.SourceCurrency, request.DestinationCurrency, request.SourceAmount, nil); err != nil {
		logger.Warn("failed to retrieve quote for Fiat currency conversion", zap.Error(err))

		return nil, http.StatusInternalServerError, retryMessage, nil, fmt.Errorf("%w", err)
	}

	// Check to make sure there is a valid Cryptocurrency amount.
	if !offer.Amount.GreaterThan(decimal.NewFromFloat(0)) {
		msg := "cryptocurrency purchase/sale amount is too small"

		return nil, http.StatusBadRequest, msg, nil, errors.New(msg)
	}

	offer.ClientID = clientID
	offer.SourceAcc = request.SourceCurrency
	offer.DestinationAcc = request.DestinationCurrency
	offer.DebitAmount = request.SourceAmount
	offer.Expires = time.Now().Add(constants.GetFiatOfferTTL()).Unix()

	// Encrypt offer ID before returning to client.
	if offer.OfferID, err = auth.EncryptToString([]byte(offerID)); err != nil {
		logger.Warn("failed to encrypt offer ID for Fiat conversion", zap.Error(err))

		return nil, http.StatusInternalServerError, retryMessage, nil, fmt.Errorf("%w", err)
	}

	// Store the offer in Redis.
	if err = cache.Set(offerID, &offer, constants.GetFiatOfferTTL()); err != nil {
		logger.Warn("failed to store Fiat conversion offer in cache", zap.Error(err))

		return nil, http.StatusInternalServerError, retryMessage, nil, fmt.Errorf("%w", err)
	}

	return &offer, 0, "", nil, nil
}

// HTTPFiatTransfer will retrieve an offer from the session cache, validate it, then update the database.
func HTTPFiatTransfer(auth auth.Auth, cache redis.Redis, db postgres.Postgres, logger *logger.Logger,
	clientID uuid.UUID, request *models.HTTPTransferRequest) (*models.HTTPFiatTransferResponse, int, string, any, error) {
	var (
		err              error
		offer            models.HTTPExchangeOfferResponse
		receipt          models.HTTPFiatTransferResponse
		offerID          string
		parsedCurrencies []postgres.Currency
	)

	if err = validator.ValidateStruct(request); err != nil {
		return nil, http.StatusBadRequest, constants.GetValidationString(), err.Error(), fmt.Errorf("%w", err)
	}

	// Extract Offer ID from request.
	{
		var rawOfferID []byte
		if rawOfferID, err = auth.DecryptFromString(request.OfferID); err != nil {
			logger.Warn("failed to decrypt Offer ID for Fiat transfer request", zap.Error(err))

			return nil, http.StatusInternalServerError, retryMessage, nil, fmt.Errorf("%w", err)
		}

		offerID = string(rawOfferID)
	}

	// Retrieve the offer from Redis. Once retrieved, the entry must be removed from the cache to block re-use of the
	// offer. If a database update fails below this point the user will need to re-request an offer.
	{
		var (
			httpStatus int
			httpMsg    string
		)
		if offer, httpStatus, httpMsg, err = HTTPGetCachedOffer(cache, logger, offerID); err != nil {
			return nil, httpStatus, httpMsg, nil, fmt.Errorf("%w", err)
		}
	}

	// Verify that the client IDs match.
	if clientID != offer.ClientID {
		logger.Warn("clientID mismatch with Fiat Offer stored in Redis",
			zap.Strings("Requester & Offer Client IDs", []string{clientID.String(), offer.ClientID.String()}))

		return nil, http.StatusInternalServerError, retryMessage, nil, fmt.Errorf("%w", err)
	}

	// Verify the offer is for a Fiat exchange.
	if offer.IsCryptoPurchase || offer.IsCryptoSale {
		return nil, http.StatusBadRequest, "invalid Fiat currency exchange offer", nil, fmt.Errorf("%w", err)
	}

	// Get currency codes.
	if parsedCurrencies, err = HTTPValidateOfferRequest(
		offer.Amount, constants.GetDecimalPlacesFiat(), offer.SourceAcc, offer.DestinationAcc); err != nil {
		logger.Warn("failed to extract source and destination currencies from Fiat exchange offer",
			zap.Error(err))

		return nil, http.StatusBadRequest, err.Error(), nil, fmt.Errorf("%w", err)
	}

	// Execute exchange.
	srcTxDetails := &postgres.FiatTransactionDetails{
		ClientID: offer.ClientID,
		Currency: parsedCurrencies[0],
		Amount:   offer.DebitAmount,
	}
	dstTxDetails := &postgres.FiatTransactionDetails{
		ClientID: offer.ClientID,
		Currency: parsedCurrencies[1],
		Amount:   offer.Amount,
	}

	if receipt.SrcTxReceipt, receipt.DstTxReceipt, err = db.
		FiatInternalTransfer(context.Background(), srcTxDetails, dstTxDetails); err != nil {
		logger.Warn("failed to complete internal Fiat transfer", zap.Error(err))

		return nil, http.StatusBadRequest, "please check you have both currency accounts and enough funds.",
			nil, fmt.Errorf("%w", err)
	}

	return &receipt, 0, "", nil, nil
}

// HTTPFiatBalance retrieves the account balance for a specific Fiat currency.
func HTTPFiatBalance(db postgres.Postgres, logger *logger.Logger, clientID uuid.UUID, ticker string) (
	*postgres.FiatAccount, int, string, any, error) {
	var (
		accDetails postgres.FiatAccount
		currency   postgres.Currency
		err        error
	)

	// Extract and validate the currency.
	if err = currency.Scan(ticker); err != nil || !currency.Valid() {
		return nil, http.StatusBadRequest, constants.GetInvalidCurrencyString(), ticker, fmt.Errorf("%w", err)
	}

	if accDetails, err = db.FiatBalance(clientID, currency); err != nil {
		var balanceErr *postgres.Error
		if !errors.As(err, &balanceErr) {
			logger.Info("failed to unpack Fiat account balance currency error", zap.Error(err))

			return nil, http.StatusInternalServerError, retryMessage, nil, fmt.Errorf("%w", err)
		}

		return nil, balanceErr.Code, balanceErr.Message, nil, fmt.Errorf("%w", err)
	}

	return &accDetails, 0, "", nil, nil
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
