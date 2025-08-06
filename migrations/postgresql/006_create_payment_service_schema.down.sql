-- Drop Payment Service Database Schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_payment_methods_updated_at ON payment_methods;
DROP TRIGGER IF EXISTS update_refunds_updated_at ON refunds;

-- Drop indexes
DROP INDEX IF EXISTS idx_payments_order_id;
DROP INDEX IF EXISTS idx_payments_user_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_transaction_id;

DROP INDEX IF EXISTS idx_payment_methods_user_id;
DROP INDEX IF EXISTS idx_payment_methods_type;
DROP INDEX IF EXISTS idx_payment_methods_is_default;
DROP INDEX IF EXISTS idx_payment_methods_stripe_id;
DROP INDEX IF EXISTS idx_payment_methods_user_default;

DROP INDEX IF EXISTS idx_refunds_payment_id;
DROP INDEX IF EXISTS idx_refunds_order_id;
DROP INDEX IF EXISTS idx_refunds_status;
DROP INDEX IF EXISTS idx_refunds_created_at;

DROP INDEX IF EXISTS idx_payment_webhooks_event_type;
DROP INDEX IF EXISTS idx_payment_webhooks_payment_id;
DROP INDEX IF EXISTS idx_payment_webhooks_processed;
DROP INDEX IF EXISTS idx_payment_webhooks_created_at;

DROP INDEX IF EXISTS idx_payment_attempts_payment_id;
DROP INDEX IF EXISTS idx_payment_attempts_status;

-- Drop tables in correct order (respecting foreign key constraints)
DROP TABLE IF EXISTS payment_attempts;
DROP TABLE IF EXISTS payment_webhooks;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS payments;
