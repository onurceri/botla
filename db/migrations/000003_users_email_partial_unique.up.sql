-- Replace the plain UNIQUE(email) constraint with a partial unique index that only
-- enforces uniqueness among active (non soft-deleted) users. This lets a soft-deleted
-- user's email be reused on re-registration (the old constraint blocked it because it
-- ignored deleted_at). The index keeps the constraint name users_email_key.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX users_email_key ON users (email) WHERE deleted_at IS NULL;

-- Drop the now-redundant plain btree index on email (added in 000001). Active-email
-- lookups are served by the partial unique index above; the login query matches on
-- LOWER(email), which this btree never served anyway.
DROP INDEX IF EXISTS idx_users_email;
