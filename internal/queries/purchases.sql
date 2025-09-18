-- name: CreatePurchase :exec
INSERT INTO purchases (
    id, sender_name, sender_contact_type, sender_contact_detail,
    purchased_items, payment_details, total_price
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetPurchaseByID :one
SELECT id, sender_name, sender_contact_type, sender_contact_detail,
       purchased_items, payment_details, total_price, status,
       created_at, updated_at
FROM purchases
WHERE id = @purchaseId::text;

-- name: UpdatePurchaseStatus :exec
UPDATE purchases
SET status = @status::text,
    updated_at = NOW()
WHERE id = @purchaseId::text;
