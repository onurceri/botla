-- Privacy/KVKK data requests
CREATE TABLE privacy_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    user_email TEXT NOT NULL,  -- Store email in case user is deleted
    request_type TEXT NOT NULL CHECK (request_type IN ('export', 'deletion', 'correction')),
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'denied')),
    reason TEXT,  -- User's reason for request
    denial_reason TEXT,  -- Admin's reason for denial
    processed_by UUID REFERENCES users(id),
    processed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    export_url TEXT,
    export_expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_privacy_requests_user ON privacy_requests(user_id);
CREATE INDEX idx_privacy_requests_status ON privacy_requests(status);
CREATE INDEX idx_privacy_requests_created ON privacy_requests(created_at DESC);

-- Consent tracking
CREATE TABLE user_consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    consent_type TEXT NOT NULL CHECK (consent_type IN ('marketing', 'analytics', 'personalization', 'third_party')),
    granted BOOLEAN DEFAULT TRUE,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    ip_address INET,
    user_agent TEXT
);

CREATE UNIQUE INDEX idx_user_consents_unique ON user_consents(user_id, consent_type);

-- Data exports
CREATE TABLE data_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    requested_by UUID REFERENCES users(id),  -- Admin or user themselves
    format TEXT NOT NULL CHECK (format IN ('json', 'csv')),
    status TEXT DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    download_url TEXT,
    file_size_bytes BIGINT,
    expires_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_data_exports_user ON data_exports(user_id);
CREATE INDEX idx_data_exports_status ON data_exports(status);
