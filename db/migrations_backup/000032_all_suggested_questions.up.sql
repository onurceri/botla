-- Add all_suggested_questions to store all LLM-generated questions
-- suggested_questions remains for user-selected visible questions
ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS all_suggested_questions JSONB DEFAULT '[]';

-- Migrate existing suggested_questions to all_suggested_questions
UPDATE chatbots 
SET all_suggested_questions = suggested_questions 
WHERE suggested_questions IS NOT NULL 
  AND suggested_questions != '[]'::jsonb 
  AND (all_suggested_questions IS NULL OR all_suggested_questions = '[]'::jsonb);
