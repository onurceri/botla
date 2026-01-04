-- Create pending_discovered_urls table for URL discovery workflow
CREATE TABLE IF NOT EXISTS pending_discovered_urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    source_id UUID REFERENCES data_sources(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    discovered_at TIMESTAMP DEFAULT NOW(),
    status TEXT DEFAULT 'pending', -- pending, selected, rejected
    
    UNIQUE(chatbot_id, url)
);

-- Index for efficient lookups by chatbot and status
CREATE INDEX idx_pending_urls_chatbot ON pending_discovered_urls(chatbot_id, status);

-- Add discovery_mode column to chatbots table
ALTER TABLE chatbots 
ADD COLUMN IF NOT EXISTS discovery_mode TEXT DEFAULT 'auto';

-- Comment: discovery_mode values: 'auto' (default - immediate addition), 'pending' (require approval), 'disabled' (no crawl)
