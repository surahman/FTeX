package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type FiatAccountDetails struct {
	ClientID pgtype.UUID `json:"clientId"`
	Currency Currency    `json:"currency"`
}

type FiatAccountTransferResult struct {
	TxID     pgtype.UUID        `json:"txId"`
	ClientID pgtype.UUID        `json:"clientId"`
	TxTS     pgtype.Timestamptz `json:"txTimestamp"`
	Balance  pgtype.Numeric     `json:"balance"`
	LastTx   pgtype.Numeric     `json:"lastTx"`
	Currency Currency           `json:"currency"`
}

// FiatExternalTransfer will deposit inbound transfers into fiat accounts using a transaction block.
/*
  		Minimize the duration for which the transaction block will be active by performing as many operations as
   		possible outside the transaction.

    [1] Convert the transaction amount to a pgtype and truncate to two decimal places. For simplicity, the truncated
        excess amount will be ignored. Consider it as a transaction fee ;).
    [2] Acquire a row lock on the destination account without holding a lock on the foreign key for the Client ID.
        There will be no update for the external account balance so there is no need for a row lock on the account.
    [3] Make the Journal entries for the external and internal accounts.
    [4] Update the balance for the internal account.
*/
func (p *Postgres) FiatExternalTransfer(parentCtx context.Context, destination *FiatAccountDetails, amount float64) (
	FiatAccountTransferResult, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)

	defer cancel()

	var (
		err        error
		tx         pgx.Tx
		txAmount   pgtype.Numeric
		result     FiatAccountTransferResult
		journalRow fiatExternalTransferJournalEntryRow
		updateRow  fiatUpdateAccountBalanceRow
	)

	// Create transaction amount.
	if err = txAmount.Scan(fmt.Sprintf("%.2f", amount)); err != nil {
		msg := "failed to truncate the external Fiat transaction amount to two decimal places"
		p.logger.Warn(msg, zap.Error(err))

		return result, fmt.Errorf(msg+" %w", err)
	}

	// Begin transaction.

	if tx, err = p.pool.Begin(ctx); err != nil {
		msg := "external transfer Fiat transaction block setup failed"
		p.logger.Warn(msg, zap.Error(err))

		return result, fmt.Errorf(msg+" %w", err)
	}

	// Set rollback in case of failure.
	defer func() {
		if errRollback := tx.Rollback(context.TODO()); errRollback != nil {
			// If the connection is closed the transaction was committed. Ignore the error from rollback in this case.
			if !errors.Is(errRollback, pgx.ErrTxClosed) {
				p.logger.Error("failed to rollback external Fiat account transaction", zap.Error(errRollback))
			}
		}
	}()

	queryTx := p.Query.WithTx(tx)

	// Row lock the destination account.
	if _, err = queryTx.fiatRowLockAccount(ctx, &fiatRowLockAccountParams{
		ClientID: destination.ClientID,
		Currency: destination.Currency,
	}); err != nil {
		msg := "failed to get row lock on destination Fiat account"
		p.logger.Warn(msg, zap.Error(err))

		return result, fmt.Errorf(msg+" %w", err)
	}

	// Make General Journal ledger entries.
	if journalRow, err = queryTx.fiatExternalTransferJournalEntry(ctx, &fiatExternalTransferJournalEntryParams{
		ClientID: destination.ClientID,
		Currency: destination.Currency,
		Amount:   txAmount,
	}); err != nil {
		msg := "failed to post Fiat account Journal entries"
		p.logger.Warn(msg, zap.Error(err))

		return result, fmt.Errorf(msg+" %w", err)
	}

	// Update the account balance.
	if updateRow, err = queryTx.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: destination.ClientID,
		Currency: destination.Currency,
		LastTx:   txAmount,
		LastTxTs: journalRow.TransactedAt,
	}); err != nil {
		msg := "failed to update Fiat account balance"
		p.logger.Warn(msg, zap.Error(err))

		return result, fmt.Errorf(msg+" %w", err)
	}

	// Commit transaction.
	if err = tx.Commit(ctx); err != nil {
		msg := "failed to commit external Fiat account transfer"
		p.logger.Warn(msg, zap.Error(err))

		return result, fmt.Errorf(msg+" %w", err)
	}

	// Prepare result.
	result.ClientID = destination.ClientID
	result.Currency = destination.Currency
	result.Balance = updateRow.Balance
	result.TxID = journalRow.TxID
	result.LastTx = updateRow.LastTx
	result.TxTS = journalRow.TransactedAt

	return result, nil
}
