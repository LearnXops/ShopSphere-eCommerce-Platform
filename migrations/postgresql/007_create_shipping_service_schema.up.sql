-- Create shipping service database schema

-- Shipping methods table
CREATE TABLE shipping_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    carrier_name VARCHAR(50) NOT NULL,
    service_type VARCHAR(50) NOT NULL,
    delivery_time_min INTEGER NOT NULL, -- in hours
    delivery_time_max INTEGER NOT NULL, -- in hours
    base_cost DECIMAL(10,2) NOT NULL,
    cost_per_kg DECIMAL(10,2) DEFAULT 0,
    cost_per_km DECIMAL(10,2) DEFAULT 0,
    max_weight_kg DECIMAL(8,2),
    max_dimensions_cm VARCHAR(50), -- format: "LxWxH"
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Shipping zones table
CREATE TABLE shipping_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    countries TEXT[] NOT NULL, -- array of country codes
    states TEXT[], -- array of state/province codes
    postal_codes TEXT[], -- array of postal code patterns
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Shipping rates table (method + zone specific rates)
CREATE TABLE shipping_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipping_method_id UUID NOT NULL REFERENCES shipping_methods(id) ON DELETE CASCADE,
    shipping_zone_id UUID NOT NULL REFERENCES shipping_zones(id) ON DELETE CASCADE,
    min_weight_kg DECIMAL(8,2) DEFAULT 0,
    max_weight_kg DECIMAL(8,2),
    min_order_value DECIMAL(10,2) DEFAULT 0,
    max_order_value DECIMAL(10,2),
    rate DECIMAL(10,2) NOT NULL,
    free_shipping_threshold DECIMAL(10,2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(shipping_method_id, shipping_zone_id, min_weight_kg, min_order_value)
);

-- Shipments table
CREATE TABLE shipments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    shipping_method_id UUID NOT NULL REFERENCES shipping_methods(id),
    tracking_number VARCHAR(100) UNIQUE,
    carrier_tracking_id VARCHAR(100),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    
    -- Shipping addresses
    from_address JSONB NOT NULL,
    to_address JSONB NOT NULL,
    
    -- Package details
    weight_kg DECIMAL(8,2) NOT NULL,
    dimensions_cm VARCHAR(50), -- format: "LxWxH"
    declared_value DECIMAL(10,2),
    
    -- Costs
    shipping_cost DECIMAL(10,2) NOT NULL,
    insurance_cost DECIMAL(10,2) DEFAULT 0,
    total_cost DECIMAL(10,2) NOT NULL,
    
    -- Timing
    estimated_delivery_date TIMESTAMP WITH TIME ZONE,
    actual_pickup_date TIMESTAMP WITH TIME ZONE,
    actual_delivery_date TIMESTAMP WITH TIME ZONE,
    
    -- Carrier integration
    carrier_response JSONB,
    label_url TEXT,
    
    -- Metadata
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Shipment tracking events table
CREATE TABLE shipment_tracking_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    description TEXT,
    location VARCHAR(200),
    event_time TIMESTAMP WITH TIME ZONE NOT NULL,
    carrier_event_id VARCHAR(100),
    carrier_raw_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Carrier configurations table
CREATE TABLE carrier_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carrier_name VARCHAR(50) NOT NULL UNIQUE,
    api_endpoint TEXT NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    api_secret_encrypted TEXT,
    webhook_secret TEXT,
    is_active BOOLEAN DEFAULT true,
    rate_limit_per_minute INTEGER DEFAULT 60,
    timeout_seconds INTEGER DEFAULT 30,
    config_data JSONB, -- carrier-specific configuration
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_shipping_methods_carrier ON shipping_methods(carrier_name);
CREATE INDEX idx_shipping_methods_active ON shipping_methods(is_active);
CREATE INDEX idx_shipping_zones_countries ON shipping_zones USING GIN(countries);
CREATE INDEX idx_shipping_rates_method_zone ON shipping_rates(shipping_method_id, shipping_zone_id);
CREATE INDEX idx_shipping_rates_weight ON shipping_rates(min_weight_kg, max_weight_kg);
CREATE INDEX idx_shipments_order ON shipments(order_id);
CREATE INDEX idx_shipments_user ON shipments(user_id);
CREATE INDEX idx_shipments_tracking ON shipments(tracking_number);
CREATE INDEX idx_shipments_status ON shipments(status);
CREATE INDEX idx_shipments_created ON shipments(created_at);
CREATE INDEX idx_tracking_events_shipment ON shipment_tracking_events(shipment_id);
CREATE INDEX idx_tracking_events_time ON shipment_tracking_events(event_time);
CREATE INDEX idx_carrier_configs_name ON carrier_configs(carrier_name);

