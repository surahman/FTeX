--liquibase formatted sql

--changeset surahman:1
--preconditions onFail:HALT onError:HALT
--comment: Users table contains the general user information and login credentials.
CREATE TABLE IF NOT EXISTS users (
    first_name VARCHAR(64)          NOT NULL,
    last_name  VARCHAR(64)          NOT NULL,
    email      VARCHAR(64)          NOT NULL,
    username   VARCHAR(32)          UNIQUE NOT NULL,
    password   VARCHAR(128)         NOT NULL,
    client_id  UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
    is_deleted BOOLEAN              DEFAULT false NOT NULL
);
--rollback DROP TABLE users;

--changeset surahman:2
--preconditions onFail:HALT onError:HALT
--comment: Enum type for currency ISO codes.
CREATE TYPE currency AS ENUM (
'AED','AFN','ALL','AMD','ANG','AOA','ARS','AUD','AWG','AZN','BAM','BBD','BDT','BGN','BHD','BIF','BMD','BND','BOB',
'BRL','BSD','BTN','BWP','BYN','BZD','CAD','CDF','CHF','CLP','CNY','COP','CRC','CUC','CUP','CVE','CZK','DJF','DKK',
'DOP','DZD','EGP','ERN','ETB','EUR','FJD','FKP','GBP','GEL','GGP','GHS','GIP','GMD','GNF','GTQ','GYD','HKD','HNL',
'HRK','HTG','HUF','IDR','ILS','IMP','INR','IQD','IRR','ISK','JEP','JMD','JOD','JPY','KES','KGS','KHR','KMF','KPW',
'KRW','KWD','KYD','KZT','LAK','LBP','LKR','LRD','LSL','LYD','MAD','MDL','MGA','MKD','MMK','MNT','MOP','MRU','MUR',
'MVR','MWK','MXN','MYR','MZN','NAD','NGN','NIO','NOK','NPR','NZD','OMR','PAB','PEN','PGK','PHP','PKR','PLN','PYG',
'QAR','RON','RSD','RUB','RWF','SAR','SBD','SCR','SDG','SEK','SGD','SHP','SLL','SOS','SPL','SRD','STN','SVC','SYP',
'SZL','THB','TJS','TMT','TND','TOP','TRY','TTD','TVD','TWD','TZS','UAH','UGX','USD','UYU','UZS','VEF','VND','VUV',
'WST','XAF','XCD','XDR','XOF','XPF','YER','ZAR','ZMW','ZWD',
'FIAT', 'CRYPTO');
--rollback DROP TYPE currency;

--changeset surahman:3
--preconditions onFail:HALT onError:HALT
--comment: Fiat currency accounts.
CREATE TABLE IF NOT EXISTS fiat_accounts (
    currency        CURRENCY        DEFAULT 'USD' NOT NULL,
    balance         NUMERIC(18,2)   DEFAULT 0 NOT NULL,
    last_tx         NUMERIC(18,2)   DEFAULT 0 NOT NULL,
    last_tx_ts      TIMESTAMPTZ     DEFAULT now() NOT NULL,
    created_at      TIMESTAMPTZ     DEFAULT now() NOT NULL,
    client_id       UUID            REFERENCES users(client_id) ON DELETE CASCADE,
    PRIMARY KEY (client_id, currency)
);

CREATE INDEX IF NOT EXISTS fiat_client_id_idx ON fiat_accounts USING btree (client_id);
--rollback DROP TABLE fiat_accounts CASCADE;

--changeset surahman:4
--preconditions onFail:HALT onError:HALT
--comment: Create Fiat currency operations user and account.
INSERT INTO users (
    first_name,
    last_name,
    email,
    username,
    password,
    is_deleted)
SELECT
   'Internal',
   'FTeX, Inc.',
   'fiat@ftex.com',
   'fiat-currencies',
   password,
   true
FROM
    substr(md5(random()::text), 0, 32) AS password;

INSERT INTO fiat_accounts (
    currency,
    client_id)
SELECT
   'FIAT',
   client_id
FROM
    users AS client_id
WHERE
    username = 'fiat-currencies';
--rollback DELETE FROM users WHERE username='fiat-currencies';

--changeset surahman:5
--preconditions onFail:HALT onError:HALT
--comment: Fiat currency accounts transactions journal.
CREATE TABLE IF NOT EXISTS fiat_journal (
    currency        CURRENCY        NOT NULL,
    amount          NUMERIC(18,2)   NOT NULL,
    transacted_at   TIMESTAMPTZ     NOT NULL,
    client_id       UUID            REFERENCES users(client_id) ON DELETE CASCADE,
    tx_id           UUID            DEFAULT gen_random_uuid() NOT NULL,
    PRIMARY KEY(tx_id, client_id, currency)
);

