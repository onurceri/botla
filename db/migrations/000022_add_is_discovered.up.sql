-- Add is_discovered column to track sources discovered via crawling
-- Discovered sources should not perform sub-page discovery themselves
ALTER TABLE data_sources ADD COLUMN IF NOT EXISTS is_discovered BOOLEAN DEFAULT FALSE;

-- Mark existing sources that came from pending_discovered_urls as discovered
UPDATE data_sources ds
SET is_discovered = TRUE
WHERE EXISTS (
    SELECT 1 FROM pending_discovered_urls pdu 
    WHERE pdu.url = ds.source_url 
    AND pdu.chatbot_id = ds.chatbot_id
    AND pdu.status = 'selected'
);
