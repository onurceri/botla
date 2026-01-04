-- Update existing models to include provider prefix
UPDATE chatbots 
SET model = 'openai:' || model 
WHERE model NOT LIKE '%:%';

-- Set default to new format
ALTER TABLE chatbots ALTER COLUMN model SET DEFAULT 'openai:gpt-4o-mini';
