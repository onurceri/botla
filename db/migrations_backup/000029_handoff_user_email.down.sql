-- Remove user_email column from handoff_requests
ALTER TABLE handoff_requests DROP COLUMN IF EXISTS user_email;
