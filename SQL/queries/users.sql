
-- name: createUser :one
-- createUser will create a new user record.
INSERT INTO users (username, password, first_name, last_name, email)
VALUES ($1, $2, $3, $4, $5)
RETURNING client_id;

-- name: getInfoUser :one
-- getInfoUser will retrieve a single users account information.
SELECT first_name, last_name, email, client_id, is_deleted
FROM users
WHERE username=$1
LIMIT 1;

-- name: getCredentialsUser :one
-- getCredentialsUser will retrieve a users client id and password.
SELECT client_id, password
FROM users
WHERE username=$1 AND is_deleted=false
LIMIT 1;

-- name: getClientIdUser :one
-- getClientIdUser will retrieve a users client id.
SELECT client_id
FROM users
WHERE username=$1
LIMIT 1;

-- name: deleteUser :exec
-- deleteUser will soft delete a users account.
UPDATE users
SET is_deleted=true
WHERE username=$1 AND is_deleted=false;