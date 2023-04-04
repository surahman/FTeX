-- name: FiatCreateAccount :execrows
-- FiatCreateAccount inserts a fiat account record.
INSERT INTO fiat_accounts (client_id, currency)
VALUES ($1, $2);

-- name: FiatRowLockAccount :one
-- FiatRowLockAccount will acquire a row level lock without locks on the foreign keys.
SELECT balance
FROM fiat_accounts
WHERE client_id=$1 AND currency=$2
LIMIT 1
FOR NO KEY UPDATE;

-- name: FiatUpdateAccountBalance :one
-- FiatUpdateAccountBalance will add an amount to a fiat accounts balance.
UPDATE fiat_accounts
SET balance=round_half_even(balance + @Amount::numeric(18, 2), 2),
    last_tx=round_half_even(@Amount::numeric(18, 2), 2),
    last_tx_ts=$3
WHERE client_id=$1 AND currency=$2
RETURNING balance, last_tx, last_tx_ts;

-- name: FiatExternalTransferJournalEntry :one
-- FiatExternalTransferJournalEntry will create both journal entries for fiat accounts inbound deposits.
WITH deposit AS (
    INSERT INTO fiat_journal (
        client_id,
        currency,
        amount,
        transacted_at,
        tx_id)
    SELECT
        (   SELECT client_id
            FROM users
            WHERE username = 'deposit-fiat'),
        $2,
        -1 * sqlc.arg(amount)::numeric(18, 2),
        now(),
        gen_random_uuid()
    RETURNING tx_id, transacted_at
)
INSERT INTO fiat_journal (
    client_id,
    currency,
    amount,
    transacted_at,
    tx_id)
SELECT
    $1,
    $2,
    sqlc.arg(amount)::numeric(18, 2),
    (   SELECT transacted_at
        FROM deposit),
    (   SELECT tx_id
        FROM deposit)
RETURNING tx_id, transacted_at;

-- name: FiatInternalTransferJournalEntry :one
-- FiatInternalTransferJournalEntry will create both journal entries for fiat account internal transfers.
WITH deposit AS (
    INSERT INTO fiat_journal(
        client_id,
        currency,
        amount,
        transacted_at,
        tx_id)
    SELECT
        sqlc.arg(source_account)::uuid,
        sqlc.arg(source_currency)::currency,
        sqlc.arg(debit_amount)::numeric(18, 2),
        now(),
        gen_random_uuid()
    RETURNING tx_id, transacted_at
)
INSERT INTO fiat_journal (
    client_id,
    currency,
    amount,
    transacted_at,
    tx_id)
SELECT
    sqlc.arg(destination_account)::uuid,
    sqlc.arg(destination_currency)::currency,
    sqlc.arg(credit_amount)::numeric(18, 2),
    (   SELECT transacted_at
        FROM deposit),
    (   SELECT tx_id
        FROM deposit)
RETURNING tx_id, transacted_at;

-- name: FiatGetJournalTransaction :many
-- FiatGetJournalTransaction will retrieve the journal entries associated with a transaction.
SELECT *
FROM fiat_journal
WHERE tx_id = $1;

-- name: FiatGetJournalTransactionForAccount :many
-- FiatGetJournalTransactionForAccount will retrieve the journal entries associated with a specific account.
SELECT *
FROM fiat_journal
WHERE client_id = $1 AND currency = $2;

-- name: FiatGetJournalTransactionForAccountBetweenDates :many
-- FiatGetJournalTransactionForAccountBetweenDates will retrieve the journal entries associated with a specific account
-- in a date range.
SELECT *
FROM fiat_journal
WHERE client_id = $1
      AND currency = $2
      AND transacted_at
          BETWEEN sqlc.arg(start_time)::timestamptz
              AND sqlc.arg(end_time)::timestamptz;

-- name: FiatGetAccount :one
-- FiatGetAccount will retrieve a specific user's account for a given currency.
SELECT *
FROM fiat_accounts
WHERE client_id=$1 AND currency=$2;

-- name: FiatGetAllAccounts :many
-- FiatGetAllAccounts will retrieve all accounts associated with a specific user.
SELECT *
FROM fiat_accounts
WHERE client_id=$1;
