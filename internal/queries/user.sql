
-- name: GetUserByID :one
SELECT 
    id,
    email,
    phone,
    bank_account_name,
    bank_account_holder,
    bank_account_number,
    created_at,
    updated_at
FROM users 
WHERE id = $1;