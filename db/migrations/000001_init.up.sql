BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    avatar_url TEXT,
    is_email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    subscription_plan VARCHAR(50) DEFAULT 'free',
    subscription_started_at TIMESTAMP,
    subscription_expires_at TIMESTAMP,
    payment_customer_id VARCHAR(255),
    kvkk_accepted BOOLEAN DEFAULT false,
    kvkk_accepted_at TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_users_email ON users(email);

CREATE TABLE chatbots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    system_prompt TEXT DEFAULT 'Sen yararlı, kibar ve bilgili bir yapay zeka asistanısın.',
    model VARCHAR(100) DEFAULT 'gpt-3.5-turbo',
    temperature FLOAT DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 512,
    theme_color VARCHAR(7) DEFAULT '#3b82f6',
    welcome_message TEXT DEFAULT 'Merhaba! Size nasıl yardımcı olabilirim?',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_chatbots_user_id ON chatbots(user_id);

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
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_data_sources_chatbot_id ON data_sources(chatbot_id);
CREATE INDEX idx_data_sources_status ON data_sources(status);

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

COMMIT;
