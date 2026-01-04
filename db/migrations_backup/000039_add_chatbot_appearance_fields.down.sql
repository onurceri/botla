ALTER TABLE chatbots
DROP COLUMN IF EXISTS bubble_radius,
DROP COLUMN IF EXISTS input_background_color,
DROP COLUMN IF EXISTS input_text_color,
DROP COLUMN IF EXISTS send_button_color;
