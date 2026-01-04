-- Restore system_prompt from custom_instruction
UPDATE chatbots 
SET system_prompt = custom_instruction 
WHERE custom_instruction IS NOT NULL AND custom_instruction != '';

-- Drop custom_instruction column
ALTER TABLE chatbots DROP COLUMN IF EXISTS custom_instruction;
