package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
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
func (p *postgresImpl) FiatBalanceCurrency(clientID uuid.UUID, currency Currency) (*FiatAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	balance, err := p.Query.fiatGetAccount(ctx, &fiatGetAccountParams{ClientID: clientID, Currency: currency})
	if err != nil {
		return nil, ErrNotFound
	}

	return &balance, nil
}
