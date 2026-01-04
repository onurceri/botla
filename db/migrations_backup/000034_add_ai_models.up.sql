CREATE TABLE IF NOT EXISTS ai_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    model_id VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    max_tokens INTEGER NOT NULL DEFAULT 4096,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed data
INSERT INTO ai_models (provider, model_id, name, max_tokens) VALUES
('OpenRouter', 'openai/gpt-4o-mini', 'GPT-4o Mini', 128000),
('OpenRouter', 'openai/gpt-4o', 'GPT-4o', 128000),
('OpenRouter', 'openai/gpt-5', 'GPT-5', 128000)
ON CONFLICT (model_id) DO UPDATE SET
    provider = EXCLUDED.provider,
    name = EXCLUDED.name,
    max_tokens = EXCLUDED.max_tokens;

-- Update Plans with new model configurations
-- Free: gpt-4o-mini
UPDATE plans SET config = jsonb_set(
    config,
    '{chat,allowed_models}',
    '["openai/gpt-4o-mini"]'::jsonb
) WHERE code = 'free';

-- Pro: gpt-4o-mini, gpt-4o
UPDATE plans SET config = jsonb_set(
    config,
    '{chat,allowed_models}',
    '["openai/gpt-4o-mini", "openai/gpt-4o"]'::jsonb
) WHERE code = 'pro';

-- Ultra: gpt-4o-mini, gpt-4o, gpt-5
UPDATE plans SET config = jsonb_set(
    config,
    '{chat,allowed_models}',
    '["openai/gpt-4o-mini", "openai/gpt-4o", "openai/gpt-5"]'::jsonb
) WHERE code = 'ultra';
