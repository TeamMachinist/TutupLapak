-- name: GetUserAuthByID :one
SELECT
    id,
    COALESCE(email, '') as email,
    COALESCE(phone, '') as phone,
    password_hash,
    created_at
FROM users_auth
WHERE id = $1;

-- name: CreateUserByPhone :one
INSERT INTO users_auth (phone, password_hash)
VALUES ($1, $2)
RETURNING 
    id, 
    COALESCE(email, '') as email,
    COALESCE(phone, '') as phone,       
    password_hash, 
    created_at;

-- name: GetUserAuthByPhone :one
SELECT
    id,
    COALESCE(email, '') as email,
    COALESCE(phone, '') as phone,
    password_hash,
    created_at
FROM users_auth
WHERE phone = $1;

-- name: CheckPhoneExists :one
SELECT EXISTS(SELECT 1 FROM users_auth WHERE phone = $1) as exists;

-- name: CreateUserByEmail :one
INSERT INTO users_auth (email, password_hash)
VALUES ($1, $2)
RETURNING 
    id, 
    COALESCE(email, '') as email,
    COALESCE(phone, '') as phone,       
    password_hash, 
    created_at;

-- name: GetUserAuthByEmail :one
SELECT
    id,
    COALESCE(email, '') as email,
    COALESCE(phone, '') as phone,
    password_hash,
    created_at
FROM users_auth
WHERE email = $1;

-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM users_auth WHERE email = $1) as exists;
