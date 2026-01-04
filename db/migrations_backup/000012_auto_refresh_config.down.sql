BEGIN;

-- Remove index
DROP INDEX IF EXISTS idx_chatbots_next_refresh;

-- Remove auto_refresh_count from usage_ingestions
ALTER TABLE usage_ingestions
DROP COLUMN IF EXISTS auto_refresh_count;

-- Remove refresh policy columns from chatbots
ALTER TABLE chatbots
DROP COLUMN IF EXISTS refresh_policy,
DROP COLUMN IF EXISTS refresh_frequency,
DROP COLUMN IF EXISTS next_refresh_at,
DROP COLUMN IF EXISTS last_refresh_at;

COMMIT;
