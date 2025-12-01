ALTER TABLE chatbots
ADD COLUMN position VARCHAR(20) DEFAULT 'bottom-right',
ADD COLUMN bot_message_color VARCHAR(7) DEFAULT '#3b82f6',
ADD COLUMN user_message_color VARCHAR(7) DEFAULT '#3b82f6',
ADD COLUMN bot_message_text_color VARCHAR(7) DEFAULT '#ffffff',
ADD COLUMN user_message_text_color VARCHAR(7) DEFAULT '#ffffff',
ADD COLUMN chat_font_family VARCHAR(50) DEFAULT 'Inter, sans-serif',
ADD COLUMN chat_header_color VARCHAR(7) DEFAULT '#3b82f6',
ADD COLUMN chat_header_text_color VARCHAR(7) DEFAULT '#ffffff',
ADD COLUMN bot_icon VARCHAR(1024),
ADD COLUMN bot_display_name VARCHAR(100);
