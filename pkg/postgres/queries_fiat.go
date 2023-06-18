package postgres

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/surahman/FTeX/pkg/constants"
	"go.uber.org/zap"
)

// FiatCreateAccount is the interface through which external methods can create a Fiat account.
func (p *postgresImpl) FiatCreateAccount(clientID uuid.UUID, currency Currency) error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	rowsAffected, err := p.Query.fiatCreateAccount(ctx, &fiatCreateAccountParams{ClientID: clientID, Currency: currency})
	if err != nil || rowsAffected != int64(1) {
		p.logger.Error("failed to create fiat account", zap.Error(err))

		return ErrCreateFiat
	}

	return nil
}

// FiatBalance is the interface through which external methods can retrieve a Fiat-account balance for a specific
// currency.
func (p *postgresImpl) FiatBalance(clientID uuid.UUID, currency Currency) (FiatAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	balance, err := p.Query.fiatGetAccount(ctx, &fiatGetAccountParams{ClientID: clientID, Currency: currency})
	if err != nil {
		return FiatAccount{}, ErrNotFound
	}

	return balance, nil
}

// FiatTxDetails is the interface through which external methods can retrieve a Fiat transaction details for a specific
// transaction.
func (p *postgresImpl) FiatTxDetails(clientID uuid.UUID, transactionID uuid.UUID) ([]FiatJournal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	journal, err := p.Query.fiatGetJournalTransaction(ctx, &fiatGetJournalTransactionParams{clientID, transactionID})
	if err != nil {
		return nil, ErrNotFound
	}

	return journal, nil
}

// FiatBalancePaginated is the interface through which external methods can retrieve all Fiat account balances for a
// specific client.
func (p *postgresImpl) FiatBalancePaginated(clientID uuid.UUID, baseCurrency Currency, limit int32) (
	[]FiatAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

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

// FiatTransactionsPaginated is the interface through which external methods can retrieve transactions on a Fiat
// account for a specific client during a specific month.
func (p *postgresImpl) FiatTransactionsPaginated(
	clientID uuid.UUID,
	currency Currency,
	limit,
	offset int32,
	startTime,
	endTime pgtype.Timestamptz) ([]FiatJournal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ThreeSeconds())

	defer cancel()

	balance, err := p.Query.fiatGetAllJournalTransactionsPaginated(ctx, &fiatGetAllJournalTransactionsPaginatedParams{
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
