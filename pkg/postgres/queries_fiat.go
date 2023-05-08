package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

// FiatCreateAccount is the interface through which external methods can create a Fiat account.
func (p *postgresImpl) FiatCreateAccount(clientID uuid.UUID, currency Currency) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	rowsAffected, err := p.Query.fiatCreateAccount(ctx, &fiatCreateAccountParams{ClientID: clientID, Currency: currency})
	if err != nil || rowsAffected != int64(1) {
		p.logger.Error("failed to create fiat account", zap.Error(err))

		return ErrCreateFiat
	}

	return nil
}

// FiatBalanceCurrency is the interface through which external methods can retrieve a Fiat account balance for a
// specific currency.
func (p *postgresImpl) FiatBalanceCurrency(clientID uuid.UUID, currency Currency) (FiatAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	balance, err := p.Query.fiatGetAccount(ctx, &fiatGetAccountParams{ClientID: clientID, Currency: currency})
	if err != nil {
		return FiatAccount{}, ErrNotFound
	}

	return balance, nil
}

// FiatTxDetailsCurrency is the interface through which external methods can retrieve a Fiat transaction details for a
// specific transaction.
func (p *postgresImpl) FiatTxDetailsCurrency(clientID uuid.UUID, transactionID uuid.UUID) ([]FiatJournal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	journal, err := p.Query.fiatGetJournalTransaction(ctx, &fiatGetJournalTransactionParams{clientID, transactionID})
	if err != nil {
		return nil, ErrNotFound
	}

	return journal, nil
}

// FiatBalanceCurrencyPaginated is the interface through which external methods can retrieve all Fiat account balances
// for a specific client.
func (p *postgresImpl) FiatBalanceCurrencyPaginated(clientID uuid.UUID, baseCurrency Currency, limit int32) (
	[]FiatAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	balance, err := p.Query.fiatGetAllAccounts(ctx, &fiatGetAllAccountsParams{
		ClientID: clientID,
		Currency: baseCurrency,
		Limit:    limit,
	})
	if err != nil {
		return []FiatAccount{}, ErrNotFound
	}

	return balance, nil
}

// FiatTransactionsCurrencyPaginated is the interface through which external methods can retrieve transactions on a Fiat
// account for a specific client during a specific month.
func (p *postgresImpl) FiatTransactionsCurrencyPaginated(
	clientID uuid.UUID,
	currency Currency,
	limit,
	offset int32,
	startTime,
	endTime pgtype.Timestamptz) ([]FiatJournal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	balance, err := p.Query.fiatGetAllJournalTransactionPaginated(ctx, &fiatGetAllJournalTransactionPaginatedParams{
		ClientID:  clientID,
		Currency:  currency,
		Offset:    offset,
		Limit:     limit,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		return []FiatJournal{}, ErrNotFound
	}

	return balance, nil
}
