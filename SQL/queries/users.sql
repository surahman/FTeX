
-- name: CreateUser :exec
INSERT INTO users (username, password, first_name, last_name, email) VALUES ($1, $2, $3, $4, $5);

-- GetInfoUser :one
SELECT first_name, last_name, email, client_id, is_deleted
FROM users
WHERE username=$1
LIMIT 1;

-- name: GetCredentialsUser :one
SELECT password
FROM users
WHERE username=$1 AND is_deleted=false
LIMIT 1;

-- name: GetClientIdUser :one
SELECT client_id
FROM users
WHERE username=$1
LIMIT 1;

-- name: DeleteUser :exec
UPDATE users
SET is_deleted=true
WHERE username=$1 AND is_deleted=false;