-- Create triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_shipping_methods_updated_at BEFORE UPDATE ON shipping_methods FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_shipping_zones_updated_at BEFORE UPDATE ON shipping_zones FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_shipping_rates_updated_at BEFORE UPDATE ON shipping_rates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_shipments_updated_at BEFORE UPDATE ON shipments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_carrier_configs_updated_at BEFORE UPDATE ON carrier_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default shipping zones
INSERT INTO shipping_zones (name, description, countries) VALUES
('Domestic', 'Domestic shipping within the country', ARRAY['US']),
('North America', 'Shipping to North American countries', ARRAY['US', 'CA', 'MX']),
('Europe', 'Shipping to European countries', ARRAY['GB', 'DE', 'FR', 'IT', 'ES', 'NL', 'BE', 'AT', 'CH']),
('Asia Pacific', 'Shipping to Asia Pacific region', ARRAY['JP', 'AU', 'SG', 'HK', 'KR', 'TW']),
('International', 'Worldwide shipping', ARRAY['*']);

-- Insert default shipping methods
INSERT INTO shipping_methods (name, description, carrier_name, service_type, delivery_time_min, delivery_time_max, base_cost, cost_per_kg) VALUES
('Standard Ground', 'Standard ground shipping', 'FedEx', 'GROUND', 72, 120, 9.99, 2.50),
('Express 2-Day', 'Express 2-day delivery', 'FedEx', 'EXPRESS_2_DAY', 24, 48, 19.99, 3.50),
('Overnight', 'Next day delivery', 'FedEx', 'OVERNIGHT', 12, 24, 39.99, 5.00),
('UPS Ground', 'UPS standard ground shipping', 'UPS', 'GROUND', 72, 120, 8.99, 2.25),
('UPS 2nd Day Air', 'UPS 2-day air delivery', 'UPS', '2ND_DAY_AIR', 24, 48, 18.99, 3.25),
('UPS Next Day Air', 'UPS next day air delivery', 'UPS', 'NEXT_DAY_AIR', 12, 24, 38.99, 4.75),
('USPS Priority', 'USPS Priority Mail', 'USPS', 'PRIORITY', 24, 72, 7.99, 1.50),
('USPS Express', 'USPS Priority Mail Express', 'USPS', 'EXPRESS', 12, 24, 29.99, 3.00);

-- Insert default shipping rates
INSERT INTO shipping_rates (shipping_method_id, shipping_zone_id, min_weight_kg, max_weight_kg, rate, free_shipping_threshold) 
SELECT 
    sm.id,
    sz.id,
    0,
    50,
    CASE 
        WHEN sz.name = 'Domestic' THEN sm.base_cost
        WHEN sz.name = 'North America' THEN sm.base_cost * 1.5
        WHEN sz.name = 'Europe' THEN sm.base_cost * 2.0
        WHEN sz.name = 'Asia Pacific' THEN sm.base_cost * 2.5
        ELSE sm.base_cost * 3.0
    END,
    CASE 
        WHEN sz.name = 'Domestic' THEN 75.00
        WHEN sz.name = 'North America' THEN 100.00
        ELSE 150.00
    END
FROM shipping_methods sm
CROSS JOIN shipping_zones sz
WHERE sm.carrier_name IN ('FedEx', 'UPS', 'USPS');
