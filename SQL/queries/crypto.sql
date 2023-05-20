-- name: cryptoCreateAccount :execrows
-- cryptoCreateAccount inserts a fiat account record.
INSERT INTO crypto_accounts (client_id, ticker)
VALUES ($1, $2);
