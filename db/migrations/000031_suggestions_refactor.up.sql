BEGIN;

-- Add manual_questions column to chatbots for user-defined questions
ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS manual_questions JSONB DEFAULT '[]'::jsonb;

-- Update plan configs with max_suggested_questions
-- Free: 3, Pro: 6, Ultra: 10
UPDATE plans SET config = jsonb_set(config, '{chat,max_suggested_questions}', '3') WHERE code = 'free';
UPDATE plans SET config = jsonb_set(config, '{chat,max_suggested_questions}', '6') WHERE code = 'pro';
UPDATE plans SET config = jsonb_set(config, '{chat,max_suggested_questions}', '10') WHERE code = 'ultra';

COMMIT;
