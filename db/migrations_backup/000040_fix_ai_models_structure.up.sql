-- Restructure ai_models table to separate model_name (bare) from api_model_id (full OpenRouter format)
-- This eliminates the need for runtime parsing of provider prefixes (/, :)

-- Add the new columns
ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS model_name VARCHAR(100);
ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS api_model_id VARCHAR(150);

-- Populate new columns from existing data
-- model_id currently stores OpenRouter format like "openai/gpt-4o-mini"
-- Extract the bare model name and keep the full api_model_id
UPDATE ai_models SET
    model_name = CASE 
        WHEN model_id LIKE '%/%' THEN split_part(model_id, '/', 2)
        WHEN model_id LIKE '%:%' THEN split_part(model_id, ':', 2)
        ELSE model_id
    END,
    api_model_id = model_id
WHERE model_name IS NULL OR api_model_id IS NULL;

-- Make model_name NOT NULL after populating
ALTER TABLE ai_models ALTER COLUMN model_name SET NOT NULL;
ALTER TABLE ai_models ALTER COLUMN api_model_id SET NOT NULL;

-- Add unique constraint on api_model_id (replace the one on model_id)
ALTER TABLE ai_models DROP CONSTRAINT IF EXISTS ai_models_model_id_key;
ALTER TABLE ai_models ADD CONSTRAINT ai_models_api_model_id_key UNIQUE (api_model_id);

-- Add index on model_name for efficient lookups
CREATE INDEX IF NOT EXISTS idx_ai_models_model_name ON ai_models(model_name);

-- Drop the old model_id column as it's now redundant
ALTER TABLE ai_models DROP COLUMN IF EXISTS model_id;

-- Update plans table to use new format (model_name instead of provider/model_id)
-- Free: gpt-4o-mini
UPDATE plans SET config = jsonb_set(
    jsonb_set(
        config,
        '{chat,default_model}',
        '"gpt-4o-mini"'::jsonb
    ),
    '{chat,allowed_models}',
    '["gpt-4o-mini"]'::jsonb
) WHERE code = 'free';

-- Pro: gpt-4o-mini, gpt-4o
UPDATE plans SET config = jsonb_set(
    jsonb_set(
        config,
        '{chat,default_model}',
        '"gpt-4o"'::jsonb
    ),
    '{chat,allowed_models}',
    '["gpt-4o-mini", "gpt-4o"]'::jsonb
) WHERE code = 'pro';

-- Ultra: gpt-4o-mini, gpt-4o, gpt-5
UPDATE plans SET config = jsonb_set(
    jsonb_set(
        config,
        '{chat,default_model}',
        '"gpt-4o"'::jsonb
    ),
    '{chat,allowed_models}',
    '["gpt-4o-mini", "gpt-4o", "gpt-5"]'::jsonb
) WHERE code = 'ultra';
