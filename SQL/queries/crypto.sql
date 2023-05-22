-- name: cryptoCreateAccount :execrows
-- cryptoCreateAccount inserts a fiat account record.
INSERT INTO crypto_accounts (client_id, ticker)
VALUES ($1, $2);

-- name: callPurchaseCrypto :exec
-- purchaseCrypto will execute a transaction to purchase a crypto currency.
CALL purchase_cryptocurrency($1,$2,$3, @fiat_debit_amount::numeric(18, 2), $4, @crypto_credit_amount::numeric(24, 8));
