-- Rollback order service database schema

-- Drop triggers
DROP TRIGGER IF EXISTS generate_order_number_trigger ON orders;
DROP TRIGGER IF EXISTS order_status_history_trigger ON orders;
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS update_shopping_carts_updated_at ON shopping_carts;
DROP TRIGGER IF EXISTS update_shopping_cart_items_updated_at ON shopping_cart_items;
DROP TRIGGER IF EXISTS update_order_fulfillments_updated_at ON order_fulfillments;

-- Drop functions
DROP FUNCTION IF EXISTS generate_order_number();
DROP FUNCTION IF EXISTS create_order_status_history();

-- Drop sequence
DROP SEQUENCE IF EXISTS order_number_seq;

-- Drop indexes
DROP INDEX IF EXISTS idx_orders_order_number;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_payment_status;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_confirmed_at;
DROP INDEX IF EXISTS idx_orders_shipped_at;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_sku;
DROP INDEX IF EXISTS idx_order_status_history_order_id;
DROP INDEX IF EXISTS idx_order_discounts_order_id;
DROP INDEX IF EXISTS idx_order_discounts_code;
DROP INDEX IF EXISTS idx_shopping_carts_user_id;
DROP INDEX IF EXISTS idx_shopping_carts_session_id;
DROP INDEX IF EXISTS idx_shopping_carts_expires_at;
DROP INDEX IF EXISTS idx_shopping_cart_items_cart_id;
DROP INDEX IF EXISTS idx_shopping_cart_items_product_id;
DROP INDEX IF EXISTS idx_order_fulfillments_order_id;
DROP INDEX IF EXISTS idx_order_fulfillments_status;
DROP INDEX IF EXISTS idx_order_fulfillment_items_fulfillment_id;
DROP INDEX IF EXISTS idx_order_fulfillment_items_order_item_id;

-- Drop tables
DROP TABLE IF EXISTS order_fulfillment_items;
DROP TABLE IF EXISTS order_fulfillments;
DROP TABLE IF EXISTS shopping_cart_items;
DROP TABLE IF EXISTS shopping_carts;
DROP TABLE IF EXISTS order_discounts;
DROP TABLE IF EXISTS order_status_history;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;