BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Canonical languages
CREATE TABLE languages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    rtl BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

-- Plans (language-agnostic)
CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    billing_cycle TEXT NOT NULL DEFAULT 'monthly',
    price NUMERIC(10,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'TRY',
    trial_days INTEGER NOT NULL DEFAULT 0,
    config JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

-- Localized plan strings
CREATE TABLE plan_translations (
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    language_id UUID NOT NULL REFERENCES languages(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    UNIQUE (plan_id, language_id)
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url TEXT,
    is_email_verified BOOLEAN DEFAULT false,
    payment_customer_id VARCHAR(255),
    kvkk_accepted BOOLEAN DEFAULT false,
    kvkk_accepted_at TIMESTAMP,
    plan_id UUID NOT NULL REFERENCES plans(id),
    preferred_language_id UUID REFERENCES languages(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_users_email ON users(email);

-- Chatbots (include appearance/security and language_id)
CREATE TABLE chatbots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    system_prompt TEXT DEFAULT 'Sen yararlı, kibar ve bilgili bir yapay zeka asistanısın.',
    language_id UUID REFERENCES languages(id),
    model VARCHAR(100) DEFAULT 'gpt-3.5-turbo',
    temperature FLOAT DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 512,
    theme_color VARCHAR(50) DEFAULT 'rgba(255, 174, 0, 1)',
    welcome_message TEXT DEFAULT 'Merhaba! Size nasıl yardımcı olabilirim?',
    position VARCHAR(20) DEFAULT 'bottom-right',
    bot_message_color VARCHAR(50) DEFAULT 'rgba(252, 252, 253, 1)',
    user_message_color VARCHAR(50) DEFAULT 'rgba(250, 171, 0, 0.91)',
    bot_message_text_color VARCHAR(50) DEFAULT 'rgba(0, 0, 0, 1)',
    user_message_text_color VARCHAR(50) DEFAULT 'rgba(255, 255, 255, 1)',
    chat_font_family VARCHAR(50) DEFAULT 'Inter, sans-serif',
    chat_header_color VARCHAR(50) DEFAULT 'rgba(242, 167, 36, 1)',
    chat_header_text_color VARCHAR(50) DEFAULT 'rgba(247, 241, 241, 1)',
    chat_background_color VARCHAR(50) DEFAULT 'rgba(255, 245, 230, 1)',
    bot_icon VARCHAR(1024),
    bot_display_name VARCHAR(100),
    allowed_domains TEXT,
    embed_secret VARCHAR(255),
    secure_embed_enabled BOOLEAN DEFAULT false,
    suggested_questions JSONB,
    suggestions_enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_chatbots_user_id ON chatbots(user_id);

-- Data sources
CREATE TABLE data_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    source_type VARCHAR(50) NOT NULL,
    source_url VARCHAR(2048),
    file_path VARCHAR(1024),
    original_filename VARCHAR(255),
    text_content TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    error_message TEXT,
    processed_at TIMESTAMP,
    chunk_count INTEGER DEFAULT 0,
    capability_summary TEXT,
    suggested_questions JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_data_sources_chatbot_id ON data_sources(chatbot_id);
CREATE INDEX idx_data_sources_status ON data_sources(status);

-- Conversations
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    session_id VARCHAR(255),
    visitor_name VARCHAR(255),
    visitor_email VARCHAR(255),
    visitor_ip_hash VARCHAR(64),
    user_agent_hash VARCHAR(64),
    message_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conversations_chatbot_id ON conversations(chatbot_id);

-- Messages
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    tokens_used INTEGER,
    thumbs_up BOOLEAN,
    feedback_text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_conversation_id ON messages(conversation_id);

-- Analytics
CREATE TABLE analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    analytics_date DATE NOT NULL,
    total_conversations INTEGER DEFAULT 0,
    total_messages INTEGER DEFAULT 0,
    unanswered_messages INTEGER DEFAULT 0,
    thumbs_up_count INTEGER DEFAULT 0,
    thumbs_down_count INTEGER DEFAULT 0,
    average_tokens_per_message FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chatbot_id, analytics_date)
);

CREATE INDEX idx_analytics_chatbot_date ON analytics(chatbot_id, analytics_date);

-- Payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'TRY',
    status VARCHAR(50) DEFAULT 'pending',
    payment_method VARCHAR(50),
    iyzico_payment_id VARCHAR(255),
    iyzico_conversation_id VARCHAR(255),
    plan_type VARCHAR(50),
    billing_period_start DATE,
    billing_period_end DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(512) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE
);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Subscription history
CREATE TABLE user_subscription_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id),
    status TEXT NOT NULL CHECK (status IN ('pending','active','canceled','expired','suspended')),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ,
    auto_renews BOOLEAN DEFAULT true,
    payment_id UUID,
    provider_subscription_id TEXT,
    source TEXT,
    reason TEXT,
    request_id UUID UNIQUE
);
CREATE INDEX idx_user_sub_hist_user_starts ON user_subscription_history(user_id, starts_at DESC);
CREATE INDEX idx_user_sub_hist_user_status ON user_subscription_history(user_id, status);
CREATE INDEX idx_user_sub_hist_active_open ON user_subscription_history(user_id) WHERE status = 'active' AND ends_at IS NULL;

-- Seed canonical languages
INSERT INTO languages (code, name, rtl)
VALUES
    ('tr-TR', 'Turkish (Türkiye)', false),
    ('en-US', 'English (United States)', false)
ON CONFLICT (code) DO NOTHING;

-- Seed plans
INSERT INTO plans (code, status, billing_cycle, price, currency, trial_days, config)
VALUES
    ('free', 'active', 'lifetime', 0, 'TRY', 0, '{}'::jsonb),
    ('pro',  'active', 'monthly',  199, 'TRY', 7, '{}'::jsonb)
ON CONFLICT (code) DO NOTHING;

-- Seed translations for plans
INSERT INTO plan_translations (plan_id, language_id, name, description)
SELECT p.id, l.id,
    CASE l.code WHEN 'tr-TR' THEN
        CASE p.code WHEN 'free' THEN 'Ücretsiz' WHEN 'pro' THEN 'Pro' END
    ELSE
        CASE p.code WHEN 'free' THEN 'Free' WHEN 'pro' THEN 'Pro' END
    END,
    CASE l.code WHEN 'tr-TR' THEN
        CASE p.code WHEN 'free' THEN 'Temel özellikler ile sınırsız deneme.' WHEN 'pro' THEN 'Gelişmiş özellikler ve öncelikli destek.' END
    ELSE
        CASE p.code WHEN 'free' THEN 'Unlimited trials with basic features.' WHEN 'pro' THEN 'Advanced features and priority support.' END
    END
FROM plans p CROSS JOIN languages l
ON CONFLICT DO NOTHING;

COMMIT;
