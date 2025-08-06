-- Drop shipping service database schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_shipping_methods_updated_at ON shipping_methods;
DROP TRIGGER IF EXISTS update_shipping_zones_updated_at ON shipping_zones;
DROP TRIGGER IF EXISTS update_shipping_rates_updated_at ON shipping_rates;
DROP TRIGGER IF EXISTS update_shipments_updated_at ON shipments;
DROP TRIGGER IF EXISTS update_carrier_configs_updated_at ON carrier_configs;

-- Drop indexes
DROP INDEX IF EXISTS idx_shipping_methods_carrier;
DROP INDEX IF EXISTS idx_shipping_methods_active;
DROP INDEX IF EXISTS idx_shipping_zones_countries;
DROP INDEX IF EXISTS idx_shipping_rates_method_zone;
DROP INDEX IF EXISTS idx_shipping_rates_weight;
DROP INDEX IF EXISTS idx_shipments_order;
DROP INDEX IF EXISTS idx_shipments_user;
DROP INDEX IF EXISTS idx_shipments_tracking;
DROP INDEX IF EXISTS idx_shipments_status;
DROP INDEX IF EXISTS idx_shipments_created;
DROP INDEX IF EXISTS idx_tracking_events_shipment;
DROP INDEX IF EXISTS idx_tracking_events_time;
DROP INDEX IF EXISTS idx_carrier_configs_name;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS shipment_tracking_events;
DROP TABLE IF EXISTS shipments;
DROP TABLE IF EXISTS shipping_rates;
DROP TABLE IF EXISTS carrier_configs;
DROP TABLE IF EXISTS shipping_zones;
DROP TABLE IF EXISTS shipping_methods;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();
