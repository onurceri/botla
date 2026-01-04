BEGIN;

-- Add refresh policy columns to chatbots
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS refresh_policy TEXT DEFAULT 'manual',
ADD COLUMN IF NOT EXISTS refresh_frequency TEXT DEFAULT NULL,
ADD COLUMN IF NOT EXISTS next_refresh_at TIMESTAMPTZ DEFAULT NULL,
ADD COLUMN IF NOT EXISTS last_refresh_at TIMESTAMPTZ DEFAULT NULL;

-- refresh_policy: 'manual', 'auto'
-- refresh_frequency: 'daily', 'weekly', 'monthly' (only used when policy is 'auto')

COMMENT ON COLUMN chatbots.refresh_policy IS 'manual or auto';
COMMENT ON COLUMN chatbots.refresh_frequency IS 'daily, weekly, or monthly (only for auto policy)';
COMMENT ON COLUMN chatbots.next_refresh_at IS 'Next scheduled auto-refresh time';
COMMENT ON COLUMN chatbots.last_refresh_at IS 'Last auto-refresh execution time';

-- Add auto_refresh_count to usage_ingestions for tracking monthly auto-refresh usage
ALTER TABLE usage_ingestions
ADD COLUMN IF NOT EXISTS auto_refresh_count INT DEFAULT 0;

-- Index for efficient scheduler queries (finding chatbots due for refresh)
CREATE INDEX IF NOT EXISTS idx_chatbots_next_refresh 
ON chatbots(next_refresh_at) 
WHERE refresh_policy = 'auto' AND deleted_at IS NULL;

COMMIT;
