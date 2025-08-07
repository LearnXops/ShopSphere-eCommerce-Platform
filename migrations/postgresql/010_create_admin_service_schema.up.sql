-- Admin Service Database Schema
-- This migration creates tables for admin dashboard functionality

-- Admin users table (extends the main users table with admin-specific data)
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'admin',
    permissions JSONB NOT NULL DEFAULT '[]',
    department VARCHAR(100),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

-- Admin activity logs table
CREATE TABLE IF NOT EXISTS admin_activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- System metrics table for dashboard analytics
CREATE TABLE IF NOT EXISTS system_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_name VARCHAR(100) NOT NULL,
    metric_value DECIMAL(15,2) NOT NULL,
    metric_type VARCHAR(50) NOT NULL, -- 'counter', 'gauge', 'histogram'
    tags JSONB DEFAULT '{}',
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Dashboard configurations table
CREATE TABLE IF NOT EXISTS dashboard_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    dashboard_name VARCHAR(100) NOT NULL,
    config JSONB NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(admin_user_id, dashboard_name)
);

-- System alerts table
CREATE TABLE IF NOT EXISTS system_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info', -- 'critical', 'warning', 'info'
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    source_service VARCHAR(50),
    metadata JSONB DEFAULT '{}',
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_by UUID REFERENCES admin_users(id),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Bulk operations table for tracking bulk admin operations
CREATE TABLE IF NOT EXISTS bulk_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    operation_type VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    total_items INTEGER NOT NULL DEFAULT 0,
    processed_items INTEGER NOT NULL DEFAULT 0,
    failed_items INTEGER NOT NULL DEFAULT 0,
    parameters JSONB DEFAULT '{}',
    results JSONB DEFAULT '{}',
    error_message TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_admin_users_user_id ON admin_users(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
CREATE INDEX IF NOT EXISTS idx_admin_activity_logs_admin_user_id ON admin_activity_logs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_admin_activity_logs_action ON admin_activity_logs(action);
CREATE INDEX IF NOT EXISTS idx_admin_activity_logs_resource_type ON admin_activity_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_admin_activity_logs_created_at ON admin_activity_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_system_metrics_metric_name ON system_metrics(metric_name);
CREATE INDEX IF NOT EXISTS idx_system_metrics_recorded_at ON system_metrics(recorded_at);
CREATE INDEX IF NOT EXISTS idx_system_metrics_metric_type ON system_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_dashboard_configs_admin_user_id ON dashboard_configs(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_system_alerts_alert_type ON system_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_system_alerts_severity ON system_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_system_alerts_is_resolved ON system_alerts(is_resolved);
CREATE INDEX IF NOT EXISTS idx_system_alerts_created_at ON system_alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_admin_user_id ON bulk_operations(admin_user_id);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_status ON bulk_operations(status);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_operation_type ON bulk_operations(operation_type);
CREATE INDEX IF NOT EXISTS idx_bulk_operations_created_at ON bulk_operations(created_at);

-- Create triggers for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_admin_users_updated_at BEFORE UPDATE ON admin_users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_dashboard_configs_updated_at BEFORE UPDATE ON dashboard_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_system_alerts_updated_at BEFORE UPDATE ON system_alerts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_bulk_operations_updated_at BEFORE UPDATE ON bulk_operations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default admin roles and permissions
INSERT INTO admin_users (user_id, role, permissions, department) 
SELECT id, 'super_admin', 
       '["users:read", "users:write", "users:delete", "products:read", "products:write", "products:delete", "orders:read", "orders:write", "orders:delete", "analytics:read", "system:admin"]',
       'IT'
FROM users 
WHERE email = 'admin@shopsphere.com' 
ON CONFLICT (user_id) DO NOTHING;

-- Insert sample system metrics for dashboard
INSERT INTO system_metrics (metric_name, metric_value, metric_type, tags) VALUES
('total_users', 0, 'gauge', '{"category": "users"}'),
('total_products', 0, 'gauge', '{"category": "products"}'),
('total_orders', 0, 'gauge', '{"category": "orders"}'),
('daily_revenue', 0, 'gauge', '{"category": "revenue", "period": "daily"}'),
('monthly_revenue', 0, 'gauge', '{"category": "revenue", "period": "monthly"}'),
('active_sessions', 0, 'gauge', '{"category": "sessions"}'),
('cart_abandonment_rate', 0, 'gauge', '{"category": "conversion"}'),
('average_order_value', 0, 'gauge', '{"category": "orders"}');

-- Insert default dashboard configuration
INSERT INTO dashboard_configs (admin_user_id, dashboard_name, config, is_default)
SELECT au.id, 'default', 
       '{
         "widgets": [
           {"type": "metric", "title": "Total Users", "metric": "total_users", "position": {"x": 0, "y": 0, "w": 3, "h": 2}},
           {"type": "metric", "title": "Total Products", "metric": "total_products", "position": {"x": 3, "y": 0, "w": 3, "h": 2}},
           {"type": "metric", "title": "Total Orders", "metric": "total_orders", "position": {"x": 6, "y": 0, "w": 3, "h": 2}},
           {"type": "metric", "title": "Daily Revenue", "metric": "daily_revenue", "position": {"x": 9, "y": 0, "w": 3, "h": 2}},
           {"type": "chart", "title": "Revenue Trend", "metrics": ["daily_revenue"], "chartType": "line", "position": {"x": 0, "y": 2, "w": 6, "h": 4}},
           {"type": "chart", "title": "Order Status Distribution", "metrics": ["orders_by_status"], "chartType": "pie", "position": {"x": 6, "y": 2, "w": 6, "h": 4}}
         ]
       }', 
       true
FROM admin_users au
WHERE au.role = 'super_admin'
ON CONFLICT (admin_user_id, dashboard_name) DO NOTHING;
