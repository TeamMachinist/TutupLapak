-- name: CreatePurchase :exec
INSERT INTO purchases (
    id, sender_name, sender_contact_type, sender_contact_detail,
    purchased_items, payment_details, total_price
) VALUES ($1, $2, $3, $4, $5, $6, $7);
