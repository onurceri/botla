-- Rollback: Remove all_suggested_questions column
ALTER TABLE chatbots DROP COLUMN IF EXISTS all_suggested_questions;
