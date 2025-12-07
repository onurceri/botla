-- Create monthly ingestion usage table
CREATE TABLE IF NOT EXISTS usage_ingestions (
    user_id VARCHAR(64) NOT NULL,
    period_month DATE NOT NULL,
    sources_count INT NOT NULL DEFAULT 0,
    embedding_tokens INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, period_month)
);

-- Extend data_sources with hash, deleted_at, size_bytes
ALTER TABLE data_sources
    ADD COLUMN IF NOT EXISTS hash VARCHAR(128),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS size_bytes BIGINT DEFAULT 0;

-- Unique per chatbot on hash (avoid duplicate PDF/text); allow NULLs
CREATE UNIQUE INDEX IF NOT EXISTS ux_data_sources_chatbot_hash
    ON data_sources(chatbot_id, hash)
    WHERE hash IS NOT NULL AND deleted_at IS NULL;
