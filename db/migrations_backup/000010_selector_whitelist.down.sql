-- Rollback selector whitelist column
ALTER TABLE chatbots DROP COLUMN IF EXISTS selector_whitelist;
