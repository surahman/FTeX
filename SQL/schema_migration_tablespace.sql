--liquibase formatted sql

--changeset surahman:1
--preconditions onFail:HALT onError:HALT
--comment: Users table contains the general user information and login credentials.
CREATE TABLE IF NOT EXISTS users (
    first_name VARCHAR(64)          NOT NULL,
    last_name  VARCHAR(64)          NOT NULL,
    email      VARCHAR(64)          NOT NULL,
    username   VARCHAR(32)          UNIQUE NOT NULL,
    password   VARCHAR(32)          NOT NULL,
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
'WST','XAF','XCD','XDR','XOF','XPF','YER','ZAR','ZMW','ZWD');
--rollback DROP TYPE currency;
