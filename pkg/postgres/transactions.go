package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/surahman/FTeX/pkg/utilities"
	"go.uber.org/zap"
)

type FiatTransactionDetails struct {
	ClientID uuid.UUID `json:"clientId"`
	Currency Currency  `json:"currency"`
	Amount   float64   `json:"amount"`
}

// Less returns a total ordering on two FiatTransactionDetails structs.
/*	IF
 [1] 	LHS UUID is equal to the RHS UUID
		AND
		LHS Currency is greater than the RHS Currency
 [2] 	OR
		LHS UUID is greater than the RHS UUID
	RETURN RHS and LHS
		ELSE
	RETURN LHS and RHS
*/
func (lhs *FiatTransactionDetails) Less(rhs *FiatTransactionDetails) (
	**FiatTransactionDetails, **FiatTransactionDetails) {
	lhsUUID := lhs.ClientID.String()
	rhsUUID := rhs.ClientID.String()

	if (lhsUUID == rhsUUID && lhs.Currency > rhs.Currency) || (lhsUUID > rhsUUID) {
		return &rhs, &lhs
	}

	return &lhs, &rhs
}

type FiatAccountTransferResult struct {
	TxID     uuid.UUID          `json:"txId"`
	ClientID uuid.UUID          `json:"clientId"`
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
func (p *Postgres) FiatExternalTransfer(parentCtx context.Context, xferDetails *FiatTransactionDetails) (
	*FiatAccountTransferResult, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second) //nolint:gomnd

	defer cancel()

	var (
		err        error
		tx         pgx.Tx
		txAmount   pgtype.Numeric
		journalRow FiatExternalTransferJournalEntryRow
		updateRow  FiatUpdateAccountBalanceRow
	)

	// Create transaction amount.
	if err = txAmount.Scan(utilities.Float64TwoDecimalPlacesString(xferDetails.Amount)); err != nil {
		msg := "failed to truncate the external Fiat transaction amount to two decimal places"
		p.logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	// Begin transaction.

	if tx, err = p.pool.Begin(ctx); err != nil {
		msg := "external transfer Fiat transaction block setup failed"
		p.logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
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
	if _, err = queryTx.FiatRowLockAccount(ctx, &FiatRowLockAccountParams{
		ClientID: xferDetails.ClientID,
		Currency: xferDetails.Currency,
	}); err != nil {
		msg := "failed to get row lock on destination Fiat account"
		p.logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	// Make General Journal ledger entries.
	if journalRow, err = queryTx.FiatExternalTransferJournalEntry(ctx, &FiatExternalTransferJournalEntryParams{
		ClientID: xferDetails.ClientID,
		Currency: xferDetails.Currency,
		Amount:   txAmount,
	}); err != nil {
		msg := "failed to post Fiat account Journal entries"
		p.logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	// Update the account balance.
	if updateRow, err = queryTx.FiatUpdateAccountBalance(ctx, &FiatUpdateAccountBalanceParams{
		ClientID: xferDetails.ClientID,
		Currency: xferDetails.Currency,
		LastTx:   txAmount,
		LastTxTs: journalRow.TransactedAt,
	}); err != nil {
		msg := "failed to update Fiat account balance"
		p.logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	// Commit transaction.
	if err = tx.Commit(ctx); err != nil {
		msg := "failed to commit external Fiat account transfer"
		p.logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	return &FiatAccountTransferResult{
			TxID:     journalRow.TxID,
			ClientID: xferDetails.ClientID,
			TxTS:     journalRow.TransactedAt,
			Balance:  updateRow.Balance,
			LastTx:   updateRow.LastTx,
			Currency: xferDetails.Currency,
		},
		nil
}
