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
CALL sell_cryptocurrency($1,$2,$3, @fiat_credit_amount::numeric(18, 2), $4, @crypto_debit_amount::numeric(24, 8));

-- name: cryptoGetAllAccounts :many
-- cryptoGetAllAccounts will retrieve all accounts associated with a specific user.
SELECT *
FROM crypto_accounts
WHERE client_id=$1 AND ticker >= $2
ORDER BY ticker
LIMIT $3;

-- name: cryptoGetAllJournalTransactionsPaginated :many
-- cryptoGetAllJournalTransactionsPaginated will retrieve the journal entries associated with a specific account
-- in a date range.
SELECT *
FROM crypto_journal
WHERE client_id = $1
      AND ticker = $2
      AND transacted_at
          BETWEEN @start_time::timestamptz
              AND @end_time::timestamptz
ORDER BY transacted_at DESC
OFFSET $3
LIMIT $4;
