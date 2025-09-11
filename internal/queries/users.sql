-- name: CreateUserAuth :one
INSERT INTO users_auth (email, phone, password_hash)
VALUES ($1, $2, $3)
RETURNING id, email, phone, password_hash, created_at;

-- name: GetUserAuthByEmail :one
SELECT id, email, phone, password_hash, created_at
FROM users_auth
WHERE email = $1;

-- name: GetUserAuthByPhone :one
SELECT id, email, phone, password_hash, created_at
FROM users_auth
WHERE phone = $1;

-- name: GetUserAuthByID :one
SELECT id, email, phone, password_hash, created_at
FROM users_auth
WHERE id = $1;

-- name: UpdateUserPhone :one
UPDATE users_auth
SET phone = $2
WHERE id = $1
RETURNING id, email, phone, password_hash, created_at;

-- name: UpdateUserEmail :one
UPDATE users_auth
SET email = $2
WHERE id = $1
RETURNING id, email, phone, password_hash, created_at;

-- name: CheckPhoneExists :one
SELECT EXISTS(SELECT 1 FROM users_auth WHERE phone = $1) as exists;

-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM users_auth WHERE email = $1) as exists;


