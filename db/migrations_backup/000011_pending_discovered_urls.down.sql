-- Remove discovery_mode column from chatbots
ALTER TABLE chatbots DROP COLUMN IF EXISTS discovery_mode;

-- Drop index
DROP INDEX IF EXISTS idx_pending_urls_chatbot;

-- Drop pending_discovered_urls table
DROP TABLE IF EXISTS pending_discovered_urls;
