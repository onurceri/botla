-- Admin role flag on users table
ALTER TABLE users ADD COLUMN is_platform_admin BOOLEAN DEFAULT FALSE;
CREATE INDEX idx_users_is_platform_admin ON users(is_platform_admin) WHERE is_platform_admin = TRUE;

-- Admin audit log
CREATE TABLE admin_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user_id UUID REFERENCES users(id),
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_admin_audit_logs_admin ON admin_audit_logs(admin_user_id);
CREATE INDEX idx_admin_audit_logs_action ON admin_audit_logs(action);
CREATE INDEX idx_admin_audit_logs_created ON admin_audit_logs(created_at DESC);
CREATE INDEX idx_admin_audit_logs_target ON admin_audit_logs(target_type, target_id);

-- Error tracking table
CREATE TABLE error_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    error_type TEXT NOT NULL,
    message TEXT NOT NULL,
    stack_trace TEXT,
    request_path TEXT,
    request_method TEXT,
    user_id UUID,
    chatbot_id UUID,
    organization_id UUID,
    severity TEXT DEFAULT 'error' CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    context JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_error_logs_type ON error_logs(error_type);
CREATE INDEX idx_error_logs_created ON error_logs(created_at DESC);
CREATE INDEX idx_error_logs_severity ON error_logs(severity);
CREATE INDEX idx_error_logs_chatbot ON error_logs(chatbot_id) WHERE chatbot_id IS NOT NULL;

-- Platform metrics (time-series for trends)
CREATE TABLE platform_metrics (
    id BIGSERIAL PRIMARY KEY,
    metric_name TEXT NOT NULL,
    metric_value BIGINT NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dimensions JSONB
);

CREATE INDEX idx_platform_metrics_name_time ON platform_metrics(metric_name, recorded_at DESC);
