BEGIN;

ALTER TABLE analytics DROP COLUMN total_tokens_used;

COMMIT;
