ALTER TABLE privacy_requests ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_privacy_requests_deleted_at ON privacy_requests(deleted_at) WHERE deleted_at IS NULL;
