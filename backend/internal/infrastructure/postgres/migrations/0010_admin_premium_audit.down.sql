DROP INDEX IF EXISTS idx_admin_audit_created_at;
DROP TABLE IF EXISTS admin_audit_log;
ALTER TABLE users DROP COLUMN IF EXISTS banned_at;
