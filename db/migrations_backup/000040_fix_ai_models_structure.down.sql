-- Revert ai_models table structure changes

-- Add back model_id column
ALTER TABLE ai_models ADD COLUMN IF NOT EXISTS model_id VARCHAR(100);

-- Restore data from api_model_id
UPDATE ai_models SET model_id = api_model_id WHERE model_id IS NULL;

-- Make model_id NOT NULL
ALTER TABLE ai_models ALTER COLUMN model_id SET NOT NULL;

-- Restore unique constraint on model_id
ALTER TABLE ai_models DROP CONSTRAINT IF EXISTS ai_models_api_model_id_key;
ALTER TABLE ai_models ADD CONSTRAINT ai_models_model_id_key UNIQUE (model_id);

-- Drop new columns
DROP INDEX IF EXISTS idx_ai_models_model_name;
ALTER TABLE ai_models DROP COLUMN IF EXISTS model_name;
ALTER TABLE ai_models DROP COLUMN IF EXISTS api_model_id;

-- Revert plans to OpenRouter format
UPDATE plans SET config = jsonb_set(
    jsonb_set(
        config,
        '{chat,default_model}',
        '"openai/gpt-4o-mini"'::jsonb
    ),
    '{chat,allowed_models}',
    '["openai/gpt-4o-mini"]'::jsonb
) WHERE code = 'free';

UPDATE plans SET config = jsonb_set(
    jsonb_set(
        config,
        '{chat,default_model}',
        '"openai/gpt-4o"'::jsonb
    ),
    '{chat,allowed_models}',
    '["openai/gpt-4o-mini", "openai/gpt-4o"]'::jsonb
) WHERE code = 'pro';

UPDATE plans SET config = jsonb_set(
    jsonb_set(
        config,
        '{chat,default_model}',
        '"openai/gpt-4o"'::jsonb
    ),
    '{chat,allowed_models}',
    '["openai/gpt-4o-mini", "openai/gpt-4o", "openai/gpt-5"]'::jsonb
) WHERE code = 'ultra';
