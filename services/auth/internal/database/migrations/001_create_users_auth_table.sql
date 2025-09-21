CREATE TABLE IF NOT EXISTS users_auth (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CRITICAL: Login endpoints hit 23K RPS combined
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_auth_email ON users_auth(email) WHERE email IS NOT NULL AND email != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_auth_phone ON users_auth(phone) WHERE phone IS NOT NULL AND phone != '';

-- Optional: Composite index for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_auth_credentials ON users_auth(email, password_hash) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_auth_phone_credentials ON users_auth(phone, password_hash) WHERE phone IS NOT NULL;