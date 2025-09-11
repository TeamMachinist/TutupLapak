-- name: CheckExistedUserByEmail :one
SELECT 1 FROM users_auth WHERE email = $1;

-- name: GetUserByEmail :one
SELECT * FROM users_auth WHERE email = $1;

-- name: RegisterWithEmail :exec
INSERT INTO users_auth (id, email, phone, hashed_password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6);