BEGIN;

-- Remove all_suggested_questions column (redundant with suggested_questions)
ALTER TABLE chatbots DROP COLUMN IF EXISTS all_suggested_questions;

COMMIT;
