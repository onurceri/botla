BEGIN;

-- Add max_manual_questions limit to plan configs
-- Same values as max_suggested_questions: Free=3, Pro=6, Ultra=10
UPDATE plans SET config = jsonb_set(config, '{chat,max_manual_questions}', '3') WHERE code = 'free';
UPDATE plans SET config = jsonb_set(config, '{chat,max_manual_questions}', '6') WHERE code = 'pro';
UPDATE plans SET config = jsonb_set(config, '{chat,max_manual_questions}', '10') WHERE code = 'ultra';

COMMIT;
