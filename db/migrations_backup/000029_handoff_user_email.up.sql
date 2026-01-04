-- Add user_email column to handoff_requests for follow-up contact
ALTER TABLE handoff_requests
ADD COLUMN IF NOT EXISTS user_email TEXT;

COMMENT ON COLUMN handoff_requests.user_email IS 'User email for follow-up contact';
