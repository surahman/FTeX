
-- name: CreateFiatAccount :exec
INSERT INTO fiat_accounts (client_id, currency)
VALUES ($1, $2);

-- name: UpdateBalanceFiatAccount :one
UPDATE fiat_accounts
SET balance=balance + $3, last_tx=$3, last_tx_ts=now()
WHERE client_id=$1 AND currency=$2
RETURNING balance, last_tx, last_tx_ts;

-- name: GeneralLedgerEntryFiatAccount :one
INSERT INTO  fiat_general_ledger (client_id, currency, ammount, transacted_at)
VALUES ($1, $2, $3, $4)
RETURNING tx_id;

-- name: GeneralLedgerDepositFiatAccount :one
INSERT INTO fiat_general_ledger (
    client_id,
    currency,
    ammount,
    transacted_at,
    tx_id)
SELECT
    client_id,
    $1,
    $2,
    $3,
    $4
FROM
    users AS client_id
WHERE
    username = 'deposit-fiat'
RETURNING
    tx_id;
