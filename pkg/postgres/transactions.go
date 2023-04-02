package postgres

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type FiatTransactionDetails struct {
	ClientID uuid.UUID       `json:"clientId"`
	Currency Currency        `json:"currency"`
	Amount   decimal.Decimal `json:"amount"`
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
	compare := bytes.Compare(lhs.ClientID.Bytes(), rhs.ClientID.Bytes())

	if (compare == 0 && lhs.Currency > rhs.Currency) || compare > 0 {
		return &rhs, &lhs
	}

	return &lhs, &rhs
}

type FiatAccountTransferResult struct {
	TxID     uuid.UUID          `json:"txId"`
	ClientID uuid.UUID          `json:"clientId"`
	TxTS     pgtype.Timestamptz `json:"txTimestamp"`
	Balance  decimal.Decimal    `json:"balance"`
	LastTx   decimal.Decimal    `json:"lastTx"`
	Currency Currency           `json:"currency"`
}

// FiatExternalTransfer will deposit inbound transfers into fiat accounts using a transaction block.
/*
  		Minimize the duration for which the transaction block will be active by performing as many operations as
   		possible outside the transaction.

    [1] Convert the transaction amount to a pgtype and truncate to two decimal places. This is to adjust for the
		floating point precision representational issues.
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
		journalRow FiatExternalTransferJournalEntryRow
		updateRow  FiatUpdateAccountBalanceRow
	)

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
		Amount:   xferDetails.Amount,
	}); err != nil {
		msg := "failed to post Fiat account Journal entries"
		p.logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	// Update the account balance.
	if updateRow, err = queryTx.FiatUpdateAccountBalance(ctx, &FiatUpdateAccountBalanceParams{
		ClientID: xferDetails.ClientID,
		Currency: xferDetails.Currency,
		LastTx:   xferDetails.Amount,
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

// FiatTransactionRowLockAndBalanceCheck will acquire row locks on the Fiat accounts in a deterministic lock order.
// It will then check to see if the balance of the source/debit account is sufficient for the transaction.
func fiatTransactionRowLockAndBalanceCheck(
	ctx context.Context,
	queryTx *Queries,
	src,
	dst *FiatTransactionDetails) error {
	// Order locks.
	lockFirst, lockSecond := src.Less(dst)

	// Row lock the accounts in order.
	balanceFirst, err := queryTx.FiatRowLockAccount(ctx, &FiatRowLockAccountParams{
		ClientID: (*lockFirst).ClientID,
		Currency: (*lockFirst).Currency,
	})
	if err != nil {
		return fmt.Errorf("failed to get row lock on first Fiat account %w", err)
	}

	balanceSecond, err := queryTx.FiatRowLockAccount(ctx, &FiatRowLockAccountParams{
		ClientID: (*lockSecond).ClientID,
		Currency: (*lockSecond).Currency,
	})
	if err != nil {
		return fmt.Errorf("failed to get row lock on second Fiat account %w", err)
	}

	// Check which lock operation returned the source/debit balance.
	debitBalance := &balanceFirst
	if *lockSecond == src {
		debitBalance = &balanceSecond
	}

	// Check for sufficient funds.
	if debitBalance.LessThan(src.Amount) {
		return fmt.Errorf("insufficient balance in source account: %s, %s", debitBalance, src.Amount)
	}

	return nil
}

// FiatInternalTransfer will perform transfers between fiat accounts using a transaction block.
/*
  		Minimize the duration for which the transaction block will be active by performing as many operations as
   		possible outside the transaction.

    [1] Convert the transaction amounts to a pgtype and truncate to two decimal places. This is to adjust for the
		floating point precision representational issues.
    [2] Acquire a row lock on the accounts without holding a lock on the foreign key for the Client ID.
        Their accounts will be compared against each other using a total order rule.
    [3] Make the Journal entries for both of the accounts.
    [4] Update the balance for both of the accounts.
*/
func (p *Postgres) FiatInternalTransfer(parentCtx context.Context, src, dst *FiatTransactionDetails) (
	*FiatAccountTransferResult, *FiatAccountTransferResult, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second) //nolint:gomnd

	defer cancel()

	var (
		err           error
		tx            pgx.Tx
		journalRow    FiatInternalTransferJournalEntryRow
		postCreditRow FiatUpdateAccountBalanceRow
		postDebitRow  FiatUpdateAccountBalanceRow
	)

	// Begin transaction.

	if tx, err = p.pool.Begin(ctx); err != nil {
		msg := "internal transfer Fiat transaction block setup failed"
		p.logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	// Set rollback in case of failure.
	defer func() {
		if errRollback := tx.Rollback(context.TODO()); errRollback != nil {
			// If the connection is closed the transaction was committed. Ignore the error from rollback in this case.
			if !errors.Is(errRollback, pgx.ErrTxClosed) {
				p.logger.Error("failed to rollback internal Fiat account transaction", zap.Error(errRollback))
			}
		}
	}()

	// Configure transaction query connection.
	queryTx := p.Query.WithTx(tx)

	// Row lock the accounts in order and check balances.
	if err = fiatTransactionRowLockAndBalanceCheck(ctx, queryTx, src, dst); err != nil {
		msg := "failed to get row lock on Fiat accounts and verify balance of debit account"
		p.logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	// Make General Journal ledger entries.
	if journalRow, err = queryTx.FiatInternalTransferJournalEntry(ctx, &FiatInternalTransferJournalEntryParams{
		DestinationAccount:  dst.ClientID,
		DestinationCurrency: dst.Currency,
		CreditAmount:        dst.Amount,
		SourceAccount:       src.ClientID,
		SourceCurrency:      src.Currency,
		DebitAmount:         src.Amount,
	}); err != nil {
		msg := "failed to post Fiat account Journal entries for internal transfer"
		p.logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	// Update the destination and then source account balances.
	if postCreditRow, err = queryTx.FiatUpdateAccountBalance(ctx, &FiatUpdateAccountBalanceParams{
		ClientID: dst.ClientID,
		Currency: dst.Currency,
		LastTx:   dst.Amount,
		LastTxTs: journalRow.TransactedAt,
	}); err != nil {
		msg := "failed to credit Fiat account balance for internal transfer"
		p.logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	if postDebitRow, err = queryTx.FiatUpdateAccountBalance(ctx, &FiatUpdateAccountBalanceParams{
		ClientID: src.ClientID,
		Currency: src.Currency,
		LastTx:   src.Amount,
		LastTxTs: journalRow.TransactedAt,
	}); err != nil {
		msg := "failed to debit Fiat account balance for internal transfer"
		p.logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	// Commit transaction.
	if err = tx.Commit(ctx); err != nil {
		msg := "failed to commit internal Fiat account transfer"
		p.logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	return &FiatAccountTransferResult{
			TxID:     journalRow.TxID,
			ClientID: src.ClientID,
			TxTS:     postDebitRow.LastTxTs,
			Balance:  postDebitRow.Balance,
			LastTx:   postDebitRow.LastTx,
			Currency: src.Currency,
		},
		&FiatAccountTransferResult{
			TxID:     journalRow.TxID,
			ClientID: dst.ClientID,
			TxTS:     postCreditRow.LastTxTs,
			Balance:  postCreditRow.Balance,
			LastTx:   postCreditRow.LastTx,
			Currency: dst.Currency,
		},
		nil
}
