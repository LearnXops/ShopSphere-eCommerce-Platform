-- Test seed data for ShopSphere eCommerce platform
-- This script populates the database with minimal test data for automated testing

-- Insert test users
INSERT INTO users (id, email, username, password_hash, first_name, last_name, phone, role, status, email_verified, created_at, updated_at) VALUES
('test-user-001', 'test.admin@test.com', 'testadmin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test', 'Admin', '+1000000001', 'admin', 'active', true, NOW(), NOW()),
('test-user-002', 'test.customer@test.com', 'testcustomer', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test', 'Customer', '+1000000002', 'customer', 'active', true, NOW(), NOW()),
('test-user-003', 'test.inactive@test.com', 'testinactive', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test', 'Inactive', '+1000000003', 'customer', 'suspended', false, NOW(), NOW());

-- Insert test addresses
INSERT INTO addresses (id, user_id, type, street, city, state, postal_code, country, is_default, created_at, updated_at) VALUES
('test-addr-001', 'test-user-002', 'shipping', '123 Test St', 'Test City', 'TS', '12345', 'USA', true, NOW(), NOW()),
('test-addr-002', 'test-user-002', 'billing', '123 Test St', 'Test City', 'TS', '12345', 'USA', true, NOW(), NOW());

-- Insert test categories
INSERT INTO categories (id, name, description, parent_id, path, level, sort_order, is_active, created_at, updated_at) VALUES
('test-cat-001', 'Test Electronics', 'Test electronic devices', NULL, '/test-electronics', 0, 1, true, NOW(), NOW()),
('test-cat-002', 'Test Phones', 'Test mobile phones', 'test-cat-001', '/test-electronics/test-phones', 1, 1, true, NOW(), NOW()),
('test-cat-003', 'Test Clothing', 'Test clothing items', NULL, '/test-clothing', 0, 2, true, NOW(), NOW());

-- Insert test products
INSERT INTO products (id, sku, name, description, short_description, category_id, price, cost_price, currency, stock, status, visibility, weight, attributes, featured, created_at, updated_at) VALUES
('test-prod-001', 'TEST-PHONE-001', 'Test Phone', 'A test smartphone for testing purposes', 'Test smartphone', 'test-cat-002', 299.99, 200.00, 'USD', 10, 'active', 'visible', 0.2, '{"brand": "TestBrand", "color": "Black"}', false, NOW(), NOW()),
('test-prod-002', 'TEST-SHIRT-001', 'Test Shirt', 'A test shirt for testing purposes', 'Test shirt', 'test-cat-003', 19.99, 10.00, 'USD', 50, 'active', 'visible', 0.1, '{"brand": "TestClothing", "size": "M"}', false, NOW(), NOW()),
('test-prod-003', 'TEST-INACTIVE-001', 'Test Inactive Product', 'An inactive test product', 'Inactive test product', 'test-cat-001', 99.99, 50.00, 'USD', 0, 'inactive', 'hidden', 0.3, '{"brand": "TestBrand"}', false, NOW(), NOW());

-- Insert test product variants
INSERT INTO product_variants (id, product_id, sku, name, price, cost_price, stock, attributes, is_default, created_at, updated_at) VALUES
('test-var-001', 'test-prod-001', 'TEST-PHONE-001-BLACK', 'Test Phone Black', 299.99, 200.00, 5, '{"color": "Black"}', true, NOW(), NOW()),
('test-var-002', 'test-prod-001', 'TEST-PHONE-001-WHITE', 'Test Phone White', 299.99, 200.00, 5, '{"color": "White"}', false, NOW(), NOW());

-- Insert test orders
INSERT INTO orders (id, order_number, user_id, status, subtotal, tax, shipping, total, currency, shipping_address, billing_address, payment_method, payment_status, created_at, updated_at) VALUES
('test-order-001', 'TEST-ORD-001', 'test-user-002', 'pending', 299.99, 24.00, 5.99, 329.98, 'USD',
 '{"street": "123 Test St", "city": "Test City", "state": "TS", "postal_code": "12345", "country": "USA"}',
 '{"street": "123 Test St", "city": "Test City", "state": "TS", "postal_code": "12345", "country": "USA"}',
 '{"type": "card", "last4": "0000", "brand": "test"}', 'pending', NOW(), NOW());

-- Insert test order items
INSERT INTO order_items (id, order_id, product_id, variant_id, sku, name, price, quantity, total, created_at) VALUES
('test-item-001', 'test-order-001', 'test-prod-001', 'test-var-001', 'TEST-PHONE-001-BLACK', 'Test Phone Black', 299.99, 1, 299.99, NOW());

-- Insert test reviews
INSERT INTO reviews (id, product_id, user_id, order_id, order_item_id, rating, title, content, status, verified_purchase, created_at, updated_at) VALUES
('test-review-001', 'test-prod-001', 'test-user-002', 'test-order-001', 'test-item-001', 5, 'Great test product', 'This test product works perfectly for testing.', 'approved', true, NOW(), NOW()),
('test-review-002', 'test-prod-002', 'test-user-002', NULL, NULL, 3, 'Average test product', 'This test product is okay for testing purposes.', 'pending', false, NOW(), NOW());

-- Insert test shopping cart
INSERT INTO shopping_carts (id, user_id, session_id, expires_at, created_at, updated_at) VALUES
('test-cart-001', 'test-user-002', NULL, NOW() + INTERVAL '1 day', NOW(), NOW());

-- Insert test cart items
INSERT INTO shopping_cart_items (id, cart_id, product_id, variant_id, sku, quantity, price, created_at, updated_at) VALUES
('test-cart-item-001', 'test-cart-001', 'test-prod-002', NULL, 'TEST-SHIRT-001', 2, 19.99, NOW(), NOW());

-- Update test product review summaries
INSERT INTO product_review_summaries (product_id, total_reviews, average_rating, rating_1_count, rating_2_count, rating_3_count, rating_4_count, rating_5_count, verified_reviews_count, last_review_date, updated_at) VALUES
('test-prod-001', 1, 5.00, 0, 0, 0, 0, 1, 1, NOW(), NOW()),
('test-prod-002', 1, 3.00, 0, 0, 1, 0, 0, 0, NOW(), NOW());

COMMIT;