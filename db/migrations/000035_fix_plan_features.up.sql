-- Comprehensive Plan Configuration Fix
-- Aligns database config with Go models and landing page features
BEGIN;

-- =====================================================
-- FREE PLAN: Basic features only
-- =====================================================
UPDATE plans SET config = jsonb_build_object(
    'scraping', jsonb_build_object(
        'dynamic_enabled', false,
        'max_urls_per_bot', 1,
        'max_pages_per_crawl', 5
    ),
    'files', jsonb_build_object(
        'ocr_enabled', false,
        'max_size_mb', 5,
        'max_files_per_bot', 1,
        'max_files_total', 5,
        'total_storage_mb', 10,
        'max_text_length', 400000
    ),
    'chat', jsonb_build_object(
        'default_model', 'openai/gpt-4o-mini',
        'allowed_models', '["openai/gpt-4o-mini"]'::jsonb,
        'max_monthly_tokens', 100000,
        'rag', jsonb_build_object(
            'top_k', 3,
            'max_context_tokens', 2000
        ),
        'max_suggested_questions', 3
    ),
    'refresh', jsonb_build_object(
        'enabled', false,
        'max_monthly', 0
    ),
    'security', jsonb_build_object(
        'secure_embed_enabled', false
    ),
    'guardrails', jsonb_build_object(
        'can_customize_thresholds', false,
        'can_use_smart_fallback', false,
        'can_use_escalate_fallback', false,
        'can_manage_topics', false,
        'can_customize_messages', false
    ),
    'branding', jsonb_build_object(
        'can_hide_branding', false,
        'can_custom_branding', false
    ),
    'rate_limits', jsonb_build_object(
        'requests_per_minute', 100,
        'window_seconds', 60,
        'endpoints', jsonb_build_object(
            'chat', jsonb_build_object(
                'requests_per_minute', 30,
                'window_seconds', 60
            ),
            'sources', jsonb_build_object(
                'requests_per_minute', 10,
                'window_seconds', 60
            )
        )
    ),
    'max_chatbots', 1,
    'max_monthly_ingestions', 50,
    'max_monthly_embedding_tokens', 250000,
    'min_readd_cooldown_minutes', 60
)
WHERE code = 'free';

-- =====================================================
-- PRO PLAN: Advanced features
-- Based on landing page: 10 chatbots, 1M tokens, 10 URLs, 20 PDFs
-- =====================================================
UPDATE plans SET config = jsonb_build_object(
    'scraping', jsonb_build_object(
        'dynamic_enabled', true,
        'max_urls_per_bot', 10,
        'max_pages_per_crawl', 50
    ),
    'files', jsonb_build_object(
        'ocr_enabled', true,
        'max_size_mb', 20,
        'max_files_per_bot', 20,
        'max_files_total', 100,
        'total_storage_mb', 500,
        'max_text_length', 400000
    ),
    'chat', jsonb_build_object(
        'default_model', 'openai/gpt-4o',
        'allowed_models', '["openai/gpt-4o-mini", "openai/gpt-4o"]'::jsonb,
        'max_monthly_tokens', 1000000,
        'rag', jsonb_build_object(
            'top_k', 5,
            'max_context_tokens', 4000
        ),
        'max_suggested_questions', 6
    ),
    'refresh', jsonb_build_object(
        'enabled', true,
        'max_monthly', 5
    ),
    'security', jsonb_build_object(
        'secure_embed_enabled', true
    ),
    'guardrails', jsonb_build_object(
        'can_customize_thresholds', true,
        'can_use_smart_fallback', true,
        'can_use_escalate_fallback', false,
        'can_manage_topics', true,
        'can_customize_messages', true
    ),
    'branding', jsonb_build_object(
        'can_hide_branding', true,
        'can_custom_branding', false
    ),
    'rate_limits', jsonb_build_object(
        'requests_per_minute', 500,
        'window_seconds', 60,
        'endpoints', jsonb_build_object(
            'chat', jsonb_build_object(
                'requests_per_minute', 100,
                'window_seconds', 60
            ),
            'sources', jsonb_build_object(
                'requests_per_minute', 30,
                'window_seconds', 60
            )
        )
    ),
    'max_chatbots', 10,
    'max_monthly_ingestions', 500,
    'max_monthly_embedding_tokens', 2500000,
    'min_readd_cooldown_minutes', 30
)
WHERE code = 'pro';

-- =====================================================
-- ULTRA PLAN: All features unlocked
-- Based on landing page: 100 chatbots, 5M tokens, 50 URLs, 100 PDFs
-- Features: Whitelabel, Human Handoff, Custom Integrations
-- =====================================================
UPDATE plans SET config = jsonb_build_object(
    'scraping', jsonb_build_object(
        'dynamic_enabled', true,
        'max_urls_per_bot', 50,
        'max_pages_per_crawl', 200
    ),
    'files', jsonb_build_object(
        'ocr_enabled', true,
        'max_size_mb', 50,
        'max_files_per_bot', 100,
        'max_files_total', 1000,
        'total_storage_mb', 2000,
        'max_text_length', 400000
    ),
    'chat', jsonb_build_object(
        'default_model', 'openai/gpt-4o',
        'allowed_models', '["openai/gpt-4o-mini", "openai/gpt-4o", "openai/gpt-5"]'::jsonb,
        'max_monthly_tokens', 5000000,
        'rag', jsonb_build_object(
            'top_k', 10,
            'max_context_tokens', 8000
        ),
        'max_suggested_questions', 10
    ),
    'refresh', jsonb_build_object(
        'enabled', true,
        'max_monthly', 100
    ),
    'security', jsonb_build_object(
        'secure_embed_enabled', true
    ),
    'guardrails', jsonb_build_object(
        'can_customize_thresholds', true,
        'can_use_smart_fallback', true,
        'can_use_escalate_fallback', true,
        'can_manage_topics', true,
        'can_customize_messages', true
    ),
    'branding', jsonb_build_object(
        'can_hide_branding', true,
        'can_custom_branding', true
    ),
    'rate_limits', jsonb_build_object(
        'requests_per_minute', 2000,
        'window_seconds', 60,
        'endpoints', jsonb_build_object(
            'chat', jsonb_build_object(
                'requests_per_minute', 500,
                'window_seconds', 60
            ),
            'sources', jsonb_build_object(
                'requests_per_minute', 100,
                'window_seconds', 60
            )
        )
    ),
    'max_chatbots', 100,
    'max_monthly_ingestions', 10000,
    'max_monthly_embedding_tokens', 100000000,
    'min_readd_cooldown_minutes', 0
)
WHERE code = 'ultra';

-- Update plan prices to match landing page
UPDATE plans SET price = 0, currency = 'TRY' WHERE code = 'free';
UPDATE plans SET price = 199, currency = 'TRY' WHERE code = 'pro';
UPDATE plans SET price = 999, currency = 'TRY' WHERE code = 'ultra';

COMMIT;
