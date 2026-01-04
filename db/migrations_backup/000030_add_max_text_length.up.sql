-- Add max_text_length to plan files config
-- Free: 10,000 chars (~2-3 pages), Pro: 100,000 chars (~25 pages), Ultra: 400,000 chars

-- Update Free plan
UPDATE plans SET config = jsonb_set(
    config,
    '{files,max_text_length}',
    '10000'::jsonb
) WHERE code = 'free';

-- Update Pro plan
UPDATE plans SET config = jsonb_set(
    config,
    '{files,max_text_length}',
    '100000'::jsonb
) WHERE code = 'pro';

-- Update Ultra plan
UPDATE plans SET config = jsonb_set(
    config,
    '{files,max_text_length}',
    '400000'::jsonb
) WHERE code = 'ultra';
