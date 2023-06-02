package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.31

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/surahman/FTeX/pkg/constants"
	graphql_generated "github.com/surahman/FTeX/pkg/graphql/generated"
	"github.com/surahman/FTeX/pkg/models"
	"github.com/surahman/FTeX/pkg/postgres"
	"github.com/surahman/FTeX/pkg/utilities"
	"github.com/surahman/FTeX/pkg/validator"
	"go.uber.org/zap"
)

// Currency is the resolver for the currency field.
func (r *fiatAccountResolver) Currency(ctx context.Context, obj *postgres.FiatAccount) (string, error) {
	return string(obj.Currency), nil
}

// Balance is the resolver for the balance field.
func (r *fiatAccountResolver) Balance(ctx context.Context, obj *postgres.FiatAccount) (float64, error) {
	return obj.Balance.InexactFloat64(), nil
}

// LastTx is the resolver for the lastTx field.
func (r *fiatAccountResolver) LastTx(ctx context.Context, obj *postgres.FiatAccount) (float64, error) {
	return obj.LastTx.InexactFloat64(), nil
}

// LastTxTs is the resolver for the lastTxTs field.
func (r *fiatAccountResolver) LastTxTs(ctx context.Context, obj *postgres.FiatAccount) (string, error) {
	return obj.LastTxTs.Time.String(), nil
}

// CreatedAt is the resolver for the createdAt field.
func (r *fiatAccountResolver) CreatedAt(ctx context.Context, obj *postgres.FiatAccount) (string, error) {
	return obj.CreatedAt.Time.String(), nil
}

// ClientID is the resolver for the clientID field.
func (r *fiatAccountResolver) ClientID(ctx context.Context, obj *postgres.FiatAccount) (string, error) {
	return obj.ClientID.String(), nil
}

// TxID is the resolver for the txId field.
func (r *fiatDepositResponseResolver) TxID(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error) {
	return obj.TxID.String(), nil
}

// ClientID is the resolver for the clientId field.
func (r *fiatDepositResponseResolver) ClientID(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error) {
	return obj.ClientID.String(), nil
}

// TxTimestamp is the resolver for the txTimestamp field.
func (r *fiatDepositResponseResolver) TxTimestamp(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error) {
	return obj.TxTS.Time.String(), nil
}

// Balance is the resolver for the balance field.
func (r *fiatDepositResponseResolver) Balance(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error) {
	return obj.Balance.String(), nil
}

// LastTx is the resolver for the lastTx field.
func (r *fiatDepositResponseResolver) LastTx(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error) {
	return obj.LastTx.String(), nil
}

// Currency is the resolver for the currency field.
func (r *fiatDepositResponseResolver) Currency(ctx context.Context, obj *postgres.FiatAccountTransferResult) (string, error) {
	return string(obj.Currency), nil
}

// DebitAmount is the resolver for the DebitAmount field.
func (r *fiatExchangeOfferResponseResolver) DebitAmount(ctx context.Context, obj *models.HTTPExchangeOfferResponse) (float64, error) {
	return obj.DebitAmount.InexactFloat64(), nil
}

// Currency is the resolver for the currency field.
func (r *fiatJournalResolver) Currency(ctx context.Context, obj *postgres.FiatJournal) (string, error) {
	return string(obj.Currency), nil
}

// Amount is the resolver for the amount field.
func (r *fiatJournalResolver) Amount(ctx context.Context, obj *postgres.FiatJournal) (float64, error) {
	return obj.Amount.InexactFloat64(), nil
}

// TransactedAt is the resolver for the transactedAt field.
func (r *fiatJournalResolver) TransactedAt(ctx context.Context, obj *postgres.FiatJournal) (string, error) {
	return obj.TransactedAt.Time.String(), nil
}

// ClientID is the resolver for the clientID field.
func (r *fiatJournalResolver) ClientID(ctx context.Context, obj *postgres.FiatJournal) (string, error) {
	return obj.ClientID.String(), nil
}

