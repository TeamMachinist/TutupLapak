-- name CheckExistedUserByEmail :one
SELECT 1 FROM users_auth WHERE email = $email;

-- name GetUserByEmail :one
SELECT * FROM users_auth WHERE email = $email;

-- name RegisterWithEmail :execresult
INSERT INTO users_auth (id, email, phone,hashed_password, created_at, updated_at) VALUES ($1, $2, $3, $4, $5);