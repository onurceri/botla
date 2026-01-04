BEGIN;

-- Remove manual_questions column from chatbots
ALTER TABLE chatbots DROP COLUMN IF EXISTS manual_questions;

-- Remove max_suggested_questions from plan configs
UPDATE plans SET config = config #- '{chat,max_suggested_questions}' WHERE code IN ('free', 'pro', 'ultra');

COMMIT;
