BEGIN;

-- Add refresh tracking column to data_sources
ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS last_refreshed_at TIMESTAMPTZ;

-- Add refresh count to usage tracking
ALTER TABLE usage_ingestions ADD COLUMN IF NOT EXISTS refresh_count INT DEFAULT 0;

-- Update Free Plan Config - refresh disabled
UPDATE plans 
SET config = jsonb_set(
    config, 
    '{refresh}', 
    '{"enabled": false, "max_monthly": 0}'::jsonb
)
WHERE code = 'free';

-- Update Pro Plan Config - 5 refreshes/month
UPDATE plans 
SET config = jsonb_set(
    config, 
    '{refresh}', 
    '{"enabled": true, "max_monthly": 5}'::jsonb
)
WHERE code = 'pro';

-- Update Ultra Plan Config - 10 refreshes/month
UPDATE plans 
SET config = jsonb_set(
    config, 
    '{refresh}', 
    '{"enabled": true, "max_monthly": 10}'::jsonb
)
WHERE code = 'ultra';

COMMIT;
