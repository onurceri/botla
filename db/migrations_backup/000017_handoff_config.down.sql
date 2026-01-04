-- Drop handoff requests table
DROP TABLE IF EXISTS handoff_requests;

-- Remove handoff columns from chatbots
ALTER TABLE chatbots
DROP COLUMN IF EXISTS handoff_enabled,
DROP COLUMN IF EXISTS handoff_type,
DROP COLUMN IF EXISTS handoff_config;
