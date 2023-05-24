-- name: cryptoCreateAccount :execrows
-- cryptoCreateAccount inserts a fiat account record.
INSERT INTO crypto_accounts (client_id, ticker)
VALUES ($1, $2);

-- name: cryptoPurchase :exec
-- cryptoPurchase will execute a transaction to purchase a Cryptocurrency using a Fiat currency.
CALL purchase_cryptocurrency($1,$2,$3, @fiat_debit_amount::numeric(18, 2), $4, @crypto_credit_amount::numeric(24, 8));

-- name: cryptoGetAccount :one
-- cryptoGetAccount will retrieve a specific user's account for a given cryptocurrency ticker.
SELECT *
FROM crypto_accounts
WHERE client_id=$1 AND ticker=$2;

-- name: cryptoGetJournalTransaction :many
-- cryptoGetJournalTransaction will retrieve the journal entries associated with a transaction.
SELECT *
FROM crypto_journal
WHERE client_id = $1 AND tx_id = $2;

-- name: cryptoSell :exec
-- cryptoSell will execute a transaction to sell a Cryptocurrency and purchase a Fiat currency.
CALL purchase_cryptocurrency($1,$2,$3, @fiat_credit_amount::numeric(18, 2), $4, @crypto_debit_amount::numeric(24, 8));
