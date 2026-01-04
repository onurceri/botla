BEGIN;

-- Remove refresh config from plans
UPDATE plans SET config = config - 'refresh';

-- Remove refresh tracking columns
ALTER TABLE usage_ingestions DROP COLUMN IF EXISTS refresh_count;
ALTER TABLE data_sources DROP COLUMN IF EXISTS last_refreshed_at;

COMMIT;
