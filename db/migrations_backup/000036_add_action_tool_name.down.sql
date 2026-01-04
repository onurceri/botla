DROP INDEX IF EXISTS idx_chatbot_actions_tool_name_unique;
ALTER TABLE chatbot_actions DROP COLUMN IF EXISTS tool_name;
