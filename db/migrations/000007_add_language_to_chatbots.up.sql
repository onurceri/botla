-- Add language column to chatbots table with default value 'tr'
ALTER TABLE chatbots ADD COLUMN language VARCHAR(10) DEFAULT 'tr' NOT NULL;
