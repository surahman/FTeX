-- Users table contains the general user information and login credentials.
CREATE TABLE IF NOT EXISTS users (
    first_name VARCHAR(64)           NOT NULL,
    last_name  VARCHAR(64)           NOT NULL,
    email      VARCHAR(64)           NOT NULL,
    username   VARCHAR(32)           NOT NULL,
    password   VARCHAR(32)           NOT NULL,
    client_id  UUID                  NOT NULL
        CONSTRAINT users_pk
            PRIMARY KEY,
    is_deleted BOOLEAN DEFAULT false NOT NULL
) TABLESPACE users_data;
