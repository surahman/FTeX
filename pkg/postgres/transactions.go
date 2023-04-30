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
	"github.com/surahman/FTeX/pkg/logger"
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

// FiatExternalTransfer controls the transaction block that the external Fiat transfer transaction executes in.
func (p *postgresImpl) FiatExternalTransfer(parentCtx context.Context, xferDetails *FiatTransactionDetails) (
	*FiatAccountTransferResult, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second) //nolint:gomnd

	defer cancel()

	var (
		err       error
		tx        pgx.Tx
		txReceipt *FiatAccountTransferResult
	)

	// Begin transaction.
	if tx, err = p.pool.Begin(ctx); err != nil {
		p.logger.Warn("external transfer Fiat transaction block setup failed", zap.Error(err))

		return nil, ErrTransactFiat
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

	// Configure transaction query connection.
	queryTx := p.queries.WithTx(tx)

	// Handoff to external fiat transaction core logic.
	if txReceipt, err = fiatExternalTransfer(ctx, p.logger, queryTx, xferDetails); err != nil {
		p.logger.Warn("failed to complete external Fiat transfer transaction", zap.Error(err))

		return nil, ErrTransactFiat
	}

	// Commit transaction.
	if err = tx.Commit(ctx); err != nil {
		p.logger.Warn("failed to commit external Fiat account transfer", zap.Error(err))

		return nil, ErrTransactFiat
	}

	return txReceipt, nil
}

