-- Add token tracking column to analytics
ALTER TABLE analytics 
ADD COLUMN IF NOT EXISTS total_tokens_used INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS handoff_count INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS avg_response_time_ms INTEGER;

-- Create index for efficient trend queries
CREATE INDEX IF NOT EXISTS idx_analytics_date_range 
ON analytics(chatbot_id, analytics_date DESC);

-- Message-level analytics
ALTER TABLE messages
ADD COLUMN IF NOT EXISTS confidence_score FLOAT,
ADD COLUMN IF NOT EXISTS sources_used UUID[] DEFAULT '{}';

-- Unanswered queries tracking
CREATE TABLE IF NOT EXISTS unanswered_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    occurrence_count INTEGER DEFAULT 1,
    last_occurred_at TIMESTAMP DEFAULT NOW(),
    addressed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(chatbot_id, query)
);

CREATE INDEX idx_unanswered_chatbot ON unanswered_queries(chatbot_id, addressed);
