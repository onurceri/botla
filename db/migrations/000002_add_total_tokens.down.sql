BEGIN;

ALTER TABLE analytics DROP COLUMN IF EXISTS total_tokens_used;

COMMIT;
