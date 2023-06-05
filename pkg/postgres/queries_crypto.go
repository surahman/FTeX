package postgres

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
//
//nolint:dupl
func (p *postgresImpl) CryptoPurchase(
	clientID uuid.UUID,
	fiatCurrency Currency,
	fiatDebitAmount decimal.Decimal,
	cryptoTicker string,
	cryptoCreditAmount decimal.Decimal) (*FiatJournal, *CryptoJournal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	txID, err := uuid.NewV4()
	if err != nil {
		p.logger.Error("failed to generate transaction id for Crypto purchase", zap.Error(err))

		return nil, nil, ErrTransactCrypto
	}

	err = p.Query.cryptoPurchase(ctx, &cryptoPurchaseParams{
		TransactionID:      txID,
		ClientID:           clientID,
		FiatCurrency:       fiatCurrency,
		CryptoTicker:       cryptoTicker,
		FiatDebitAmount:    fiatDebitAmount,
		CryptoCreditAmount: cryptoCreditAmount,
	})
	if err != nil {
		return nil, nil, ErrTransactCrypto
	}

	fiatJournal, err := p.Query.fiatGetJournalTransaction(ctx, &fiatGetJournalTransactionParams{
		ClientID: clientID,
		TxID:     txID,
	})
	if err != nil {
		p.logger.Error("failed to retrieve Fiat transaction details post Crypto purchase", zap.Error(err))

		return nil, nil, ErrTransactCryptoDetails
	}

	cryptoJournal, err := p.Query.cryptoGetJournalTransaction(ctx, &cryptoGetJournalTransactionParams{
		ClientID: clientID,
		TxID:     txID,
	})
	if err != nil {
		p.logger.Error("failed to retrieve Crypto transaction details post Crypto purchase", zap.Error(err))

		return nil, nil, ErrTransactCryptoDetails
	}

	return &fiatJournal[0], &cryptoJournal[0], nil
}

// CryptoSell is the interface through which external methods can sell a specific Cryptocurrency.
//
//nolint:dupl
func (p *postgresImpl) CryptoSell(
	clientID uuid.UUID,
	fiatCurrency Currency,
	fiatCreditAmount decimal.Decimal,
	cryptoTicker string,
	cryptoDebitAmount decimal.Decimal) (*FiatJournal, *CryptoJournal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	txID, err := uuid.NewV4()
	if err != nil {
		p.logger.Error("failed to generate transaction id for Crypto sale", zap.Error(err))

		return nil, nil, ErrTransactCrypto
	}

	err = p.Query.cryptoSell(ctx, &cryptoSellParams{
		TransactionID:     txID,
		ClientID:          clientID,
		FiatCurrency:      fiatCurrency,
		CryptoTicker:      cryptoTicker,
		FiatCreditAmount:  fiatCreditAmount,
		CryptoDebitAmount: cryptoDebitAmount,
	})
	if err != nil {
		return nil, nil, ErrTransactCrypto
	}

	fiatJournal, err := p.Query.fiatGetJournalTransaction(ctx, &fiatGetJournalTransactionParams{
		ClientID: clientID,
		TxID:     txID,
	})
	if err != nil {
		p.logger.Error("failed to retrieve Fiat transaction details post Crypto sale", zap.Error(err))

		return nil, nil, ErrTransactCryptoDetails
	}

	cryptoJournal, err := p.Query.cryptoGetJournalTransaction(ctx, &cryptoGetJournalTransactionParams{
		ClientID: clientID,
		TxID:     txID,
	})
	if err != nil {
		p.logger.Error("failed to retrieve Crypto transaction details post Crypto sale", zap.Error(err))

		return nil, nil, ErrTransactCryptoDetails
	}

	return &fiatJournal[0], &cryptoJournal[0], nil
}

// CryptoBalancesPaginated is the interface through which external methods can retrieve all Crypto account balances
// for a specific client.
func (p *postgresImpl) CryptoBalancesPaginated(clientID uuid.UUID, ticker string, limit int32) (
	[]CryptoAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	balance, err := p.Query.cryptoGetAllAccounts(ctx, &cryptoGetAllAccountsParams{
		ClientID: clientID,
		Ticker:   ticker,
		Limit:    limit,
	})
	if err != nil {
		return []CryptoAccount{}, ErrNotFound
	}

	return balance, nil
}

// CryptoTransactionsPaginated is the interface through which external methods can retrieve transactions on a Crypto
// account for a specific client during a specific month.
func (p *postgresImpl) CryptoTransactionsPaginated(
	clientID uuid.UUID,
	ticker string,
	limit,
	offset int32,
	startTime,
	endTime pgtype.Timestamptz) ([]CryptoJournal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) //nolint:gomnd

	defer cancel()

	balance, err := p.Query.cryptoGetAllJournalTransactionsPaginated(ctx, &cryptoGetAllJournalTransactionsPaginatedParams{
		ClientID:  clientID,
		Ticker:    ticker,
		Offset:    offset,
		Limit:     limit,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		return []CryptoJournal{}, ErrNotFound
	}

	return balance, nil
}
