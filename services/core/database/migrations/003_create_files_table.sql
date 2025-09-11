-- Opsional: hanya jika Anda ingin menyimpan metadata file lokal
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY,
    uri TEXT NOT NULL,
    thumbnail_uri TEXT,
    user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_files_user_id ON files(user_id);