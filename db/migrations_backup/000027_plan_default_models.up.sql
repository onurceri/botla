-- Add default_model to plan chat config for plan-based model selection
-- This enables OpenRouter as single provider with different models per plan tier

-- Update Free plan: gpt-4o-mini (cost-effective)
UPDATE plans SET config = jsonb_set(
    config,
    '{chat,default_model}',
    '"openai/gpt-4o-mini"'::jsonb
) WHERE code = 'free';

-- Update Pro plan: gpt-4o (better quality)
UPDATE plans SET config = jsonb_set(
    config,
    '{chat,default_model}',
    '"openai/gpt-4o"'::jsonb
) WHERE code = 'pro';

-- Update Ultra plan: claude-3.5-sonnet (best quality)
UPDATE plans SET config = jsonb_set(
    config,
    '{chat,default_model}',
    '"anthropic/claude-3.5-sonnet"'::jsonb
) WHERE code = 'ultra';
