-- Add chat_tokens column to usage_ingestions for atomic token quota enforcement
-- This column tracks the cumulative monthly chat tokens used per user
-- and allows atomic check-and-increment operations to prevent race conditions
ALTER TABLE usage_ingestions
ADD COLUMN IF NOT EXISTS chat_tokens INT NOT NULL DEFAULT 0;