// TxID is the resolver for the txID field.
func (r *fiatJournalResolver) TxID(ctx context.Context, obj *postgres.FiatJournal) (string, error) {
	return obj.TxID.String(), nil
}

// OpenFiat is the resolver for the openFiat field.
func (r *mutationResolver) OpenFiat(ctx context.Context, currency string) (*models.FiatOpenAccountResponse, error) {
	var (
		clientID   uuid.UUID
		pgCurrency postgres.Currency
		err        error
	)

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	// Extract and validate the currency.
	if err = pgCurrency.Scan(currency); err != nil || !pgCurrency.Valid() {
		return nil, errors.New("invalid currency")
	}

	if err = r.db.FiatCreateAccount(clientID, pgCurrency); err != nil {
		var createErr *postgres.Error
		if !errors.As(err, &createErr) {
			r.logger.Info("failed to unpack open Fiat account error", zap.Error(err))

			return nil, errors.New("please retry your request later")
		}

		return nil, errors.New(createErr.Message)
	}

	return &models.FiatOpenAccountResponse{ClientID: clientID.String(), Currency: currency}, nil
}

// DepositFiat is the resolver for the depositFiat field.
func (r *mutationResolver) DepositFiat(ctx context.Context, input models.HTTPDepositCurrencyRequest) (*postgres.FiatAccountTransferResult, error) {
	var (
		clientID        uuid.UUID
		currency        postgres.Currency
		err             error
		transferReceipt *postgres.FiatAccountTransferResult
	)

	if err = validator.ValidateStruct(&input); err != nil {
		return nil, fmt.Errorf("validation %w", err)
	}

	// Extract and validate the currency.
	if err = currency.Scan(input.Currency); err != nil || !currency.Valid() {
		return nil, fmt.Errorf("invalid currency")
	}

	// Check for correct decimal places.
	if !input.Amount.Equal(input.Amount.Truncate(constants.GetDecimalPlacesFiat())) || input.Amount.IsNegative() {
		return nil, fmt.Errorf("invalid amount")
	}

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	if transferReceipt, err = r.db.FiatExternalTransfer(context.Background(),
		&postgres.FiatTransactionDetails{
			ClientID: clientID,
			Currency: currency,
			Amount:   input.Amount}); err != nil {
		var createErr *postgres.Error
		if !errors.As(err, &createErr) {
			r.logger.Info("failed to unpack deposit Fiat account error", zap.Error(err))

			return nil, errors.New("please retry your request later")
		}

		return nil, errors.New(createErr.Message)
	}

	return transferReceipt, nil
}

// ExchangeOfferFiat is the resolver for the exchangeOfferFiat field.
func (r *mutationResolver) ExchangeOfferFiat(ctx context.Context, input models.HTTPExchangeOfferRequest) (*models.HTTPExchangeOfferResponse, error) {
	var (
		err     error
		offer   models.HTTPExchangeOfferResponse
		offerID = xid.New().String()
	)

	if err = validator.ValidateStruct(&input); err != nil {
		return nil, fmt.Errorf("validation %w", err)
	}

	// Extract and validate the currency.
	if _, err = utilities.HTTPValidateOfferRequest(
		input.SourceAmount, constants.GetDecimalPlacesFiat(), input.SourceCurrency, input.DestinationCurrency); err != nil {

		return nil, errors.New(constants.GetInvalidRequest())
	}

	if offer.ClientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	// Compile exchange rate offer.
	if offer.Rate, offer.Amount, err = r.quotes.FiatConversion(
		input.SourceCurrency, input.DestinationCurrency, input.SourceAmount, nil); err != nil {
		r.logger.Warn("failed to retrieve quote for Fiat currency conversion", zap.Error(err))

		return nil, errors.New("please retry your request later")
	}

	offer.SourceAcc = input.SourceCurrency
	offer.DestinationAcc = input.DestinationCurrency
	offer.DebitAmount = input.SourceAmount
	offer.Expires = time.Now().Add(constants.GetFiatOfferTTL()).Unix()

	// Encrypt offer ID before returning to client.
	if offer.OfferID, err = r.auth.EncryptToString([]byte(offerID)); err != nil {
		r.logger.Warn("failed to encrypt offer ID for Fiat conversion", zap.Error(err))

		return nil, errors.New("please retry your request later")
	}

	// Store the offer in Redis.
	if err = r.cache.Set(offerID, &offer, constants.GetFiatOfferTTL()); err != nil {
		r.logger.Warn("failed to store Fiat conversion offer in cache", zap.Error(err))

		return nil, errors.New("please retry your request later")
	}

	return &offer, nil
}

