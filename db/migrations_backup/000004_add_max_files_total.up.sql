BEGIN;

-- Update Free Plan Config
UPDATE plans 
SET config = jsonb_set(config, '{files,max_files_total}', '5'::jsonb)
WHERE code = 'free';

-- Update Pro Plan Config
UPDATE plans 
SET config = jsonb_set(config, '{files,max_files_total}', '100'::jsonb)
WHERE code = 'pro';

-- Update Ultra Plan Config
UPDATE plans 
SET config = jsonb_set(config, '{files,max_files_total}', '1000'::jsonb)
WHERE code = 'ultra';

COMMIT;
