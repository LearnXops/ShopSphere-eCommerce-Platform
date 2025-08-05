-- Create order service database schema
-- This migration creates the order processing and management tables

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    order_number VARCHAR(50) UNIQUE NOT NULL, -- human-readable order number
    user_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'processing', 'shipped', 'delivered', 'cancelled', 'refunded')),
    subtotal DECIMAL(10,2) NOT NULL CHECK (subtotal >= 0),
    tax DECIMAL(10,2) DEFAULT 0 CHECK (tax >= 0),
    shipping DECIMAL(10,2) DEFAULT 0 CHECK (shipping >= 0),
    discount DECIMAL(10,2) DEFAULT 0 CHECK (discount >= 0),
    total DECIMAL(10,2) NOT NULL CHECK (total >= 0),
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Address information stored as JSONB for flexibility
    shipping_address JSONB NOT NULL,
    billing_address JSONB NOT NULL,
    
    -- Payment information
    payment_method JSONB,
    payment_status VARCHAR(20) DEFAULT 'pending' CHECK (payment_status IN ('pending', 'processing', 'completed', 'failed', 'refunded', 'partially_refunded')),
    payment_reference VARCHAR(255), -- external payment gateway reference
    
    -- Shipping information
    shipping_method VARCHAR(100),
    tracking_number VARCHAR(100),
    estimated_delivery_date DATE,
    actual_delivery_date DATE,
    
    -- Additional information
    notes TEXT,
    internal_notes TEXT, -- admin-only notes
    source VARCHAR(50) DEFAULT 'web', -- web, mobile, api, etc.
    
    -- Timestamps
    confirmed_at TIMESTAMP,
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order items table
CREATE TABLE IF NOT EXISTS order_items (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL,
    variant_id VARCHAR(36), -- if product has variants
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    total DECIMAL(10,2) NOT NULL CHECK (total >= 0),
    
    -- Product snapshot at time of order
    product_attributes JSONB DEFAULT '{}',
    image_url TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order status history table for tracking status changes
CREATE TABLE IF NOT EXISTS order_status_history (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    from_status VARCHAR(20),
    to_status VARCHAR(20) NOT NULL,
    reason TEXT,
    notes TEXT,
    changed_by VARCHAR(36), -- user who made the change
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order discounts table for tracking applied discounts
CREATE TABLE IF NOT EXISTS order_discounts (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('coupon', 'promotion', 'loyalty', 'manual')),
    discount_code VARCHAR(100),
    discount_name VARCHAR(255) NOT NULL,
    discount_amount DECIMAL(10,2) NOT NULL CHECK (discount_amount >= 0),
    discount_percentage DECIMAL(5,2) CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    applied_to VARCHAR(20) DEFAULT 'order' CHECK (applied_to IN ('order', 'shipping', 'item')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Shopping carts table (temporary storage before order creation)
CREATE TABLE IF NOT EXISTS shopping_carts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36), -- null for guest carts
    session_id VARCHAR(255), -- for guest carts
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT cart_user_or_session CHECK (
        (user_id IS NOT NULL AND session_id IS NULL) OR 
        (user_id IS NULL AND session_id IS NOT NULL)
    )
);

-- Shopping cart items table
CREATE TABLE IF NOT EXISTS shopping_cart_items (
    id VARCHAR(36) PRIMARY KEY,
    cart_id VARCHAR(36) NOT NULL REFERENCES shopping_carts(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL,
    variant_id VARCHAR(36),
    sku VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0), -- price at time of adding to cart
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cart_id, product_id, variant_id)
);

-- Order fulfillment table for tracking fulfillment status
CREATE TABLE IF NOT EXISTS order_fulfillments (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    fulfillment_status VARCHAR(20) DEFAULT 'pending' CHECK (fulfillment_status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled')),
    tracking_number VARCHAR(100),
    carrier VARCHAR(100),
    tracking_url TEXT,
    shipped_at TIMESTAMP,
    delivered_at TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Order fulfillment items table (for partial fulfillments)
CREATE TABLE IF NOT EXISTS order_fulfillment_items (
    id VARCHAR(36) PRIMARY KEY,
    fulfillment_id VARCHAR(36) NOT NULL REFERENCES order_fulfillments(id) ON DELETE CASCADE,
    order_item_id VARCHAR(36) NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_confirmed_at ON orders(confirmed_at);
CREATE INDEX idx_orders_shipped_at ON orders(shipped_at);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_order_items_sku ON order_items(sku);
CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX idx_order_discounts_order_id ON order_discounts(order_id);
CREATE INDEX idx_order_discounts_code ON order_discounts(discount_code);
CREATE INDEX idx_shopping_carts_user_id ON shopping_carts(user_id);
CREATE INDEX idx_shopping_carts_session_id ON shopping_carts(session_id);
CREATE INDEX idx_shopping_carts_expires_at ON shopping_carts(expires_at);
CREATE INDEX idx_shopping_cart_items_cart_id ON shopping_cart_items(cart_id);
CREATE INDEX idx_shopping_cart_items_product_id ON shopping_cart_items(product_id);
CREATE INDEX idx_order_fulfillments_order_id ON order_fulfillments(order_id);
CREATE INDEX idx_order_fulfillments_status ON order_fulfillments(fulfillment_status);
CREATE INDEX idx_order_fulfillment_items_fulfillment_id ON order_fulfillment_items(fulfillment_id);
CREATE INDEX idx_order_fulfillment_items_order_item_id ON order_fulfillment_items(order_item_id);

-- Create triggers for updated_at timestamps
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shopping_carts_updated_at BEFORE UPDATE ON shopping_carts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_shopping_cart_items_updated_at BEFORE UPDATE ON shopping_cart_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_order_fulfillments_updated_at BEFORE UPDATE ON order_fulfillments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to generate order numbers
CREATE OR REPLACE FUNCTION generate_order_number()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.order_number IS NULL THEN
        NEW.order_number := 'ORD-' || TO_CHAR(CURRENT_DATE, 'YYYYMMDD') || '-' || LPAD(nextval('order_number_seq')::TEXT, 6, '0');
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create sequence for order numbers
CREATE SEQUENCE IF NOT EXISTS order_number_seq START 1;

-- Create trigger for order number generation
CREATE TRIGGER generate_order_number_trigger BEFORE INSERT ON orders
    FOR EACH ROW EXECUTE FUNCTION generate_order_number();

-- Function to automatically create order status history
CREATE OR REPLACE FUNCTION create_order_status_history()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO order_status_history (id, order_id, to_status, reason, created_at)
        VALUES (gen_random_uuid()::TEXT, NEW.id, NEW.status, 'Order created', CURRENT_TIMESTAMP);
    ELSIF TG_OP = 'UPDATE' AND OLD.status != NEW.status THEN
        INSERT INTO order_status_history (id, order_id, from_status, to_status, reason, created_at)
        VALUES (gen_random_uuid()::TEXT, NEW.id, OLD.status, NEW.status, 'Status updated', CURRENT_TIMESTAMP);
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER order_status_history_trigger AFTER INSERT OR UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION create_order_status_history();