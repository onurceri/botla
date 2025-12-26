-- Add version column for optimistic locking to prevent race conditions
-- during concurrent action updates (especially when LLM calls are involved)
ALTER TABLE chatbot_actions ADD COLUMN version INT DEFAULT 1 NOT NULL;
