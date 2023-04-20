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
    client_id  UUID                 PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    is_deleted BOOLEAN              DEFAULT false NOT NULL
) TABLESPACE users_data;
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
'DEPOSIT');
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
) TABLESPACE fiat_accounts_data;

CREATE INDEX IF NOT EXISTS fiat_client_id_idx ON fiat_accounts USING btree (client_id) TABLESPACE fiat_accounts_data;
--rollback DROP TABLE fiat_accounts;

--changeset surahman:4
--preconditions onFail:HALT onError:HALT
--comment: Create fiat currency deposit user and account.
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
   'deposit@ftex.com',
   'deposit-fiat',
   password,
   true
FROM
    substr(md5(random()::text), 0, 32) AS password;

INSERT INTO fiat_accounts (
    currency,
    client_id)
SELECT
   'DEPOSIT',
   client_id
FROM
    users AS client_id
WHERE
    username = 'deposit-fiat';
--rollback DELETE FROM users WHERE username='deposit-fiat';

--changeset surahman:5
--preconditions onFail:HALT onError:HALT
--comment: Fiat currency accounts general ledger.
CREATE TABLE IF NOT EXISTS fiat_general_ledger (
    currency        CURRENCY        NOT NULL,
    ammount         NUMERIC(18,2)   NOT NULL,
    transacted_at   TIMESTAMPTZ     NOT NULL,
    client_id       UUID            REFERENCES users(client_id) ON DELETE CASCADE,
    tx_id           UUID            DEFAULT gen_random_uuid() NOT NULL,
    PRIMARY KEY(tx_id, client_id, currency)
) TABLESPACE fiat_general_ledger_data;

CREATE INDEX IF NOT EXISTS fiat_general_ledger_client_idx ON fiat_general_ledger USING btree (client_id) TABLESPACE fiat_general_ledger_data;
CREATE INDEX IF NOT EXISTS fiat_general_ledger_tx_idx ON fiat_general_ledger USING btree (tx_id) TABLESPACE fiat_general_ledger_data;
--rollback DROP TABLE fiat_general_ledger CASCADE;
