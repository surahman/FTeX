--liquibase formatted sql

--changeset surahman:1
--preconditions onFail:HALT onError:HALT
--comment: Users table contains the general user information and login credentials.
CREATE TABLE IF NOT EXISTS users (
    first_name VARCHAR(64)           NOT NULL,
    last_name  VARCHAR(64)           NOT NULL,
    email      VARCHAR(64)           NOT NULL,
    username   VARCHAR(32)           UNIQUE NOT NULL,
    password   VARCHAR(32)           NOT NULL,
    client_id  UUID                  PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    is_deleted BOOLEAN DEFAULT false NOT NULL
);
--rollback DROP TABLE users;
