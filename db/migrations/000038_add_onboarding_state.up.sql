-- Add onboarding tracking fields to users table
ALTER TABLE users ADD COLUMN onboarding_completed BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN onboarding_step INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN onboarding_skipped BOOLEAN DEFAULT false;
ALTER TABLE users ADD COLUMN onboarding_data JSONB DEFAULT '{}'::jsonb;

-- Create index for querying users who need onboarding
CREATE INDEX idx_users_onboarding ON users(onboarding_completed, onboarding_skipped) WHERE deleted_at IS NULL;
