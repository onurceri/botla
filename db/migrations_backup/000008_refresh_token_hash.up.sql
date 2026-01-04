-- Rename token column to token_hash for secure storage
ALTER TABLE refresh_tokens RENAME COLUMN token TO token_hash;

-- Add index on token_hash for faster lookups
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
