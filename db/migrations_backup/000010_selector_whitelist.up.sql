-- Add CSS selector whitelist for targeted content extraction
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS selector_whitelist TEXT[] DEFAULT '{}';

COMMENT ON COLUMN chatbots.selector_whitelist IS 'CSS selectors for content extraction (empty = full body)';
