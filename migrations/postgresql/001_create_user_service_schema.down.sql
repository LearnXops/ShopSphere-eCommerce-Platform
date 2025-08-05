-- Rollback user service database schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_addresses_updated_at ON addresses;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_addresses_user_id;
DROP INDEX IF EXISTS idx_addresses_type;
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;
DROP INDEX IF EXISTS idx_password_reset_tokens_user_id;
DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP INDEX IF EXISTS idx_email_verification_tokens_user_id;

-- Drop tables
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS users;