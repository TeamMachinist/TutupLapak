-- name: GetUserAuthByID :one
SELECT id, email, phone, password_hash, created_at
FROM users_auth
WHERE id = $1;

-- name: CreateUserAuth :one
INSERT INTO users_auth (phone, password_hash)
VALUES ($1, $2)
RETURNING id, email, phone, password_hash, created_at;

-- name: GetUserAuthByPhone :one
SELECT id, email, phone, password_hash, created_at
FROM users_auth
WHERE phone = $1;

-- name: CheckPhoneExists :one
SELECT EXISTS(SELECT 1 FROM users_auth WHERE phone = $1) as exists;

-- name: RegisterWithEmail :one
INSERT INTO users_auth (email, password_hash)
VALUES ($1, $2)
RETURNING id, email, phone, password_hash, created_at;

-- name: GetUserAuthByEmail :one
SELECT id, email, phone, password_hash, created_at
FROM users_auth
WHERE email = $1;

-- name: CheckExistedUserAuthByEmail :one
SELECT EXISTS(SELECT 1 FROM users_auth WHERE email = $1) as exists;
