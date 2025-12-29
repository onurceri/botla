ALTER TABLE chatbots
ADD COLUMN bubble_radius VARCHAR(50) NOT NULL DEFAULT '22px',
ADD COLUMN input_background_color VARCHAR(50) NOT NULL DEFAULT 'rgba(255, 255, 255, 0.5)',
ADD COLUMN input_text_color VARCHAR(50) NOT NULL DEFAULT 'rgba(28, 28, 30, 1)',
ADD COLUMN send_button_color VARCHAR(50) NOT NULL DEFAULT 'rgba(246, 140, 0, 1)';
