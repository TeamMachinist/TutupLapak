CREATE TYPE purchase_status AS ENUM ('unpaid', 'paid');

CREATE TABLE IF NOT EXISTS purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_name VARCHAR(55) NOT NULL CHECK (LENGTH(sender_name) BETWEEN 4 AND 55),
    sender_contact_type VARCHAR(10) NOT NULL CHECK (sender_contact_type IN ('email', 'phone')),
    sender_contact_detail TEXT NOT NULL,
    purchased_items JSONB NOT NULL,
    payment_details JSONB NOT NULL DEFAULT '[]'::JSONB,
    total_price INTEGER NOT NULL,
    status purchase_status NOT NULL DEFAULT 'unpaid',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index untuk query berdasarkan waktu dan kontak
CREATE INDEX IF NOT EXISTS idx_purchases_created_at ON purchases(created_at);
CREATE INDEX IF NOT EXISTS idx_purchases_sender_contact ON purchases(sender_contact_type, sender_contact_detail);
CREATE INDEX IF NOT EXISTS idx_purchases_status ON purchases(status);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_purchases_updated_at
    BEFORE UPDATE ON purchases
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
