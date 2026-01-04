DROP TABLE IF EXISTS usage_ingestions;

ALTER TABLE data_sources
    DROP COLUMN IF EXISTS hash,
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS size_bytes;

DROP INDEX IF EXISTS ux_data_sources_chatbot_hash;

