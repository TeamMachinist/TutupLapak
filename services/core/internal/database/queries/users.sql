-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByAuthID :one
SELECT * FROM users WHERE user_auth_id = $1;

-- name: UpdateUser :one
UPDATE users SET
    file_id = COALESCE($2, file_id),
    bank_account_name = $3,
    bank_account_holder = $4,
    bank_account_number = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserEmail :one
UPDATE users SET
    email = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserPhone :one
UPDATE users SET
    phone = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1) as exists;

-- name: CheckPhoneExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1) as exists;

-- name: CreateUserFromUserAuth :one
INSERT INTO users (id, user_auth_id, file_id, email, phone, bank_account_name, bank_account_holder, bank_account_number)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING *;