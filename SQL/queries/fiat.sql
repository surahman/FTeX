-- name: fiatCreateAccount :execrows
-- fiatCreateAccount inserts a fiat account record.
INSERT INTO fiat_accounts (client_id, currency)
VALUES ($1, $2);

-- name: fiatRowLockAccount :one
-- fiatRowLockAccount will acquire a row level lock without locks on the foreign keys.
SELECT balance
FROM fiat_accounts
WHERE client_id=$1 AND currency=$2
LIMIT 1
FOR NO KEY UPDATE;

-- name: fiatUpdateAccountBalance :one
-- fiatUpdateAccountBalance will add an amount to a fiat accounts balance.
UPDATE fiat_accounts
SET balance=round_half_even(balance + @Amount::numeric(18, 2), 2),
    last_tx=round_half_even(@Amount::numeric(18, 2), 2),
    last_tx_ts=$3
WHERE client_id=$1 AND currency=$2
RETURNING balance, last_tx, last_tx_ts;

-- name: fiatExternalTransferJournalEntry :one
-- fiatExternalTransferJournalEntry will create both journal entries for fiat accounts inbound deposits.
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
        round_half_even(-1 * @amount::numeric(18, 2), 2),
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
    round_half_even(@amount::numeric(18, 2), 2),
    (   SELECT transacted_at
        FROM deposit),
    (   SELECT tx_id
        FROM deposit)
RETURNING tx_id, transacted_at;

-- name: fiatInternalTransferJournalEntry :one
-- fiatInternalTransferJournalEntry will create both journal entries for fiat account internal transfers.
WITH deposit AS (
    INSERT INTO fiat_journal(
        client_id,
        currency,
        amount,
        transacted_at,
        tx_id)
    SELECT
        @source_account::uuid,
        @source_currency::currency,
        round_half_even(@debit_amount::numeric(18, 2), 2),
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
    @destination_account::uuid,
    @destination_currency::currency,
    round_half_even(@credit_amount::numeric(18, 2), 2),
    (   SELECT transacted_at
        FROM deposit),
    (   SELECT tx_id
        FROM deposit)
RETURNING tx_id, transacted_at;

-- name: fiatGetJournalTransaction :many
-- fiatGetJournalTransaction will retrieve the journal entries associated with a transaction.
SELECT *
FROM fiat_journal
WHERE client_id = $1 AND tx_id = $2;

-- name: fiatGetJournalTransactionForAccount :many
-- fiatGetJournalTransactionForAccount will retrieve the journal entries associated with a specific account.
SELECT *
FROM fiat_journal
WHERE client_id = $1 AND currency = $2;

-- name: fiatGetJournalTransactionForAccountBetweenDates :many
-- fiatGetJournalTransactionForAccountBetweenDates will retrieve the journal entries associated with a specific account
-- in a date range.
SELECT *
FROM fiat_journal
WHERE client_id = $1
      AND currency = $2
      AND transacted_at
          BETWEEN @start_time::timestamptz
              AND @end_time::timestamptz;

-- name: fiatGetAccount :one
-- fiatGetAccount will retrieve a specific user's account for a given currency.
SELECT *
FROM fiat_accounts
WHERE client_id=$1 AND currency=$2;

-- name: fiatGetAllAccounts :many
-- fiatGetAllAccounts will retrieve all accounts associated with a specific user.
SELECT *
FROM fiat_accounts
WHERE client_id=$1 AND currency >= $2
ORDER BY currency
LIMIT $3;