CREATE INDEX IF NOT EXISTS fiat_journal_transacted_at_idx ON fiat_journal USING btree (transacted_at);
CREATE INDEX IF NOT EXISTS fiat_journal_tx_idx ON fiat_journal USING btree (tx_id);
--rollback DROP TABLE fiat_journal CASCADE;

--changeset surahman:6
--preconditions onFail:HALT onError:HALT
--comment: Rounds a number with arbitrary precision to a specified scale using the the Round Half-Even/Bankers' Algorithm.
CREATE OR REPLACE FUNCTION round_half_even(num NUMERIC, scale INTEGER)
RETURNS NUMERIC
LANGUAGE plpgsql
    IMMUTABLE
    STRICT
    PARALLEL SAFE
AS '
    DECLARE
        rounded     NUMERIC;
        difference  NUMERIC;
        multiplier  NUMERIC;
    BEGIN
        -- Check to see if rounding is needed.
        IF SCALE(num) <= scale THEN
            RETURN num;
        END IF;

        multiplier := (10::NUMERIC ^ scale);
        rounded    := round(num, scale);
        difference := rounded - num;

        -- IF half-way between two integers AND even THEN round-down:
        IF ABS(difference) * multiplier = 0.5::NUMERIC AND
            (rounded * multiplier) % 2::NUMERIC != 0::NUMERIC
        THEN
            rounded := round(num - difference, scale);
        END IF;

        RETURN rounded;

    END;
';
--rollback DROP FUNCTION round_half_even;

--changeset surahman:7
--preconditions onFail:HALT onError:HALT
--comment: Cryptocurrency accounts.
CREATE TABLE IF NOT EXISTS crypto_accounts (
    ticker          VARCHAR(6)      NOT NULL,
    balance         NUMERIC(24,8)   DEFAULT 0 NOT NULL,
    last_tx         NUMERIC(24,8)   DEFAULT 0 NOT NULL,
    last_tx_ts      TIMESTAMPTZ     DEFAULT now() NOT NULL,
    created_at      TIMESTAMPTZ     DEFAULT now() NOT NULL,
    client_id       UUID            REFERENCES users(client_id) ON DELETE CASCADE,
    PRIMARY KEY (client_id, ticker)
);

CREATE INDEX IF NOT EXISTS crypto_client_id_idx ON crypto_accounts USING btree (client_id);
--rollback DROP TABLE crypto_accounts CASCADE;

--changeset surahman:8
--preconditions onFail:HALT onError:HALT
--comment: Create Cryptocurrency operations user and account.
INSERT INTO users (
    first_name,
    last_name,
    email,
    username,
    password,
    is_deleted)
SELECT
   'Internal',
   'FTeX, Inc.',
   'crypto@ftex.com',
   'crypto-currencies',
   password,
   true
FROM
    substr(md5(random()::text), 0, 32) AS password;

INSERT INTO crypto_accounts (
    ticker,
    client_id)
SELECT
   'CRYPTO',
   client_id
FROM
    users AS client_id
WHERE
    username = 'crypto-currencies';
--rollback DELETE FROM users WHERE username='crypto-currencies';

--changeset surahman:9
--preconditions onFail:HALT onError:HALT
--comment: Cryptocurrency accounts transactions journal.
CREATE TABLE IF NOT EXISTS crypto_journal (
    ticker          VARCHAR(6)      NOT NULL,
    amount          NUMERIC(24,8)   NOT NULL,
    transacted_at   TIMESTAMPTZ     NOT NULL,
    client_id       UUID            REFERENCES users(client_id) ON DELETE CASCADE,
    tx_id           UUID            DEFAULT gen_random_uuid() NOT NULL,
    PRIMARY KEY(tx_id, client_id, ticker)
);

CREATE INDEX IF NOT EXISTS crypto_journal_transacted_at_idx ON crypto_journal USING btree (transacted_at);
CREATE INDEX IF NOT EXISTS crypto_journal_tx_idx ON crypto_journal USING btree (tx_id);
--rollback DROP TABLE crypto_journal CASCADE;

