-- Add custom_instruction column for user-editable instructions
ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS custom_instruction TEXT DEFAULT '';

-- Migrate existing system_prompt content to custom_instruction
-- This preserves user's custom prompts in the new field
UPDATE chatbots 
SET custom_instruction = system_prompt 
WHERE system_prompt IS NOT NULL 
  AND system_prompt != ''
  AND custom_instruction = '';

-- Reset system_prompt to empty (will be generated dynamically from langconfig)
UPDATE chatbots SET system_prompt = '';

COMMENT ON COLUMN chatbots.custom_instruction IS 'User-editable custom instructions appended to base system prompt';
