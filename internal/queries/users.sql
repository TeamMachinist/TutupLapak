-- name: CreateUserAuth :one
INSERT INTO users_auth (phone, password_hash)
VALUES ($1, $2)
RETURNING id, phone, password_hash, created_at;

-- name: GetUserAuthByPhone :one
SELECT id, phone, password_hash, created_at
FROM users_auth
WHERE phone = $1;

-- name: GetUserAuthByID :one
SELECT id, phone, password_hash, created_at
FROM users_auth
WHERE id = $1;

-- name: CheckPhoneExists :one
SELECT EXISTS(SELECT 1 FROM users_auth WHERE phone = $1) as exists;


