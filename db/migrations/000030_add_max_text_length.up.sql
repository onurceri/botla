-- Add max_text_length to plan files config
-- Default: 400,000 characters

-- Update Free plan
UPDATE plans SET config = jsonb_set(
    config,
    '{files,max_text_length}',
    '400000'::jsonb
) WHERE code = 'free';

-- Update Pro plan
UPDATE plans SET config = jsonb_set(
    config,
    '{files,max_text_length}',
    '400000'::jsonb
) WHERE code = 'pro';

-- Update Ultra plan
UPDATE plans SET config = jsonb_set(
    config,
    '{files,max_text_length}',
    '400000'::jsonb
) WHERE code = 'ultra';
