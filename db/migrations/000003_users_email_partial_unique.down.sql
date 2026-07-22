-- Reverse 000003: restore the plain UNIQUE(email) constraint and the btree index.
-- Note: this fails if soft-deleted rows share an email with an active row (a state the
-- original schema could not represent); that is expected for a rollback of this change.
DROP INDEX IF EXISTS users_email_key;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
CREATE INDEX idx_users_email ON users USING btree (email);
