-- Add tool_name column for LLM-generated API-compatible identifier
ALTER TABLE chatbot_actions ADD COLUMN tool_name TEXT;

-- Add unique constraint: same chatbot can't have duplicate tool_names for enabled actions
CREATE UNIQUE INDEX idx_chatbot_actions_tool_name_unique 
ON chatbot_actions(chatbot_id, tool_name) 
WHERE enabled = true AND tool_name IS NOT NULL;
