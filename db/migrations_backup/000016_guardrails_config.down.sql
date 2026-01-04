ALTER TABLE chatbots
DROP COLUMN IF EXISTS confidence_threshold,
DROP COLUMN IF EXISTS fallback_messages,
DROP COLUMN IF EXISTS topic_restrictions;
