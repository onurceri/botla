-- Remove openai: prefix
UPDATE chatbots 
SET model = REPLACE(model, 'openai:', '') 
WHERE model LIKE 'openai:%';

-- Revert default
ALTER TABLE chatbots ALTER COLUMN model SET DEFAULT 'gpt-4o-mini';
