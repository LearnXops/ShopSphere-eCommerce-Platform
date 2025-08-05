-- Create product service database schema
-- This migration creates the product catalog and inventory management tables

-- Categories table with hierarchical structure
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id VARCHAR(36) REFERENCES categories(id) ON DELETE SET NULL,
    path VARCHAR(500) NOT NULL, -- materialized path for efficient queries
    level INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    short_description TEXT,
    category_id VARCHAR(36) REFERENCES categories(id) ON DELETE SET NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    compare_price DECIMAL(10,2) CHECK (compare_price >= 0), -- original price for discounts
    cost_price DECIMAL(10,2) CHECK (cost_price >= 0), -- cost for profit calculations
    currency VARCHAR(3) DEFAULT 'USD',
    stock INTEGER DEFAULT 0 CHECK (stock >= 0),
    reserved_stock INTEGER DEFAULT 0 CHECK (reserved_stock >= 0), -- stock reserved in carts
    low_stock_threshold INTEGER DEFAULT 10,
    track_inventory BOOLEAN DEFAULT TRUE,
    status VARCHAR(20) DEFAULT 'inactive' CHECK (status IN ('active', 'inactive', 'out_of_stock', 'discontinued')),
    visibility VARCHAR(20) DEFAULT 'visible' CHECK (visibility IN ('visible', 'hidden', 'catalog_only')),
    weight DECIMAL(8,3), -- in kg
    length DECIMAL(8,2), -- in cm
    width DECIMAL(8,2),  -- in cm
    height DECIMAL(8,2), -- in cm
    images TEXT[], -- Array of image URLs
    attributes JSONB DEFAULT '{}', -- flexible attributes storage
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords TEXT,
    featured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product variants table for products with variations (size, color, etc.)
CREATE TABLE IF NOT EXISTS product_variants (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    compare_price DECIMAL(10,2) CHECK (compare_price >= 0),
    cost_price DECIMAL(10,2) CHECK (cost_price >= 0),
    stock INTEGER DEFAULT 0 CHECK (stock >= 0),
    reserved_stock INTEGER DEFAULT 0 CHECK (reserved_stock >= 0),
    weight DECIMAL(8,3),
    attributes JSONB DEFAULT '{}', -- variant-specific attributes
    image_url TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product images table for multiple images per product
CREATE TABLE IF NOT EXISTS product_images (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id VARCHAR(36) REFERENCES product_variants(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    alt_text VARCHAR(255),
    sort_order INTEGER DEFAULT 0,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product tags table for flexible tagging system
CREATE TABLE IF NOT EXISTS product_tags (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product-tag relationship table
CREATE TABLE IF NOT EXISTS product_tag_relations (
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tag_id VARCHAR(36) NOT NULL REFERENCES product_tags(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (product_id, tag_id)
);

-- Inventory movements table for tracking stock changes
CREATE TABLE IF NOT EXISTS inventory_movements (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    variant_id VARCHAR(36) REFERENCES product_variants(id) ON DELETE CASCADE,
    movement_type VARCHAR(20) NOT NULL CHECK (movement_type IN ('in', 'out', 'adjustment', 'reserved', 'released')),
    quantity INTEGER NOT NULL,
    reference_type VARCHAR(50), -- order, adjustment, etc.
    reference_id VARCHAR(36),
    reason TEXT,
    created_by VARCHAR(36), -- user who made the change
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_path ON categories(path);
CREATE INDEX idx_categories_is_active ON categories(is_active);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_visibility ON products(visibility);
CREATE INDEX idx_products_featured ON products(featured);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_stock ON products(stock);
CREATE INDEX idx_products_attributes ON products USING GIN(attributes);
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_sku ON product_variants(sku);
CREATE INDEX idx_product_images_product_id ON product_images(product_id);
CREATE INDEX idx_product_images_variant_id ON product_images(variant_id);
CREATE INDEX idx_product_tag_relations_product_id ON product_tag_relations(product_id);
CREATE INDEX idx_product_tag_relations_tag_id ON product_tag_relations(tag_id);
CREATE INDEX idx_inventory_movements_product_id ON inventory_movements(product_id);
CREATE INDEX idx_inventory_movements_variant_id ON inventory_movements(variant_id);
CREATE INDEX idx_inventory_movements_type ON inventory_movements(movement_type);
CREATE INDEX idx_inventory_movements_created_at ON inventory_movements(created_at);

-- Create triggers for updated_at timestamps
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_product_variants_updated_at BEFORE UPDATE ON product_variants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to update product stock based on inventory movements
CREATE OR REPLACE FUNCTION update_product_stock()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.movement_type = 'in' THEN
        UPDATE products SET stock = stock + NEW.quantity WHERE id = NEW.product_id;
        IF NEW.variant_id IS NOT NULL THEN
            UPDATE product_variants SET stock = stock + NEW.quantity WHERE id = NEW.variant_id;
        END IF;
    ELSIF NEW.movement_type = 'out' THEN
        UPDATE products SET stock = stock - NEW.quantity WHERE id = NEW.product_id;
        IF NEW.variant_id IS NOT NULL THEN
            UPDATE product_variants SET stock = stock - NEW.quantity WHERE id = NEW.variant_id;
        END IF;
    ELSIF NEW.movement_type = 'reserved' THEN
        UPDATE products SET reserved_stock = reserved_stock + NEW.quantity WHERE id = NEW.product_id;
        IF NEW.variant_id IS NOT NULL THEN
            UPDATE product_variants SET reserved_stock = reserved_stock + NEW.quantity WHERE id = NEW.variant_id;
        END IF;
    ELSIF NEW.movement_type = 'released' THEN
        UPDATE products SET reserved_stock = reserved_stock - NEW.quantity WHERE id = NEW.product_id;
        IF NEW.variant_id IS NOT NULL THEN
            UPDATE product_variants SET reserved_stock = reserved_stock - NEW.quantity WHERE id = NEW.variant_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER inventory_movement_trigger AFTER INSERT ON inventory_movements
    FOR EACH ROW EXECUTE FUNCTION update_product_stock();