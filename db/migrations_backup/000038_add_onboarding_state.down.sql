-- Rollback onboarding tracking fields
DROP INDEX IF EXISTS idx_users_onboarding;
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_data;
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_skipped;
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_step;
ALTER TABLE users DROP COLUMN IF EXISTS onboarding_completed;
