-- Training jobs table for async job tracking
CREATE TABLE training_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    
    -- Job status
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    current_step VARCHAR(50),
    progress_percent INTEGER DEFAULT 0,
    
    -- Error tracking
    error_code VARCHAR(100),
    error_message TEXT,
    failed_step VARCHAR(50),
    retry_count INTEGER DEFAULT 0,
    
    -- Timing
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    CONSTRAINT valid_status CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled'))
);

-- Indexes for common queries
CREATE INDEX idx_training_jobs_source_id ON training_jobs(source_id);
CREATE INDEX idx_training_jobs_chatbot_id ON training_jobs(chatbot_id);
CREATE INDEX idx_training_jobs_status ON training_jobs(status);
CREATE INDEX idx_training_jobs_created_at ON training_jobs(created_at DESC);

-- Index for finding jobs to retry
CREATE INDEX idx_training_jobs_retry ON training_jobs(status, retry_count) 
    WHERE status = 'failed' AND retry_count < 3;

-- Updated at trigger
CREATE OR REPLACE FUNCTION update_training_jobs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER training_jobs_updated_at
    BEFORE UPDATE ON training_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_training_jobs_updated_at();

-- Add comment
COMMENT ON TABLE training_jobs IS 'Tracks async training job execution for data sources';