// fiatExternalTransfer will execute the logic to complete the external Fiat transfer transaction.
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
func fiatExternalTransfer(
	ctx context.Context,
	logger *logger.Logger,
	queryTx Querier,
	xferDetails *FiatTransactionDetails) (*FiatAccountTransferResult, error) {
	var (
		err        error
		journalRow fiatExternalTransferJournalEntryRow
		updateRow  fiatUpdateAccountBalanceRow
	)

	// Row lock the destination account.
	if _, err = queryTx.fiatRowLockAccount(ctx, &fiatRowLockAccountParams{
		ClientID: xferDetails.ClientID,
		Currency: xferDetails.Currency,
	}); err != nil {
		msg := "failed to get row lock on destination Fiat account"
		logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	// Make General Journal ledger entries.
	if journalRow, err = queryTx.fiatExternalTransferJournalEntry(ctx, &fiatExternalTransferJournalEntryParams{
		ClientID: xferDetails.ClientID,
		Currency: xferDetails.Currency,
		Amount:   xferDetails.Amount,
	}); err != nil {
		msg := "failed to post Fiat account Journal entries"
		logger.Warn(msg, zap.Error(err))

		return nil, fmt.Errorf(msg+" %w", err)
	}

	// Update the account balance.
	if updateRow, err = queryTx.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: xferDetails.ClientID,
		Currency: xferDetails.Currency,
		Amount:   xferDetails.Amount,
		LastTxTs: journalRow.TransactedAt,
	}); err != nil {
		msg := "failed to update Fiat account balance"
		logger.Warn(msg, zap.Error(err))

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

// fiatTransactionRowLockAndBalanceCheck will acquire row locks on the Fiat accounts in a deterministic lock order.
// It will then check to see if the balance of the source/debit account is sufficient for the transaction.
func fiatTransactionRowLockAndBalanceCheck(
	ctx context.Context,
	queryTx Querier,
	src,
	dst *FiatTransactionDetails) error {
	// Check for negative values.
	if src.Amount.IsNegative() || dst.Amount.IsNegative() {
		return fmt.Errorf("amounts contains negative value")
	}

	// Order locks.
	lockFirst, lockSecond := src.Less(dst)

	// Row lock the accounts in order.
	balanceFirst, err := queryTx.fiatRowLockAccount(ctx, &fiatRowLockAccountParams{
		ClientID: (*lockFirst).ClientID,
		Currency: (*lockFirst).Currency,
	})
	if err != nil {
		return fmt.Errorf("failed to get row lock on first Fiat account %w", err)
	}

	balanceSecond, err := queryTx.fiatRowLockAccount(ctx, &fiatRowLockAccountParams{
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

// FiatInternalTransfer controls the transaction block that the internal Fiat transfer transaction executes in.
func (p *postgresImpl) FiatInternalTransfer(
	parentCtx context.Context,
	src,
	dst *FiatTransactionDetails) (*FiatAccountTransferResult, *FiatAccountTransferResult, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 3*time.Second) //nolint:gomnd

	defer cancel()

	var (
		err          error
		tx           pgx.Tx
		dstTxReceipt *FiatAccountTransferResult
		srcTxReceipt *FiatAccountTransferResult
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
	queryTx := p.queries.WithTx(tx)

	// Handoff to internal fiat transaction core logic.
	if srcTxReceipt, dstTxReceipt, err = fiatInternalTransfer(ctx, p.logger, queryTx, src, dst); err != nil {
		msg := "failed to complete internal Fiat transfer transaction"
		p.logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	// Commit transaction.
	if err = tx.Commit(ctx); err != nil {
		msg := "failed to commit internal Fiat account transfer"
		p.logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	return srcTxReceipt, dstTxReceipt, nil
}

// fiatInternalTransfer will execute the logic to complete the internal Fiat transfer transaction.
/*
  		Minimize the duration for which the transaction block will be active by performing as many operations as
   		possible outside the transaction.

    [1] Convert the transaction amounts to a pgtype and truncate to two decimal places. This is to adjust for the
		floating point precision representational issues.
    [2] Acquire a row lock on the accounts without holding a lock on the foreign key for the Client ID.
        Their accounts will be compared against each other using a total order rule.
    [3] Make the Journal entries for both of the accounts.
    [4] Update the balance for the source and destination accounts.
*/
func fiatInternalTransfer(
	ctx context.Context,
	logger *logger.Logger,
	queryTx Querier,
	src,
	dst *FiatTransactionDetails) (*FiatAccountTransferResult, *FiatAccountTransferResult, error) {
	var (
		err           error
		journalRow    fiatInternalTransferJournalEntryRow
		postCreditRow fiatUpdateAccountBalanceRow
		postDebitRow  fiatUpdateAccountBalanceRow
	)

	// Row lock the accounts in order and check balances.
	if err = fiatTransactionRowLockAndBalanceCheck(ctx, queryTx, src, dst); err != nil {
		msg := "failed to get row lock on Fiat accounts and verify balance of debit account"
		logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	// Make General Journal ledger entries.
	if journalRow, err = queryTx.fiatInternalTransferJournalEntry(ctx, &fiatInternalTransferJournalEntryParams{
		DestinationAccount:  dst.ClientID,
		DestinationCurrency: dst.Currency,
		CreditAmount:        dst.Amount,
		SourceAccount:       src.ClientID,
		SourceCurrency:      src.Currency,
		DebitAmount:         src.Amount,
	}); err != nil {
		msg := "failed to post Fiat account Journal entries for internal transfer"
		logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	// Update the destination and then source account balances.
	if postCreditRow, err = queryTx.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: dst.ClientID,
		Currency: dst.Currency,
		Amount:   dst.Amount,
		LastTxTs: journalRow.TransactedAt,
	}); err != nil {
		msg := "failed to credit Fiat account balance for internal transfer"
		logger.Warn(msg, zap.Error(err))

		return nil, nil, fmt.Errorf(msg+" %w", err)
	}

	if postDebitRow, err = queryTx.fiatUpdateAccountBalance(ctx, &fiatUpdateAccountBalanceParams{
		ClientID: src.ClientID,
		Currency: src.Currency,
		Amount:   src.Amount.Mul(decimal.NewFromFloat(-1.0)),
		LastTxTs: journalRow.TransactedAt,
	}); err != nil {
		msg := "failed to debit Fiat account balance for internal transfer"
		logger.Warn(msg, zap.Error(err))

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
