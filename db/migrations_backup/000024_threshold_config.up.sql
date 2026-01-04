-- Add threshold_config JSONB column for tiered confidence thresholds
ALTER TABLE chatbots 
ADD COLUMN IF NOT EXISTS threshold_config JSONB DEFAULT '{
  "high_threshold": 0.50,
  "medium_threshold": 0.30,
  "fallback_mode": "smart",
  "show_confidence_warning": true
}'::jsonb;

-- Migrate existing confidence_threshold values to the new config
-- If user had a custom threshold, use it as high_threshold
UPDATE chatbots 
SET threshold_config = jsonb_set(
  threshold_config, 
  '{high_threshold}', 
  to_jsonb(confidence_threshold)
)
WHERE confidence_threshold IS NOT NULL 
  AND confidence_threshold != 0.7;

-- Update default confidence_threshold to a more reasonable value
ALTER TABLE chatbots 
ALTER COLUMN confidence_threshold SET DEFAULT 0.35;

-- Add comment for documentation
COMMENT ON COLUMN chatbots.threshold_config IS 'Tiered threshold configuration: high_threshold (strong match), medium_threshold (weak match), fallback_mode (smart|static|escalate), show_confidence_warning (bool)';
