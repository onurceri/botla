-- Rollback: Re-add ocr_enabled to plan configs
-- Note: All plans will default to ocr_enabled: false

UPDATE plans 
SET config = jsonb_set(config, '{files,ocr_enabled}', 'false'::jsonb, true)
WHERE config->'files' IS NOT NULL;