// ExchangeTransferFiat is the resolver for the exchangeTransferFiat field.
func (r *mutationResolver) ExchangeTransferFiat(ctx context.Context, offerID string) (*models.FiatExchangeTransferResponse, error) {
	var (
		err              error
		clientID         uuid.UUID
		offer            models.HTTPExchangeOfferResponse
		receipt          models.FiatExchangeTransferResponse
		parsedCurrencies []postgres.Currency
	)

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	// Extract Offer ID from request.
	{
		var rawOfferID []byte
		if rawOfferID, err = r.auth.DecryptFromString(offerID); err != nil {
			r.logger.Warn("failed to decrypt Offer ID for Fiat transfer request", zap.Error(err))

			return nil, errors.New("please retry your request later")
		}

		offerID = string(rawOfferID)
	}

	// Retrieve the offer from Redis. Once retrieved, the entry must be removed from the cache to block re-use of
	// the offer. If a database update fails below this point the user will need to re-request an offer.
	{
		var msg string
		if offer, _, msg, err = utilities.HTTPGetCachedOffer(r.cache, r.logger, offerID); err != nil {
			return nil, errors.New(msg)
		}
	}

	// Verify that the client IDs match.
	if clientID != offer.ClientID {
		r.logger.Warn("clientID mismatch with Fiat Offer stored in Redis",
			zap.Strings("Requester & Offer Client IDs", []string{clientID.String(), offer.ClientID.String()}))

		return nil, errors.New("please retry your request later")
	}

	// Get currency codes.
	if parsedCurrencies, err = utilities.HTTPValidateOfferRequest(
		offer.Amount, constants.GetDecimalPlacesFiat(), offer.SourceAcc, offer.DestinationAcc); err != nil {
		r.logger.Warn("failed to extract source and destination currencies from Fiat exchange offer",
			zap.Error(err))

		return nil, errors.New(err.Error())
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

	if receipt.SourceReceipt, receipt.DestinationReceipt, err = r.db.
		FiatInternalTransfer(context.Background(), srcTxDetails, dstTxDetails); err != nil {
		r.logger.Warn("failed to complete internal Fiat transfer", zap.Error(err))

		return nil, errors.New("please check you have both currency accounts and enough funds")
	}

	return &receipt, nil
}

// BalanceFiat is the resolver for the balanceFiat field.
func (r *queryResolver) BalanceFiat(ctx context.Context, currencyCode string) (*postgres.FiatAccount, error) {
	var (
		accDetails postgres.FiatAccount
		clientID   uuid.UUID
		currency   postgres.Currency
		err        error
	)

	// Extract and validate the currency.
	if err = currency.Scan(currencyCode); err != nil || !currency.Valid() {
		return nil, errors.New("invalid currency")
	}

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	if accDetails, err = r.db.FiatBalanceCurrency(clientID, currency); err != nil {
		var balanceErr *postgres.Error
		if !errors.As(err, &balanceErr) {
			r.logger.Info("failed to unpack Fiat account balance currency error", zap.Error(err))

			return nil, errors.New("please retry your request later")
		}

		return nil, errors.New(balanceErr.Message)
	}

	return &accDetails, nil
}

// BalanceAllFiat is the resolver for the balanceAllFiat field.
func (r *queryResolver) BalanceAllFiat(ctx context.Context, pageCursor *string, pageSize *int32) (*models.HTTPFiatDetailsPaginated, error) {
	var (
		accDetails []postgres.FiatAccount
		clientID   uuid.UUID
		currency   postgres.Currency
		err        error
	)

	if pageSize == nil {
		pageSize = new(int32)
	}

	if pageCursor == nil {
		pageCursor = new(string)
	}

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	// Extract and assemble the page cursor and page size.
	if currency, *pageSize, err = utilities.HTTPFiatBalancePaginatedRequest(
		r.auth, *pageCursor, strconv.Itoa(int(*pageSize))); err != nil {

		return nil, errors.New("invalid page cursor or page size")
	}

	if accDetails, err = r.db.FiatBalanceCurrencyPaginated(clientID, currency, *pageSize+1); err != nil {
		var balanceErr *postgres.Error
		if !errors.As(err, &balanceErr) {
			r.logger.Info("failed to unpack Fiat account balance currency error", zap.Error(err))

			return nil, errors.New("please retry your request later")
		}

		return nil, errors.New(balanceErr.Message)
	}

	// Reset and generate the next page link by pulling the last item returned if the page size is N + 1 of the requested.
	*pageCursor = ""
	lastRecordIdx := int(*pageSize)
	if len(accDetails) == lastRecordIdx+1 {
		// Generate next page link.
		if *pageCursor, err = r.auth.EncryptToString([]byte(accDetails[*pageSize].Currency)); err != nil {
			r.logger.Error("failed to encrypt Fiat currency for use as cursor", zap.Error(err))

			return nil, errors.New("please retry your request later")
		}

		// Remove last element.
		accDetails = accDetails[:*pageSize]
	}

	return &models.HTTPFiatDetailsPaginated{
		AccountBalances: accDetails,
		Links: models.HTTPLinks{
			PageCursor: *pageCursor,
		},
	}, nil
}

// TransactionDetailsFiat is the resolver for the transactionDetailsFiat field.
func (r *queryResolver) TransactionDetailsFiat(ctx context.Context, transactionID string) ([]postgres.FiatJournal, error) {
	var (
		journalEntries []postgres.FiatJournal
		clientID       uuid.UUID
		txID           uuid.UUID
		err            error
	)

	// Extract and validate the transactionID.
	if txID, err = uuid.FromString(transactionID); err != nil {
		return nil, fmt.Errorf("invalid transaction id %s", transactionID)
	}

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	if journalEntries, err = r.db.FiatTxDetailsCurrency(clientID, txID); err != nil {
		var balanceErr *postgres.Error
		if !errors.As(err, &balanceErr) {
			r.logger.Info("failed to unpack Fiat account balance transactionID error", zap.Error(err))

			return nil, errors.New("please retry your request later")
		}

		return nil, errors.New(balanceErr.Message)
	}

	if len(journalEntries) == 0 {
		return nil, errors.New("transaction id not found")
	}

	return journalEntries, nil
}

// TransactionDetailsAllFiat is the resolver for the transactionDetailsAllFiat field.
func (r *queryResolver) TransactionDetailsAllFiat(ctx context.Context, input models.FiatPaginatedTxDetailsRequest) (*models.FiatTransactionsPaginated, error) {
	var (
		journalEntries []postgres.FiatJournal
		clientID       uuid.UUID
		currency       postgres.Currency
		err            error
		params         utilities.HTTPPaginatedTxParams
	)

	if input.PageSize == nil {
		input.PageSize = new(string)
	}
	params.PageSizeStr = *input.PageSize

	if input.PageCursor == nil {
		input.PageCursor = new(string)
	}
	params.PageCursorStr = *input.PageCursor

	if input.Timezone == nil {
		input.Timezone = new(string)
	}
	params.TimezoneStr = *input.Timezone

	if input.Month == nil {
		input.Month = new(string)
	}
	params.MonthStr = *input.Month

	if input.Year == nil {
		input.Year = new(string)
	}
	params.YearStr = *input.Year

	if clientID, _, err = AuthorizationCheck(ctx, r.auth, r.logger, r.authHeaderKey); err != nil {
		return nil, errors.New("authorization failure")
	}

	// Extract and validate the currency.
	if err = currency.Scan(input.Currency); err != nil || !currency.Valid() {
		return nil, fmt.Errorf("invalid currency %s", currency)
	}

	// Check for required parameters.
	if len(*input.PageCursor) == 0 && (len(*input.Month) == 0 || len(*input.Year) == 0) {
		return nil, errors.New("missing required parameters")
	}

	// Decrypt values from page cursor, if present. Otherwise, prepare values using query strings.
	if _, err = utilities.HTTPTxParseQueryParams(r.auth, r.logger, &params); err != nil {
		return nil, errors.New(err.Error())
	}

	// Retrieve transaction details page.
	if journalEntries, err = r.db.FiatTransactionsCurrencyPaginated(
		clientID, currency, params.PageSize+1, params.Offset, params.PeriodStart, params.PeriodEnd); err != nil {
		var balanceErr *postgres.Error
		if !errors.As(err, &balanceErr) {
			r.logger.Info("failed to unpack Fiat transactions request error", zap.Error(err))

			return nil, errors.New("please retry your request later")
		}

		return nil, errors.New(balanceErr.Message)
	}

	if len(journalEntries) == 0 {
		return nil, errors.New("no transactions")
	}

	// Check if there are further pages of data. If not, set the next link to be empty.
	if len(journalEntries) > int(params.PageSize) {
		journalEntries = journalEntries[:int(params.PageSize)]
	} else {
		params.NextPage = ""
	}

	return &models.FiatTransactionsPaginated{
		Transactions: journalEntries,
		Links: &models.HTTPLinks{
			PageCursor: params.NextPage,
		},
	}, nil
}

// Amount is the resolver for the amount field.
func (r *fiatDepositRequestResolver) Amount(ctx context.Context, obj *models.HTTPDepositCurrencyRequest, data float64) error {
	obj.Amount = decimal.NewFromFloat(data)

	return nil
}

// SourceAmount is the resolver for the sourceAmount field.
func (r *fiatExchangeOfferRequestResolver) SourceAmount(ctx context.Context, obj *models.HTTPExchangeOfferRequest, data float64) error {
	obj.SourceAmount = decimal.NewFromFloat(data)

	return nil
}

// FiatAccount returns graphql_generated.FiatAccountResolver implementation.
func (r *Resolver) FiatAccount() graphql_generated.FiatAccountResolver {
	return &fiatAccountResolver{r}
}

// FiatDepositResponse returns graphql_generated.FiatDepositResponseResolver implementation.
func (r *Resolver) FiatDepositResponse() graphql_generated.FiatDepositResponseResolver {
	return &fiatDepositResponseResolver{r}
}

// FiatExchangeOfferResponse returns graphql_generated.FiatExchangeOfferResponseResolver implementation.
func (r *Resolver) FiatExchangeOfferResponse() graphql_generated.FiatExchangeOfferResponseResolver {
	return &fiatExchangeOfferResponseResolver{r}
}

// FiatJournal returns graphql_generated.FiatJournalResolver implementation.
func (r *Resolver) FiatJournal() graphql_generated.FiatJournalResolver {
	return &fiatJournalResolver{r}
}

// FiatDepositRequest returns graphql_generated.FiatDepositRequestResolver implementation.
func (r *Resolver) FiatDepositRequest() graphql_generated.FiatDepositRequestResolver {
	return &fiatDepositRequestResolver{r}
}

// FiatExchangeOfferRequest returns graphql_generated.FiatExchangeOfferRequestResolver implementation.
func (r *Resolver) FiatExchangeOfferRequest() graphql_generated.FiatExchangeOfferRequestResolver {
	return &fiatExchangeOfferRequestResolver{r}
}

type fiatAccountResolver struct{ *Resolver }
type fiatDepositResponseResolver struct{ *Resolver }
type fiatExchangeOfferResponseResolver struct{ *Resolver }
type fiatJournalResolver struct{ *Resolver }
type fiatDepositRequestResolver struct{ *Resolver }
type fiatExchangeOfferRequestResolver struct{ *Resolver }
