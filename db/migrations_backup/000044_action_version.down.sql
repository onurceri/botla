-- Rollback: remove version column from chatbot_actions
ALTER TABLE chatbot_actions DROP COLUMN IF EXISTS version;
