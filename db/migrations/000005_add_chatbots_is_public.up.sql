-- Add is_public flag to chatbots for public endpoint information hiding.
-- Existing chatbots default to public (true) to preserve current behavior.
ALTER TABLE chatbots ADD COLUMN IF NOT EXISTS is_public BOOLEAN NOT NULL DEFAULT TRUE;
