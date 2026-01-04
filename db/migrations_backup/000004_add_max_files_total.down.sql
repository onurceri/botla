BEGIN;

-- Remove max_files_total from config
UPDATE plans
SET config = config #- '{files,max_files_total}';

COMMIT;
