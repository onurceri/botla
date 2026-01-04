BEGIN;

-- =====================================================
-- STEP 1: Create the normalized plan_limits table
-- =====================================================
CREATE TABLE plan_limits (
    plan_id UUID PRIMARY KEY REFERENCES plans(id) ON DELETE CASCADE,
    
    -- Top-level limits
    max_chatbots INTEGER NOT NULL DEFAULT 1,
    max_monthly_ingestions INTEGER NOT NULL DEFAULT 50,
    max_monthly_embedding_tokens INTEGER NOT NULL DEFAULT 250000,
    min_readd_cooldown_minutes INTEGER NOT NULL DEFAULT 60,
    
    -- Scraping config
    scraping_dynamic_enabled BOOLEAN NOT NULL DEFAULT false,
    scraping_max_urls_per_bot INTEGER NOT NULL DEFAULT 1,
    scraping_max_pages_per_crawl INTEGER NOT NULL DEFAULT 5,
    
    -- Files config
    files_max_size_mb INTEGER NOT NULL DEFAULT 5,
    files_max_files_per_bot INTEGER NOT NULL DEFAULT 1,
    files_max_files_total INTEGER NOT NULL DEFAULT 5,
    files_total_storage_mb INTEGER NOT NULL DEFAULT 10,
    files_max_text_length INTEGER NOT NULL DEFAULT 400000,
    
    -- Chat config
    chat_default_model TEXT DEFAULT 'openai/gpt-4o-mini',
    chat_allowed_models TEXT[] NOT NULL DEFAULT ARRAY['openai/gpt-4o-mini'],
    chat_max_monthly_tokens INTEGER NOT NULL DEFAULT 100000,
    chat_rag_top_k INTEGER NOT NULL DEFAULT 3,
    chat_rag_max_context_tokens INTEGER NOT NULL DEFAULT 2000,
    chat_max_suggested_questions INTEGER NOT NULL DEFAULT 3,
    chat_max_manual_questions INTEGER NOT NULL DEFAULT 3,
    chat_min_response_token_limit INTEGER NOT NULL DEFAULT 1,
    chat_max_response_token_limit INTEGER NOT NULL DEFAULT 4096,
    
    -- Refresh config
    refresh_enabled BOOLEAN NOT NULL DEFAULT false,
    refresh_max_monthly INTEGER NOT NULL DEFAULT 0,
    
    -- Security config
    security_secure_embed_enabled BOOLEAN NOT NULL DEFAULT false,
    
    -- Guardrails config
    guardrails_can_customize_thresholds BOOLEAN NOT NULL DEFAULT false,
    guardrails_can_use_smart_fallback BOOLEAN NOT NULL DEFAULT true,
    guardrails_can_use_escalate_fallback BOOLEAN NOT NULL DEFAULT false,
    guardrails_can_manage_topics BOOLEAN NOT NULL DEFAULT false,
    guardrails_can_customize_messages BOOLEAN NOT NULL DEFAULT false,
    
    -- Branding config
    branding_can_hide_branding BOOLEAN NOT NULL DEFAULT false,
    branding_can_custom_branding BOOLEAN NOT NULL DEFAULT false,
    
    -- Rate limits config
    rate_limits_requests_per_minute INTEGER NOT NULL DEFAULT 100,
    rate_limits_window_seconds INTEGER NOT NULL DEFAULT 60,
    rate_limits_chat_rpm INTEGER NOT NULL DEFAULT 30,
    rate_limits_chat_window INTEGER NOT NULL DEFAULT 60,
    rate_limits_sources_rpm INTEGER NOT NULL DEFAULT 10,
    rate_limits_sources_window INTEGER NOT NULL DEFAULT 60,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

-- =====================================================
-- STEP 2: Add CHECK constraints for validation
-- =====================================================
ALTER TABLE plan_limits ADD CONSTRAINT chk_max_chatbots CHECK (max_chatbots >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_max_monthly_ingestions CHECK (max_monthly_ingestions >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_max_monthly_embedding_tokens CHECK (max_monthly_embedding_tokens >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_min_readd_cooldown_minutes CHECK (min_readd_cooldown_minutes >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_scraping_max_urls_per_bot CHECK (scraping_max_urls_per_bot >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_scraping_max_pages_per_crawl CHECK (scraping_max_pages_per_crawl >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_files_max_size_mb CHECK (files_max_size_mb > 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_files_max_files_per_bot CHECK (files_max_files_per_bot >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_files_max_files_total CHECK (files_max_files_total >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_files_total_storage_mb CHECK (files_total_storage_mb > 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_files_max_text_length CHECK (files_max_text_length >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_chat_max_monthly_tokens CHECK (chat_max_monthly_tokens >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_chat_rag_top_k CHECK (chat_rag_top_k >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_chat_rag_max_context_tokens CHECK (chat_rag_max_context_tokens >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_chat_max_suggested_questions CHECK (chat_max_suggested_questions >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_chat_max_manual_questions CHECK (chat_max_manual_questions >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_chat_min_response_token_limit CHECK (chat_min_response_token_limit >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_chat_max_response_token_limit CHECK (chat_max_response_token_limit >= chat_min_response_token_limit);
ALTER TABLE plan_limits ADD CONSTRAINT chk_refresh_max_monthly CHECK (refresh_max_monthly >= 0);
ALTER TABLE plan_limits ADD CONSTRAINT chk_rate_limits_requests_per_minute CHECK (rate_limits_requests_per_minute >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_rate_limits_window_seconds CHECK (rate_limits_window_seconds >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_rate_limits_chat_rpm CHECK (rate_limits_chat_rpm >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_rate_limits_chat_window CHECK (rate_limits_chat_window >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_rate_limits_sources_rpm CHECK (rate_limits_sources_rpm >= 1);
ALTER TABLE plan_limits ADD CONSTRAINT chk_rate_limits_sources_window CHECK (rate_limits_sources_window >= 1);

-- =====================================================
-- STEP 3: Migrate existing data from plans.config JSONB
-- =====================================================
INSERT INTO plan_limits (
    plan_id,
    max_chatbots,
    max_monthly_ingestions,
    max_monthly_embedding_tokens,
    min_readd_cooldown_minutes,
    scraping_dynamic_enabled,
    scraping_max_urls_per_bot,
    scraping_max_pages_per_crawl,
    files_max_size_mb,
    files_max_files_per_bot,
    files_max_files_total,
    files_total_storage_mb,
    files_max_text_length,
    chat_default_model,
    chat_allowed_models,
    chat_max_monthly_tokens,
    chat_rag_top_k,
    chat_rag_max_context_tokens,
    chat_max_suggested_questions,
    chat_max_manual_questions,
    chat_min_response_token_limit,
    chat_max_response_token_limit,
    refresh_enabled,
    refresh_max_monthly,
    security_secure_embed_enabled,
    guardrails_can_customize_thresholds,
    guardrails_can_use_smart_fallback,
    guardrails_can_use_escalate_fallback,
    guardrails_can_manage_topics,
    guardrails_can_customize_messages,
    branding_can_hide_branding,
    branding_can_custom_branding,
    rate_limits_requests_per_minute,
    rate_limits_window_seconds,
    rate_limits_chat_rpm,
    rate_limits_chat_window,
    rate_limits_sources_rpm,
    rate_limits_sources_window
)
SELECT 
    p.id,
    COALESCE((p.config->>'max_chatbots')::int, 1),
    COALESCE((p.config->>'max_monthly_ingestions')::int, 50),
    COALESCE((p.config->>'max_monthly_embedding_tokens')::int, 250000),
    COALESCE((p.config->>'min_readd_cooldown_minutes')::int, 60),
    COALESCE((p.config->'scraping'->>'dynamic_enabled')::boolean, false),
    COALESCE((p.config->'scraping'->>'max_urls_per_bot')::int, 1),
    COALESCE((p.config->'scraping'->>'max_pages_per_crawl')::int, 5),
    COALESCE((p.config->'files'->>'max_size_mb')::int, 5),
    COALESCE((p.config->'files'->>'max_files_per_bot')::int, 1),
    COALESCE((p.config->'files'->>'max_files_total')::int, 5),
    COALESCE((p.config->'files'->>'total_storage_mb')::int, 10),
    COALESCE((p.config->'files'->>'max_text_length')::int, 400000),
    COALESCE(p.config->'chat'->>'default_model', 'openai/gpt-4o-mini'),
    CASE 
        WHEN p.config->'chat'->'allowed_models' IS NOT NULL 
             AND jsonb_typeof(p.config->'chat'->'allowed_models') = 'array'
        THEN ARRAY(SELECT jsonb_array_elements_text(p.config->'chat'->'allowed_models'))
        ELSE ARRAY['openai/gpt-4o-mini']
    END,
    COALESCE((p.config->'chat'->>'max_monthly_tokens')::int, 100000),
    COALESCE((p.config->'chat'->'rag'->>'top_k')::int, 3),
    COALESCE((p.config->'chat'->'rag'->>'max_context_tokens')::int, 2000),
    COALESCE((p.config->'chat'->>'max_suggested_questions')::int, 3),
    COALESCE((p.config->'chat'->>'max_manual_questions')::int, 3),
    COALESCE((p.config->'chat'->>'min_response_token_limit')::int, 1),
    COALESCE((p.config->'chat'->>'max_response_token_limit')::int, 4096),
    COALESCE((p.config->'refresh'->>'enabled')::boolean, false),
    COALESCE((p.config->'refresh'->>'max_monthly')::int, 0),
    COALESCE((p.config->'security'->>'secure_embed_enabled')::boolean, false),
    COALESCE((p.config->'guardrails'->>'can_customize_thresholds')::boolean, false),
    COALESCE((p.config->'guardrails'->>'can_use_smart_fallback')::boolean, true),
    COALESCE((p.config->'guardrails'->>'can_use_escalate_fallback')::boolean, false),
    COALESCE((p.config->'guardrails'->>'can_manage_topics')::boolean, false),
    COALESCE((p.config->'guardrails'->>'can_customize_messages')::boolean, false),
    COALESCE((p.config->'branding'->>'can_hide_branding')::boolean, false),
    COALESCE((p.config->'branding'->>'can_custom_branding')::boolean, false),
    COALESCE((p.config->'rate_limits'->>'requests_per_minute')::int, 100),
    COALESCE((p.config->'rate_limits'->>'window_seconds')::int, 60),
    COALESCE((p.config->'rate_limits'->'endpoints'->'chat'->>'requests_per_minute')::int, 30),
    COALESCE((p.config->'rate_limits'->'endpoints'->'chat'->>'window_seconds')::int, 60),
    COALESCE((p.config->'rate_limits'->'endpoints'->'sources'->>'requests_per_minute')::int, 10),
    COALESCE((p.config->'rate_limits'->'endpoints'->'sources'->>'window_seconds')::int, 60)
FROM plans p
WHERE p.deleted_at IS NULL;

-- =====================================================
-- STEP 4: Drop the JSONB column
-- =====================================================
ALTER TABLE plans DROP COLUMN config;

COMMIT;
