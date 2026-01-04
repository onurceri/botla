ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS confidence_threshold FLOAT DEFAULT 0.7,
ADD COLUMN IF NOT EXISTS fallback_messages JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS topic_restrictions JSONB DEFAULT '{}';

COMMENT ON COLUMN chatbots.confidence_threshold IS 'Minimum RAG score to provide answer (0.0-1.0)';
