-- Remove ocr_enabled from all plan configs
-- OCR functionality has been removed from the platform

UPDATE plans 
SET config = config #- '{files,ocr_enabled}'
WHERE config->'files'->>'ocr_enabled' IS NOT NULL;
