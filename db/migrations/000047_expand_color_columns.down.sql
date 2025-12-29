-- Revert color columns back to VARCHAR(7)
-- Note: This may truncate data if RGBA values were stored

ALTER TABLE chatbots
  ALTER COLUMN theme_color TYPE VARCHAR(7) USING SUBSTRING(theme_color FROM 1 FOR 7),
  ALTER COLUMN bot_message_color TYPE VARCHAR(7) USING SUBSTRING(bot_message_color FROM 1 FOR 7),
  ALTER COLUMN user_message_color TYPE VARCHAR(7) USING SUBSTRING(user_message_color FROM 1 FOR 7),
  ALTER COLUMN bot_message_text_color TYPE VARCHAR(7) USING SUBSTRING(bot_message_text_color FROM 1 FOR 7),
  ALTER COLUMN user_message_text_color TYPE VARCHAR(7) USING SUBSTRING(user_message_text_color FROM 1 FOR 7),
  ALTER COLUMN chat_header_color TYPE VARCHAR(7) USING SUBSTRING(chat_header_color FROM 1 FOR 7),
  ALTER COLUMN chat_header_text_color TYPE VARCHAR(7) USING SUBSTRING(chat_header_text_color FROM 1 FOR 7),
  ALTER COLUMN chat_background_color TYPE VARCHAR(7) USING SUBSTRING(chat_background_color FROM 1 FOR 7);
