-- Rollback product service database schema

-- Drop triggers
DROP TRIGGER IF EXISTS inventory_movement_trigger ON inventory_movements;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_product_variants_updated_at ON product_variants;

-- Drop functions
DROP FUNCTION IF EXISTS update_product_stock();

-- Drop indexes
DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_categories_path;
DROP INDEX IF EXISTS idx_categories_is_active;
DROP INDEX IF EXISTS idx_products_sku;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_status;
DROP INDEX IF EXISTS idx_products_visibility;
DROP INDEX IF EXISTS idx_products_featured;
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_stock;
DROP INDEX IF EXISTS idx_products_attributes;
DROP INDEX IF EXISTS idx_product_variants_product_id;
DROP INDEX IF EXISTS idx_product_variants_sku;
DROP INDEX IF EXISTS idx_product_images_product_id;
DROP INDEX IF EXISTS idx_product_images_variant_id;
DROP INDEX IF EXISTS idx_product_tag_relations_product_id;
DROP INDEX IF EXISTS idx_product_tag_relations_tag_id;
DROP INDEX IF EXISTS idx_inventory_movements_product_id;
DROP INDEX IF EXISTS idx_inventory_movements_variant_id;
DROP INDEX IF EXISTS idx_inventory_movements_type;
DROP INDEX IF EXISTS idx_inventory_movements_created_at;

-- Drop tables
DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS product_tag_relations;
DROP TABLE IF EXISTS product_tags;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;