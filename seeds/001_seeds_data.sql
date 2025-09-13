-- seeds/001_data_users.sql
-- Test users for development (password is "password123")

-- Insert sample users auth
INSERT INTO users_auth (
    id, email, phone, password_hash
) VALUES (
    '00000000-0000-0000-0000-000000000000', 
    'tutuplapak@projectsprint.com', 
    '', 
    '$2a$10$Jy7S4Ea8VgcbeVLrhNaQz.TdHKUEyMFzTkAercFa3TPqee/bKV67.'
), (
    '00000000-0000-0000-0000-000000000001',
    '', 
    '+628562856285', 
    '$2a$10$Jy7S4Ea8VgcbeVLrhNaQz.TdHKUEyMFzTkAercFa3TPqee/bKV67.'
), (
    '00000000-0000-0000-0000-000000000002',
    'seller@example.com',
    '+628123456789',
    '$2a$10$Jy7S4Ea8VgcbeVLrhNaQz.TdHKUEyMFzTkAercFa3TPqee/bKV67.'
), (
    '00000000-0000-0000-0000-000000000003',
    'buyer@example.com', 
    '+628987654321',
    '$2a$10$Jy7S4Ea8VgcbeVLrhNaQz.TdHKUEyMFzTkAercFa3TPqee/bKV67.'
);

-- Insert sample users
INSERT INTO users (
    id, user_auth_id, email, phone, bank_account_name, bank_account_holder, bank_account_number
) VALUES (
    '00000000-0000-0000-0000-000000000010',
    '00000000-0000-0000-0000-000000000000', 
    'tutuplapak@projectsprint.com', 
    '', 
    'Tutup Lapak Corp', 
    'Tutup Lapak', 
    '1234567890'
), (
    '00000000-0000-0000-0000-000000000011',
    '00000000-0000-0000-0000-000000000001', 
    '', 
    '+628562856285', 
    'Buka Jalan Corp', 
    'Buka Jalan', 
    '0987654321'
), (
    '00000000-0000-0000-0000-000000000012',
    '00000000-0000-0000-0000-000000000002', 
    'seller@example.com',
    '+628123456789',
    'Seller Business',
    'John Seller',
    '1122334455'
), (
    '00000000-0000-0000-0000-000000000013',
    '00000000-0000-0000-0000-000000000003',
    'buyer@example.com',
    '+628987654321',
    'Buyer Account',
    'Jane Buyer',
    '5544332211'
);

-- Insert sample files
INSERT INTO files (
    id, user_id, file_uri, file_thumbnail_uri
) VALUES (
    '00000000-0000-0000-0000-000000000100',
    '00000000-0000-0000-0000-000000000012',
    'uploads/sample-product-1.jpg',
    'uploads/sample-product-1-thumb.jpg'
), (
    '00000000-0000-0000-0000-000000000200',
    '00000000-0000-0000-0000-000000000013',
    'uploads/sample-product-2.jpg', 
    'uploads/sample-product-2-thumb.jpg'
);

-- Insert sample products
INSERT INTO products (
    id, name, category, qty, price, sku, file_id, user_id
) VALUES (
    '00000000-0000-0000-0000-100000000000',
    'Gaming Laptop',
    'Electronics',
    5,
    15000000,
    'LAPTOP-001',
    '00000000-0000-0000-0000-000000000100',
    '00000000-0000-0000-0000-000000000012'
), (
    '00000000-0000-0000-0000-200000000000',
    'Wireless Mouse',
    'Electronics', 
    25,
    500000,
    'MOUSE-001',
    '00000000-0000-0000-0000-000000000200',
    '00000000-0000-0000-0000-000000000013'
);