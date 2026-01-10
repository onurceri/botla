DROP INDEX IF EXISTS idx_privacy_requests_deleted_at;

ALTER TABLE privacy_requests DROP COLUMN IF EXISTS deleted_at;
