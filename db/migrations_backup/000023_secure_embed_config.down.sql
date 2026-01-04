BEGIN;

UPDATE plans 
SET config = config - 'security';

COMMIT;

