-- Development seed data for ShopSphere eCommerce platform
-- This script populates the database with sample data for development and testing

-- Insert sample users
INSERT INTO users (id, email, username, password_hash, first_name, last_name, phone, role, status, email_verified, created_at, updated_at) VALUES
('user-001', 'admin@shopsphere.com', 'admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Admin', 'User', '+1234567890', 'admin', 'active', true, NOW(), NOW()),
('user-002', 'john.doe@example.com', 'johndoe', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'John', 'Doe', '+1234567891', 'customer', 'active', true, NOW(), NOW()),
('user-003', 'jane.smith@example.com', 'janesmith', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Jane', 'Smith', '+1234567892', 'customer', 'active', true, NOW(), NOW()),
('user-004', 'moderator@shopsphere.com', 'moderator', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Moderator', 'User', '+1234567893', 'moderator', 'active', true, NOW(), NOW()),
('user-005', 'customer@example.com', 'customer', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test', 'Customer', '+1234567894', 'customer', 'active', false, NOW(), NOW());

-- Insert sample addresses
INSERT INTO addresses (id, user_id, type, street, city, state, postal_code, country, is_default, created_at, updated_at) VALUES
('addr-001', 'user-002', 'shipping', '123 Main St', 'New York', 'NY', '10001', 'USA', true, NOW(), NOW()),
('addr-002', 'user-002', 'billing', '123 Main St', 'New York', 'NY', '10001', 'USA', true, NOW(), NOW()),
('addr-003', 'user-003', 'shipping', '456 Oak Ave', 'Los Angeles', 'CA', '90210', 'USA', true, NOW(), NOW()),
('addr-004', 'user-003', 'billing', '456 Oak Ave', 'Los Angeles', 'CA', '90210', 'USA', true, NOW(), NOW()),
('addr-005', 'user-005', 'shipping', '789 Pine Rd', 'Chicago', 'IL', '60601', 'USA', true, NOW(), NOW());

-- Insert sample categories
INSERT INTO categories (id, name, description, parent_id, path, level, sort_order, is_active, created_at, updated_at) VALUES
('cat-001', 'Electronics', 'Electronic devices and accessories', NULL, '/electronics', 0, 1, true, NOW(), NOW()),
('cat-002', 'Smartphones', 'Mobile phones and accessories', 'cat-001', '/electronics/smartphones', 1, 1, true, NOW(), NOW()),
('cat-003', 'Laptops', 'Portable computers and accessories', 'cat-001', '/electronics/laptops', 1, 2, true, NOW(), NOW()),
('cat-004', 'Clothing', 'Apparel and fashion items', NULL, '/clothing', 0, 2, true, NOW(), NOW()),
('cat-005', 'Men''s Clothing', 'Men''s apparel', 'cat-004', '/clothing/mens', 1, 1, true, NOW(), NOW()),
('cat-006', 'Women''s Clothing', 'Women''s apparel', 'cat-004', '/clothing/womens', 1, 2, true, NOW(), NOW()),
('cat-007', 'Books', 'Books and educational materials', NULL, '/books', 0, 3, true, NOW(), NOW()),
('cat-008', 'Fiction', 'Fiction books', 'cat-007', '/books/fiction', 1, 1, true, NOW(), NOW()),
('cat-009', 'Non-Fiction', 'Non-fiction books', 'cat-007', '/books/non-fiction', 1, 2, true, NOW(), NOW()),
('cat-010', 'Home & Garden', 'Home improvement and garden supplies', NULL, '/home-garden', 0, 4, true, NOW(), NOW());

-- Insert sample products
INSERT INTO products (id, sku, name, description, short_description, category_id, price, compare_price, cost_price, currency, stock, low_stock_threshold, status, visibility, weight, length, width, height, images, attributes, featured, created_at, updated_at) VALUES
('prod-001', 'IPHONE-15-128', 'iPhone 15 128GB', 'Latest iPhone with advanced features and improved camera system', 'iPhone 15 with 128GB storage', 'cat-002', 799.00, 899.00, 650.00, 'USD', 50, 10, 'active', 'visible', 0.174, 14.76, 7.15, 0.78, ARRAY['https://example.com/iphone15-1.jpg', 'https://example.com/iphone15-2.jpg'], '{"brand": "Apple", "color": "Blue", "storage": "128GB", "warranty": "1 year"}', true, NOW(), NOW()),
('prod-002', 'MACBOOK-AIR-M2', 'MacBook Air M2', 'Powerful and lightweight laptop with M2 chip', 'MacBook Air with M2 processor', 'cat-003', 1199.00, 1299.00, 950.00, 'USD', 25, 5, 'active', 'visible', 1.24, 30.41, 21.5, 1.13, ARRAY['https://example.com/macbook-air-1.jpg', 'https://example.com/macbook-air-2.jpg'], '{"brand": "Apple", "processor": "M2", "ram": "8GB", "storage": "256GB SSD"}', true, NOW(), NOW()),
('prod-003', 'TSHIRT-COTTON-M', 'Cotton T-Shirt Medium', 'Comfortable 100% cotton t-shirt', 'Premium cotton t-shirt', 'cat-005', 29.99, 39.99, 15.00, 'USD', 100, 20, 'active', 'visible', 0.2, 25.0, 20.0, 2.0, ARRAY['https://example.com/tshirt-1.jpg'], '{"brand": "ShopSphere", "material": "100% Cotton", "size": "M", "color": "Blue"}', false, NOW(), NOW()),
('prod-004', 'DRESS-SUMMER-S', 'Summer Dress Small', 'Elegant summer dress for women', 'Lightweight summer dress', 'cat-006', 79.99, 99.99, 40.00, 'USD', 30, 10, 'active', 'visible', 0.3, 30.0, 25.0, 3.0, ARRAY['https://example.com/dress-1.jpg'], '{"brand": "Fashion Co", "material": "Polyester", "size": "S", "color": "Red"}', false, NOW(), NOW()),
('prod-005', 'BOOK-FICTION-001', 'The Great Adventure', 'An exciting fiction novel about adventure and discovery', 'Adventure fiction novel', 'cat-008', 14.99, 19.99, 8.00, 'USD', 200, 50, 'active', 'visible', 0.4, 23.0, 15.0, 2.5, ARRAY['https://example.com/book-1.jpg'], '{"author": "John Author", "pages": 350, "publisher": "Great Books", "isbn": "978-1234567890"}', false, NOW(), NOW()),
('prod-006', 'SAMSUNG-S24-256', 'Samsung Galaxy S24 256GB', 'Latest Samsung flagship smartphone', 'Galaxy S24 with 256GB storage', 'cat-002', 899.00, 999.00, 720.00, 'USD', 40, 10, 'active', 'visible', 0.168, 14.7, 7.06, 0.76, ARRAY['https://example.com/galaxy-s24-1.jpg'], '{"brand": "Samsung", "color": "Black", "storage": "256GB", "ram": "8GB"}', true, NOW(), NOW()),
('prod-007', 'LAPTOP-DELL-XPS', 'Dell XPS 13', 'Premium ultrabook with Intel processor', 'Dell XPS 13 laptop', 'cat-003', 1099.00, 1199.00, 850.00, 'USD', 15, 5, 'active', 'visible', 1.2, 29.6, 19.9, 1.45, ARRAY['https://example.com/dell-xps-1.jpg'], '{"brand": "Dell", "processor": "Intel i7", "ram": "16GB", "storage": "512GB SSD"}', false, NOW(), NOW()),
('prod-008', 'JEANS-DENIM-32', 'Denim Jeans 32"', 'Classic denim jeans for men', 'Premium denim jeans', 'cat-005', 69.99, 89.99, 35.00, 'USD', 75, 15, 'active', 'visible', 0.6, 32.0, 28.0, 4.0, ARRAY['https://example.com/jeans-1.jpg'], '{"brand": "Denim Co", "material": "Denim", "size": "32", "color": "Blue"}', false, NOW(), NOW()),
('prod-009', 'COOKBOOK-ITALIAN', 'Italian Cooking Masterclass', 'Learn authentic Italian cooking techniques', 'Italian cookbook', 'cat-009', 24.99, 29.99, 12.00, 'USD', 80, 20, 'active', 'visible', 0.8, 26.0, 20.0, 3.0, ARRAY['https://example.com/cookbook-1.jpg'], '{"author": "Chef Mario", "pages": 280, "publisher": "Culinary Press", "cuisine": "Italian"}', false, NOW(), NOW()),
('prod-010', 'PLANT-POT-CERAMIC', 'Ceramic Plant Pot', 'Beautiful ceramic pot for indoor plants', 'Decorative ceramic plant pot', 'cat-010', 34.99, 44.99, 18.00, 'USD', 60, 15, 'active', 'visible', 1.2, 20.0, 20.0, 18.0, ARRAY['https://example.com/plant-pot-1.jpg'], '{"material": "Ceramic", "color": "White", "size": "Medium", "drainage": "Yes"}', false, NOW(), NOW());

-- Insert product variants
INSERT INTO product_variants (id, product_id, sku, name, price, compare_price, cost_price, stock, attributes, is_default, created_at, updated_at) VALUES
('var-001', 'prod-001', 'IPHONE-15-128-BLUE', 'iPhone 15 128GB Blue', 799.00, 899.00, 650.00, 25, '{"color": "Blue"}', true, NOW(), NOW()),
('var-002', 'prod-001', 'IPHONE-15-128-BLACK', 'iPhone 15 128GB Black', 799.00, 899.00, 650.00, 25, '{"color": "Black"}', false, NOW(), NOW()),
('var-003', 'prod-003', 'TSHIRT-COTTON-M-BLUE', 'Cotton T-Shirt Medium Blue', 29.99, 39.99, 15.00, 50, '{"color": "Blue", "size": "M"}', true, NOW(), NOW()),
('var-004', 'prod-003', 'TSHIRT-COTTON-M-RED', 'Cotton T-Shirt Medium Red', 29.99, 39.99, 15.00, 50, '{"color": "Red", "size": "M"}', false, NOW(), NOW());

-- Insert product tags
INSERT INTO product_tags (id, name, slug, description, created_at) VALUES
('tag-001', 'Bestseller', 'bestseller', 'Top selling products', NOW()),
('tag-002', 'New Arrival', 'new-arrival', 'Recently added products', NOW()),
('tag-003', 'Sale', 'sale', 'Products on sale', NOW()),
('tag-004', 'Premium', 'premium', 'High-end premium products', NOW()),
('tag-005', 'Eco-Friendly', 'eco-friendly', 'Environmentally friendly products', NOW());

-- Insert product-tag relationships
INSERT INTO product_tag_relations (product_id, tag_id, created_at) VALUES
('prod-001', 'tag-001', NOW()),
('prod-001', 'tag-004', NOW()),
('prod-002', 'tag-001', NOW()),
('prod-002', 'tag-004', NOW()),
('prod-003', 'tag-002', NOW()),
('prod-004', 'tag-002', NOW()),
('prod-005', 'tag-003', NOW()),
('prod-006', 'tag-001', NOW()),
('prod-010', 'tag-005', NOW());

-- Insert sample orders
INSERT INTO orders (id, order_number, user_id, status, subtotal, tax, shipping, total, currency, shipping_address, billing_address, payment_method, payment_status, shipping_method, notes, confirmed_at, created_at, updated_at) VALUES
('order-001', 'ORD-20241208-000001', 'user-002', 'delivered', 828.99, 66.32, 9.99, 905.30, 'USD', 
 '{"street": "123 Main St", "city": "New York", "state": "NY", "postal_code": "10001", "country": "USA"}',
 '{"street": "123 Main St", "city": "New York", "state": "NY", "postal_code": "10001", "country": "USA"}',
 '{"type": "card", "last4": "1234", "brand": "visa"}', 'completed', 'standard', 'Please deliver to front door', NOW() - INTERVAL '7 days', NOW() - INTERVAL '10 days', NOW() - INTERVAL '7 days'),
('order-002', 'ORD-20241208-000002', 'user-003', 'shipped', 1199.00, 95.92, 0.00, 1294.92, 'USD',
 '{"street": "456 Oak Ave", "city": "Los Angeles", "state": "CA", "postal_code": "90210", "country": "USA"}',
 '{"street": "456 Oak Ave", "city": "Los Angeles", "state": "CA", "postal_code": "90210", "country": "USA"}',
 '{"type": "card", "last4": "5678", "brand": "mastercard"}', 'completed', 'express', NULL, NOW() - INTERVAL '2 days', NOW() - INTERVAL '3 days', NOW() - INTERVAL '2 days'),
('order-003', 'ORD-20241208-000003', 'user-005', 'processing', 109.98, 8.80, 5.99, 124.77, 'USD',
 '{"street": "789 Pine Rd", "city": "Chicago", "state": "IL", "postal_code": "60601", "country": "USA"}',
 '{"street": "789 Pine Rd", "city": "Chicago", "state": "IL", "postal_code": "60601", "country": "USA"}',
 '{"type": "paypal", "email": "customer@example.com"}', 'completed', 'standard', NULL, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- Insert order items
INSERT INTO order_items (id, order_id, product_id, variant_id, sku, name, price, quantity, total, product_attributes, created_at) VALUES
('item-001', 'order-001', 'prod-001', 'var-001', 'IPHONE-15-128-BLUE', 'iPhone 15 128GB Blue', 799.00, 1, 799.00, '{"brand": "Apple", "color": "Blue", "storage": "128GB"}', NOW() - INTERVAL '10 days'),
('item-002', 'order-001', 'prod-003', 'var-003', 'TSHIRT-COTTON-M-BLUE', 'Cotton T-Shirt Medium Blue', 29.99, 1, 29.99, '{"brand": "ShopSphere", "color": "Blue", "size": "M"}', NOW() - INTERVAL '10 days'),
('item-003', 'order-002', 'prod-002', NULL, 'MACBOOK-AIR-M2', 'MacBook Air M2', 1199.00, 1, 1199.00, '{"brand": "Apple", "processor": "M2", "ram": "8GB"}', NOW() - INTERVAL '3 days'),
('item-004', 'order-003', 'prod-004', NULL, 'DRESS-SUMMER-S', 'Summer Dress Small', 79.99, 1, 79.99, '{"brand": "Fashion Co", "size": "S", "color": "Red"}', NOW() - INTERVAL '1 day'),
('item-005', 'order-003', 'prod-003', 'var-003', 'TSHIRT-COTTON-M-BLUE', 'Cotton T-Shirt Medium Blue', 29.99, 1, 29.99, '{"brand": "ShopSphere", "color": "Blue", "size": "M"}', NOW() - INTERVAL '1 day');

-- Insert sample reviews
INSERT INTO reviews (id, product_id, user_id, order_id, order_item_id, rating, title, content, status, verified_purchase, helpful_count, created_at, updated_at) VALUES
('review-001', 'prod-001', 'user-002', 'order-001', 'item-001', 5, 'Excellent phone!', 'The iPhone 15 is amazing. Great camera quality and battery life. Highly recommended!', 'approved', true, 5, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('review-002', 'prod-003', 'user-002', 'order-001', 'item-002', 4, 'Good quality shirt', 'Nice cotton material and comfortable fit. Good value for money.', 'approved', true, 2, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('review-003', 'prod-002', 'user-003', 'order-002', 'item-003', 5, 'Perfect laptop for work', 'The MacBook Air M2 is incredibly fast and lightweight. Perfect for my daily work needs.', 'approved', true, 8, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
('review-004', 'prod-001', 'user-005', NULL, NULL, 4, 'Great phone but expensive', 'Love the features but wish it was more affordable. Still a solid choice.', 'approved', false, 1, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days');

-- Insert review votes
INSERT INTO review_votes (id, review_id, user_id, vote_type, created_at) VALUES
('vote-001', 'review-001', 'user-003', 'helpful', NOW() - INTERVAL '4 days'),
('vote-002', 'review-001', 'user-005', 'helpful', NOW() - INTERVAL '3 days'),
('vote-003', 'review-002', 'user-003', 'helpful', NOW() - INTERVAL '4 days'),
('vote-004', 'review-003', 'user-002', 'helpful', NOW() - INTERVAL '1 day'),
('vote-005', 'review-003', 'user-005', 'helpful', NOW() - INTERVAL '1 day');

-- Insert shopping carts
INSERT INTO shopping_carts (id, user_id, session_id, expires_at, created_at, updated_at) VALUES
('cart-001', 'user-002', NULL, NOW() + INTERVAL '30 days', NOW(), NOW()),
('cart-002', NULL, 'guest-session-123', NOW() + INTERVAL '7 days', NOW(), NOW());

-- Insert shopping cart items
INSERT INTO shopping_cart_items (id, cart_id, product_id, variant_id, sku, quantity, price, created_at, updated_at) VALUES
('cart-item-001', 'cart-001', 'prod-006', NULL, 'SAMSUNG-S24-256', 1, 899.00, NOW(), NOW()),
('cart-item-002', 'cart-001', 'prod-008', NULL, 'JEANS-DENIM-32', 2, 69.99, NOW(), NOW()),
('cart-item-003', 'cart-002', 'prod-005', NULL, 'BOOK-FICTION-001', 1, 14.99, NOW(), NOW());

-- Insert inventory movements
INSERT INTO inventory_movements (id, product_id, variant_id, movement_type, quantity, reference_type, reference_id, reason, created_by, created_at) VALUES
('inv-001', 'prod-001', 'var-001', 'out', 1, 'order', 'order-001', 'Order fulfillment', 'user-001', NOW() - INTERVAL '10 days'),
('inv-002', 'prod-003', 'var-003', 'out', 1, 'order', 'order-001', 'Order fulfillment', 'user-001', NOW() - INTERVAL '10 days'),
('inv-003', 'prod-002', NULL, 'out', 1, 'order', 'order-002', 'Order fulfillment', 'user-001', NOW() - INTERVAL '3 days'),
('inv-004', 'prod-004', NULL, 'reserved', 1, 'order', 'order-003', 'Order processing', 'user-001', NOW() - INTERVAL '1 day'),
('inv-005', 'prod-003', 'var-003', 'reserved', 1, 'order', 'order-003', 'Order processing', 'user-001', NOW() - INTERVAL '1 day');

-- Update product review summaries
INSERT INTO product_review_summaries (product_id, total_reviews, average_rating, rating_1_count, rating_2_count, rating_3_count, rating_4_count, rating_5_count, verified_reviews_count, last_review_date, updated_at) VALUES
('prod-001', 2, 4.50, 0, 0, 0, 1, 1, 1, NOW() - INTERVAL '3 days', NOW()),
('prod-002', 1, 5.00, 0, 0, 0, 0, 1, 1, NOW() - INTERVAL '1 day', NOW()),
('prod-003', 1, 4.00, 0, 0, 0, 1, 0, 1, NOW() - INTERVAL '5 days', NOW());

COMMIT;