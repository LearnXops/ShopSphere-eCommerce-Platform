-- Drop notification service schema
-- This migration removes all notification-related tables and functions

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS notification_retry_queue CASCADE;
DROP TABLE IF EXISTS notification_delivery_events CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS notification_preferences CASCADE;
DROP TABLE IF EXISTS notification_templates CASCADE;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_notification_templates_updated_at ON notification_templates;
DROP TRIGGER IF EXISTS trigger_notification_preferences_updated_at ON notification_preferences;
DROP TRIGGER IF EXISTS trigger_notifications_updated_at ON notifications;

-- Drop functions
DROP FUNCTION IF EXISTS update_notification_updated_at();

-- Drop indexes (they will be dropped automatically with tables, but listing for clarity)
-- DROP INDEX IF EXISTS idx_notification_templates_channel;
-- DROP INDEX IF EXISTS idx_notification_templates_name;
-- DROP INDEX IF EXISTS idx_notification_preferences_user_id;
-- DROP INDEX IF EXISTS idx_notification_preferences_user_channel;
-- DROP INDEX IF EXISTS idx_notifications_user_id;
-- DROP INDEX IF EXISTS idx_notifications_status;
-- DROP INDEX IF EXISTS idx_notifications_channel;
-- DROP INDEX IF EXISTS idx_notifications_scheduled_at;
-- DROP INDEX IF EXISTS idx_notifications_created_at;
-- DROP INDEX IF EXISTS idx_notification_delivery_events_notification_id;
-- DROP INDEX IF EXISTS idx_notification_delivery_events_event_type;
-- DROP INDEX IF EXISTS idx_notification_retry_queue_scheduled_for;
-- DROP INDEX IF EXISTS idx_notification_retry_queue_notification_id;
