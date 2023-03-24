-- name: createFiatAccount :execrows
-- createFiatAccount inserts a fiat account record.
INSERT INTO fiat_accounts (client_id, currency)
VALUES ($1, $2);

-- name: rowLockFiatAccount :one
-- rowLockFiatAccount will acquire a row level lock without locks on the foreign keys.
SELECT balance
FROM fiat_accounts
WHERE client_id=$1 AND currency=$2
LIMIT 1
FOR NO KEY UPDATE;

-- name: updateBalanceFiatAccount :one
-- updateBalanceFiatAccount will add an amount to a fiat accounts balance.
UPDATE fiat_accounts
SET balance=balance + $3, last_tx=$3, last_tx_ts=now()
WHERE client_id=$1 AND currency=$2
RETURNING balance, last_tx, last_tx_ts;

-- name: generalLedgerExternalFiatAccount :one
-- generalLedgerExternalFiatAccount will create both general ledger entries for fiat accounts inbound deposits.
WITH deposit AS (
    INSERT INTO fiat_general_ledger (
        client_id,
        currency,
        ammount,
        transacted_at,
        tx_id)
    SELECT
        (   SELECT client_id
            FROM users
            WHERE username = 'deposit-fiat'),
        $2,
        -1 * $3,
        now(),
        gen_random_uuid()
    RETURNING tx_id, transacted_at
)
INSERT INTO  fiat_general_ledger (
    client_id,
    currency,
    ammount,
    transacted_at,
    tx_id)
SELECT
    $1,
    $2,
    $3,
    (   SELECT transacted_at
        FROM deposit),
    (   SELECT tx_id
        FROM deposit)
RETURNING tx_id;

-- name: generalLedgerInternalFiatAccount :one
-- generalLedgerEntriesInternalAccount will create both general ledger entries for fiat accounts internal transfers.
WITH deposit AS (
    INSERT INTO fiat_general_ledger (
        client_id,
        currency,
        ammount,
        transacted_at,
        tx_id)
    SELECT
        sqlc.arg(source_account)::uuid,
        sqlc.arg(debit_amount)::numeric,
        $1,
        now(),
        gen_random_uuid()
    RETURNING tx_id, transacted_at
)
INSERT INTO  fiat_general_ledger (
    client_id,
    currency,
    ammount,
    transacted_at,
    tx_id)
SELECT
    sqlc.arg(destination_account)::uuid,
    $1,
    sqlc.arg(credit_amount)::numeric,
    (   SELECT transacted_at
        FROM deposit),
    (   SELECT tx_id
        FROM deposit)
RETURNING tx_id;
