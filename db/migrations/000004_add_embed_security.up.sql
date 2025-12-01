ALTER TABLE chatbots
ADD COLUMN allowed_domains TEXT,
ADD COLUMN embed_secret VARCHAR(255);
