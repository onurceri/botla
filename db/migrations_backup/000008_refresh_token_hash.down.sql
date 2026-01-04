-- Revert token_hash back to token
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
ALTER TABLE refresh_tokens RENAME COLUMN token_hash TO token;
