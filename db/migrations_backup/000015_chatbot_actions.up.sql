CREATE TABLE IF NOT EXISTS chatbot_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    action_type TEXT NOT NULL, -- 'builtin', 'http', 'zapier'
    config JSONB NOT NULL DEFAULT '{}',
    parameters JSONB NOT NULL DEFAULT '{}', -- JSON Schema
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(chatbot_id, name)
);

CREATE INDEX idx_chatbot_actions_bot ON chatbot_actions(chatbot_id) WHERE enabled = true;
