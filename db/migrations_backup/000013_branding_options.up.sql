-- Add branding options to chatbots table
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS hide_branding BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS custom_branding JSONB DEFAULT NULL;

COMMENT ON COLUMN chatbots.hide_branding IS 'Hide Powered by Botla branding (requires Pro+ plan)';
COMMENT ON COLUMN chatbots.custom_branding IS 'Custom branding config: {logo_url, text, link} (requires Enterprise plan)';