--changeset surahman:10
--preconditions onFail:HALT onError:HALT
--comment: Purchase a crypto currency using a base Fiat currency.
CREATE OR REPLACE PROCEDURE purchase_cryptocurrency(
    _client_id              UUID,
    _fiat_currency          Currency,
    _fiat_debit_amount      NUMERIC(20, 2),
    _crypto_ticker          VARCHAR(6),
    _crypto_credit_amount   NUMERIC(24,8)
)
LANGUAGE plpgsql
AS '
    DECLARE
      fiat_balance        NUMERIC(20,2);  -- current balance of the Fiat account.
      crypto_balance      NUMERIC(24,8);  -- current balance of the Crypto account.
      current_timestamp   TIMESTAMPTZ;    -- current timestamp with timezone to be used as transaction timestamp.
      transaction_id      UUID;           -- transaction''s id.
      ftex_fiat_id        UUID;           -- FTeX Fiat operations account id.
      ftex_crypto_id      UUID;           -- FTeX Crypto operations account id.
    BEGIN

      -- Generate the timestamp with timezone for this transaction.
      SELECT NOW() INTO STRICT current_timestamp;

      -- Generate the transaction id for this purchase.
      transaction_id = gen_random_uuid() ;

      RAISE NOTICE ''generated timestamp % and transaction id %'', current_timestamp, transaction_id;

      -- Round Half-to-Even the Fiat debit amount.
      _fiat_debit_amount = round_half_even(_fiat_debit_amount, 2);

      RAISE NOTICE ''rounded debit amount.'';

      -- Get FTeX operations account IDs.
      SELECT client_id INTO STRICT ftex_fiat_id
      FROM users
      WHERE username = ''fiat-currencies'';

      SELECT client_id INTO STRICT ftex_crypto_id
      FROM users
      WHERE username = ''crypto-currencies'';

      RAISE NOTICE ''retrieved operations account IDs.'';

      -- Get balances and row lock the Fiat and then Crypto accounts without locking the foreign keys.
      SELECT fa.balance INTO STRICT fiat_balance
      FROM fiat_accounts AS fa
      WHERE fa.client_id = _client_id AND fa.currency = _fiat_currency
      LIMIT 1
      FOR NO KEY UPDATE;

      SELECT ca.balance INTO STRICT crypto_balance
      FROM crypto_accounts AS ca
      WHERE ca.client_id = _client_id AND ca.ticker = _crypto_ticker
      LIMIT 1
      FOR NO KEY UPDATE;

      RAISE NOTICE ''row locks acquired.'';

      -- Check for sufficient Fiat balance to complete purchase.
      IF _fiat_debit_amount > fiat_balance THEN
         RAISE EXCEPTION ''purchase_cryptocurrency: insufficient Fiat currency funds, delta %'', fiat_balance - _fiat_debit_amount;
      END IF;

      -- Debit the Fiat account and create the Fiat Journal entry.
      UPDATE fiat_accounts
      SET balance = round_half_even(fiat_balance - _fiat_debit_amount, 2),
          last_tx = - _fiat_debit_amount,
          last_tx_ts = current_timestamp
      WHERE client_id = _client_id AND currency = _fiat_currency;

      IF NOT FOUND THEN
        RAISE EXCEPTION ''purchase_cryptocurrency: failed to update Fiat balance'';
      END IF;

      INSERT INTO fiat_journal (client_id, currency, amount, transacted_at, tx_id)
      VALUES (_client_id, _fiat_currency, - _fiat_debit_amount, current_timestamp, transaction_id);

      IF NOT FOUND THEN
        RAISE EXCEPTION ''purchase_cryptocurrency: failed to create Fiat Journal debit entry'';
      END IF;

      RAISE NOTICE ''fiat account debited and journal entries made.'';

      -- Credit the Crypto account and create the Crypto Journal entries for the credit.
      UPDATE crypto_accounts
      SET balance = round_half_even(crypto_balance + _crypto_credit_amount, 8),
          last_tx = _crypto_credit_amount,
          last_tx_ts = current_timestamp
      WHERE client_id = _client_id AND ticker = _crypto_ticker;

      IF NOT FOUND THEN
        RAISE EXCEPTION ''purchase_cryptocurrency: failed to update Crypto balance'';
      END IF;

      INSERT INTO crypto_journal (client_id, ticker, amount, transacted_at, tx_id)
      VALUES (_client_id, _crypto_ticker, _crypto_credit_amount, current_timestamp, transaction_id);

      IF NOT FOUND THEN
        RAISE EXCEPTION ''purchase_cryptocurrency: failed to create Crypto Journal credit entry'';
      END IF;

      RAISE NOTICE ''crypto account debited and journal entries made.'';

      COMMIT;
    END;
';
--rollback DROP PROCEDURE purchase_cryptocurrency;
