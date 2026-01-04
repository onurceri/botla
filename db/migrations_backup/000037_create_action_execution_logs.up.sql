CREATE TABLE IF NOT EXISTS action_execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES chatbot_actions(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL,
    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_action_logs_chatbot_created ON action_execution_logs(chatbot_id, created_at DESC);
CREATE INDEX idx_action_logs_action_created ON action_execution_logs(action_id, created_at DESC);
