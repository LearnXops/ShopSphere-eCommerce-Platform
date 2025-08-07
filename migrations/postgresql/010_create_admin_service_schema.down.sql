-- Admin Service Database Schema Rollback
-- This migration drops all admin service tables and related objects

-- Drop triggers first
DROP TRIGGER IF EXISTS update_admin_users_updated_at ON admin_users;
DROP TRIGGER IF EXISTS update_dashboard_configs_updated_at ON dashboard_configs;
DROP TRIGGER IF EXISTS update_system_alerts_updated_at ON system_alerts;
DROP TRIGGER IF EXISTS update_bulk_operations_updated_at ON bulk_operations;

-- Drop indexes
DROP INDEX IF EXISTS idx_admin_users_user_id;
DROP INDEX IF EXISTS idx_admin_users_role;
DROP INDEX IF EXISTS idx_admin_activity_logs_admin_user_id;
DROP INDEX IF EXISTS idx_admin_activity_logs_action;
DROP INDEX IF EXISTS idx_admin_activity_logs_resource_type;
DROP INDEX IF EXISTS idx_admin_activity_logs_created_at;
DROP INDEX IF EXISTS idx_system_metrics_metric_name;
DROP INDEX IF EXISTS idx_system_metrics_recorded_at;
DROP INDEX IF EXISTS idx_system_metrics_metric_type;
DROP INDEX IF EXISTS idx_dashboard_configs_admin_user_id;
DROP INDEX IF EXISTS idx_system_alerts_alert_type;
DROP INDEX IF EXISTS idx_system_alerts_severity;
DROP INDEX IF EXISTS idx_system_alerts_is_resolved;
DROP INDEX IF EXISTS idx_system_alerts_created_at;
DROP INDEX IF EXISTS idx_bulk_operations_admin_user_id;
DROP INDEX IF EXISTS idx_bulk_operations_status;
DROP INDEX IF EXISTS idx_bulk_operations_operation_type;
DROP INDEX IF EXISTS idx_bulk_operations_created_at;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS bulk_operations;
DROP TABLE IF EXISTS system_alerts;
DROP TABLE IF EXISTS dashboard_configs;
DROP TABLE IF EXISTS system_metrics;
DROP TABLE IF EXISTS admin_activity_logs;
DROP TABLE IF EXISTS admin_users;
