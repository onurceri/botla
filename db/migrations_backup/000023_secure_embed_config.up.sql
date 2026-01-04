BEGIN;

UPDATE plans 
SET config = jsonb_set(config, '{security}', '{"secure_embed_enabled": false}'::jsonb, true)
WHERE code = 'free';

UPDATE plans 
SET config = jsonb_set(config, '{security}', '{"secure_embed_enabled": true}'::jsonb, true)
WHERE code = 'pro';

UPDATE plans 
SET config = jsonb_set(config, '{security}', '{"secure_embed_enabled": true}'::jsonb, true)
WHERE code = 'ultra';

COMMIT;

