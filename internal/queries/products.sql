
-- name: CreateProduct :one
INSERT INTO products (
    id,
    name,
    category,
    qty,
    price,
    sku,
    file_id,
    user_id,
    created_at,
    updated_at
) VALUES (
    @id::uuid,
    @name::text,
    @category::text,
    @qty,             
    @price,     
    @sku::text,
    @file_id::uuid,
    @user_id::uuid,
    @created_at,
    @updated_at
)
RETURNING *;

-- name: CheckSKUExistsByUser :one
SELECT id, sku
FROM products
WHERE sku = @sku::text AND user_id = @user_id::uuid
LIMIT 1;

-- name: GetAllProducts :many
SELECT 
    p.id,
    p.name,
    p.category,
    p.qty,
    p.price,
    p.sku,
    p.file_id,
    p.user_id,
    p.created_at,
    p.updated_at
FROM products p
WHERE 
    p.id = COALESCE(NULLIF(@product_id::uuid, '00000000-0000-0000-0000-000000000000'::uuid), p.id)
    AND p.sku = COALESCE(NULLIF(@sku::text, ''), p.sku)
    AND p.category = COALESCE(NULLIF(@category::text, ''), p.category)
ORDER BY 
    CASE WHEN @sort_by::text = 'newest' THEN GREATEST(p.created_at, p.updated_at) END DESC,
    CASE WHEN @sort_by::text = 'oldest' THEN GREATEST(p.created_at, p.updated_at) END ASC,
    CASE WHEN @sort_by::text = 'cheapest' THEN p.price END ASC,
    CASE WHEN @sort_by::text = 'expensive' THEN p.price END DESC,
    p.created_at DESC
LIMIT COALESCE(@limit_count::int, 5)
OFFSET COALESCE(@offset_count::int, 0);

-- name: UpdateProduct :one
UPDATE products SET
    name = COALESCE(NULLIF(@name::text, ''), name),
    category = COALESCE(NULLIF(@category::text, ''), category),
    qty = COALESCE(@qty, qty),
    price = COALESCE(@price, price),
    sku = COALESCE(NULLIF(@sku::text, ''), sku),
    file_id = COALESCE(NULLIF(@file_id::uuid, '00000000-0000-0000-0000-000000000000'::uuid), file_id),
    updated_at = @updated_at
WHERE id = @id::uuid
RETURNING id, name, category, qty, price, sku, file_id, created_at, updated_at;


-- name: CheckProductOwnership :one
SELECT EXISTS(
    SELECT 1 FROM products
    WHERE id = @product_id::uuid AND user_id = @user_id::uuid
) as exists;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = @id::uuid AND user_id = @user_id::uuid;

-- name: UpdateProductQty :execrows
UPDATE products 
SET qty = qty - $2 
WHERE id = $1;

-- name: GetProductByID :one
SELECT 
    id,
    name,
    category,
    qty,
    price,
    sku,
    file_id,
    user_id,
    created_at,
    updated_at
FROM products 
WHERE id = $1;