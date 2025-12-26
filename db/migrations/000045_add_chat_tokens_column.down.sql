-- Remove chat_tokens column from usage_ingestions
ALTER TABLE usage_ingestions
DROP COLUMN IF EXISTS chat_tokens;
