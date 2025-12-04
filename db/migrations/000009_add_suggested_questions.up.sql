ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS suggested_questions JSONB;
ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS suggested_questions JSONB;
