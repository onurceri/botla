ALTER TABLE chatbots
DROP COLUMN IF EXISTS allowed_domains,
DROP COLUMN IF EXISTS embed_secret;
