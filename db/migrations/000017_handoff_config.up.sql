-- Add handoff configuration to chatbots
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS handoff_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS handoff_type TEXT DEFAULT 'email',
ADD COLUMN IF NOT EXISTS handoff_config JSONB DEFAULT '{}';

COMMENT ON COLUMN chatbots.handoff_enabled IS 'Whether human handoff is enabled for this chatbot';
COMMENT ON COLUMN chatbots.handoff_type IS 'Type of handoff: email';
COMMENT ON COLUMN chatbots.handoff_config IS 'Configuration for handoff (email_to, email_subject, etc.)';

-- Create handoff requests tracking table
CREATE TABLE IF NOT EXISTS handoff_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    status TEXT DEFAULT 'pending',
    assigned_to TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    resolved_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_handoff_requests_chatbot ON handoff_requests(chatbot_id);
CREATE INDEX IF NOT EXISTS idx_handoff_requests_status ON handoff_requests(status);
CREATE INDEX IF NOT EXISTS idx_handoff_requests_created ON handoff_requests(created_at DESC);
