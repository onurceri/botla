-- Expand color columns to support RGBA format (e.g., rgba(255, 255, 255, 0.5))
-- Original columns were VARCHAR(7) which only supports HEX (#RRGGBB)

ALTER TABLE chatbots
  ALTER COLUMN theme_color TYPE VARCHAR(50),
  ALTER COLUMN bot_message_color TYPE VARCHAR(50),
  ALTER COLUMN user_message_color TYPE VARCHAR(50),
  ALTER COLUMN bot_message_text_color TYPE VARCHAR(50),
  ALTER COLUMN user_message_text_color TYPE VARCHAR(50),
  ALTER COLUMN chat_header_color TYPE VARCHAR(50),
  ALTER COLUMN chat_header_text_color TYPE VARCHAR(50),
  ALTER COLUMN chat_background_color TYPE VARCHAR(50);
