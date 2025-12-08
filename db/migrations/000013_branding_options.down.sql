-- Remove branding options from chatbots table
ALTER TABLE chatbots
DROP COLUMN IF EXISTS hide_branding,
DROP COLUMN IF EXISTS custom_branding;
