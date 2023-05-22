package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// CryptoCreateAccount is the interface through which external methods can create a Crypto account.
func (p *postgresImpl) CryptoCreateAccount(clientID uuid.UUID, ticker string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	rowsAffected, err := p.Query.cryptoCreateAccount(ctx, &cryptoCreateAccountParams{ClientID: clientID, Ticker: ticker})
	if err != nil || rowsAffected != int64(1) {
		p.logger.Error("failed to create Crypto account", zap.Error(err))

		return ErrCreateFiat
	}

	return nil
}

// CryptoBalanceCurrency is the interface through which external methods can retrieve a Crypto account balance for a
// specific cryptocurrency.
func (p *postgresImpl) CryptoBalanceCurrency(clientID uuid.UUID, ticker string) (CryptoAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	balance, err := p.Query.cryptoGetAccount(ctx, &cryptoGetAccountParams{clientID, ticker})
	if err != nil {
		return CryptoAccount{}, ErrNotFound
	}

	return balance, nil
}

// CryptoTxDetailsCurrency is the interface through which external methods can retrieve a Crypto transaction details for
// a specific transaction.
func (p *postgresImpl) CryptoTxDetailsCurrency(clientID uuid.UUID, transactionID uuid.UUID) ([]CryptoJournal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	journal, err := p.Query.cryptoGetJournalTransaction(ctx, &cryptoGetJournalTransactionParams{clientID, transactionID})
	if err != nil {
		return nil, ErrNotFound
	}

	return journal, nil
}

// CryptoPurchase is the interface through which external methods can purchase a specific Cryptocurrency.
func (p *postgresImpl) CryptoPurchase(
	clientID uuid.UUID,
	fiatCurrency Currency,
	fiatDebitAmount decimal.Decimal,
	cryptoTicker string,
	cryptoCreditAmount decimal.Decimal) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	txID, err := uuid.NewV4()
	if err != nil {
		p.logger.Error("failed to generate transaction id for Crypto purchase", zap.Error(err))

		return uuid.UUID{}, ErrTransactCrypto
	}

	err = p.Query.cryptoPurchase(ctx, &cryptoPurchaseParams{
		TransactionID:      uuid.UUID{},
		ClientID:           clientID,
		FiatCurrency:       fiatCurrency,
		CryptoTicker:       cryptoTicker,
		FiatDebitAmount:    fiatDebitAmount,
		CryptoCreditAmount: cryptoCreditAmount,
	})
	if err != nil {
		return uuid.UUID{}, ErrTransactCrypto
	}

	return txID, nil
}
