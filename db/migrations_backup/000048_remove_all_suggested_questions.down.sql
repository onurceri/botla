BEGIN;

-- Re-add all_suggested_questions column for rollback
ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS all_suggested_questions JSONB DEFAULT '[]'::jsonb;

COMMIT;
