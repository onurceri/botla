BEGIN;

UPDATE plans SET config = '{}'::jsonb WHERE code IN ('free', 'pro', 'ultra');

COMMIT;
