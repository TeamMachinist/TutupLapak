CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    qty INTEGER NOT NULL DEFAULT 0,
    price INTEGER NOT NULL DEFAULT 0,
    sku VARCHAR(25) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    file_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_created_at ON products(created_at);
CREATE INDEX idx_products_user_id ON products(user_id);
CREATE INDEX idx_products_file_id ON products(file_id);

DROP INDEX IF EXISTS idx_products_name;

CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);

CREATE INDEX IF NOT EXISTS idx_products_price_asc ON products(price ASC);
CREATE INDEX IF NOT EXISTS idx_products_price_desc ON products(price DESC);

CREATE INDEX IF NOT EXISTS idx_products_greatest_created_updated_desc 
ON products (GREATEST(created_at, updated_at) DESC);

CREATE INDEX IF NOT EXISTS idx_products_greatest_created_updated_asc 
ON products (GREATEST(created_at, updated_at) ASC);

CREATE INDEX IF NOT EXISTS idx_products_category_price_asc 
ON products (category, price ASC);

CREATE INDEX IF NOT EXISTS idx_products_category_price_desc 
ON products (category, price DESC);

CREATE INDEX IF NOT EXISTS idx_products_category_newest 
ON products (category, GREATEST(created_at, updated_at) DESC);

CREATE INDEX IF NOT EXISTS idx_products_category_oldest 
ON products (category, GREATEST(created_at, updated_at) ASC);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();