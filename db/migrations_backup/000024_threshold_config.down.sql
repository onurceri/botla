-- Revert threshold_config changes
ALTER TABLE chatbots DROP COLUMN IF EXISTS threshold_config;

-- Reset default confidence_threshold to original value
ALTER TABLE chatbots 
ALTER COLUMN confidence_threshold SET DEFAULT 0.7;
