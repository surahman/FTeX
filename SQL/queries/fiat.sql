-- name: createFiatAccount :exec
-- createFiatAccount inserts a fiat account record.
INSERT INTO fiat_accounts (client_id, currency)
VALUES ($1, $2);

-- name: rowLockFiatAccount :exec
-- rowLockFiatAccount will acquire a row level lock without locks on the foreign keys.
SELECT
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

-- name: generalLedgerEntryFiatAccount :one
-- generalLedgerEntryFiatAccount will create general ledger entry.
-- [$1] is the Client ID. If <deposit> is specified the <deposit-fiat> Client ID will be looked up.
-- [$5] is the TX ID. A random one will be generated if not supplied.
INSERT INTO  fiat_general_ledger (
    client_id,
    currency,
    ammount,
    transacted_at,
    tx_id)
SELECT
    CASE
      WHEN sqlc.arg(client_id_str)::text='deposit' THEN (
        SELECT client_id
        FROM users
        WHERE username = 'deposit-fiat')
      ELSE $1
    END AS client_id,
    $2,
    $3,
    $4,
    CASE
      WHEN length(sqlc.arg(tx_id_str)::text)=0 THEN gen_random_uuid()
      ELSE $5
    END AS tx_id
RETURNING tx_id;
