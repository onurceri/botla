-- Suggestion regeneration jobs table for async job tracking
CREATE TABLE suggestion_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,

    -- Job status
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error_message TEXT,

    -- Result (populated on completion)
    suggested_questions TEXT[],

    -- Timing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_status CHECK (status IN ('pending', 'running', 'completed', 'failed'))
);

-- Indexes for common queries
CREATE INDEX idx_suggestion_jobs_chatbot_id ON suggestion_jobs(chatbot_id);
CREATE INDEX idx_suggestion_jobs_status ON suggestion_jobs(status);
CREATE INDEX idx_suggestion_jobs_created_at ON suggestion_jobs(created_at DESC);

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_suggestion_jobs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER suggestion_jobs_updated_at
    BEFORE UPDATE ON suggestion_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_suggestion_jobs_updated_at();

-- Add comment
COMMENT ON TABLE suggestion_jobs IS 'Tracks async suggestion regeneration job execution for chatbots';
