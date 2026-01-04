DROP TABLE IF EXISTS platform_metrics;
DROP TABLE IF EXISTS error_logs;
DROP TABLE IF EXISTS admin_audit_logs;
DROP INDEX IF EXISTS idx_users_is_platform_admin;
ALTER TABLE users DROP COLUMN IF EXISTS is_platform_admin;
