BEGIN;

-- =====================================================
-- STEP 1: Recreate the config JSONB column
-- =====================================================
ALTER TABLE plans ADD COLUMN config JSONB DEFAULT '{}'::jsonb;

-- =====================================================
-- STEP 2: Migrate data back from plan_limits to JSONB
-- =====================================================
UPDATE plans p SET config = jsonb_build_object(
    'max_chatbots', pl.max_chatbots,
    'max_monthly_ingestions', pl.max_monthly_ingestions,
    'max_monthly_embedding_tokens', pl.max_monthly_embedding_tokens,
    'min_readd_cooldown_minutes', pl.min_readd_cooldown_minutes,
    'scraping', jsonb_build_object(
        'dynamic_enabled', pl.scraping_dynamic_enabled,
        'max_urls_per_bot', pl.scraping_max_urls_per_bot,
        'max_pages_per_crawl', pl.scraping_max_pages_per_crawl
    ),
    'files', jsonb_build_object(
        'max_size_mb', pl.files_max_size_mb,
        'max_files_per_bot', pl.files_max_files_per_bot,
        'max_files_total', pl.files_max_files_total,
        'total_storage_mb', pl.files_total_storage_mb,
        'max_text_length', pl.files_max_text_length
    ),
    'chat', jsonb_build_object(
        'default_model', pl.chat_default_model,
        'allowed_models', to_jsonb(pl.chat_allowed_models),
        'max_monthly_tokens', pl.chat_max_monthly_tokens,
        'rag', jsonb_build_object(
            'top_k', pl.chat_rag_top_k,
            'max_context_tokens', pl.chat_rag_max_context_tokens
        ),
        'max_suggested_questions', pl.chat_max_suggested_questions,
        'max_manual_questions', pl.chat_max_manual_questions,
        'min_response_token_limit', pl.chat_min_response_token_limit,
        'max_response_token_limit', pl.chat_max_response_token_limit
    ),
    'refresh', jsonb_build_object(
        'enabled', pl.refresh_enabled,
        'max_monthly', pl.refresh_max_monthly
    ),
    'security', jsonb_build_object(
        'secure_embed_enabled', pl.security_secure_embed_enabled
    ),
    'guardrails', jsonb_build_object(
        'can_customize_thresholds', pl.guardrails_can_customize_thresholds,
        'can_use_smart_fallback', pl.guardrails_can_use_smart_fallback,
        'can_use_escalate_fallback', pl.guardrails_can_use_escalate_fallback,
        'can_manage_topics', pl.guardrails_can_manage_topics,
        'can_customize_messages', pl.guardrails_can_customize_messages
    ),
    'branding', jsonb_build_object(
        'can_hide_branding', pl.branding_can_hide_branding,
        'can_custom_branding', pl.branding_can_custom_branding
    ),
    'rate_limits', jsonb_build_object(
        'requests_per_minute', pl.rate_limits_requests_per_minute,
        'window_seconds', pl.rate_limits_window_seconds,
        'endpoints', jsonb_build_object(
            'chat', jsonb_build_object(
                'requests_per_minute', pl.rate_limits_chat_rpm,
                'window_seconds', pl.rate_limits_chat_window
            ),
            'sources', jsonb_build_object(
                'requests_per_minute', pl.rate_limits_sources_rpm,
                'window_seconds', pl.rate_limits_sources_window
            )
        )
    )
)
FROM plan_limits pl
WHERE pl.plan_id = p.id;

-- =====================================================
-- STEP 3: Drop the plan_limits table
-- =====================================================
DROP TABLE IF EXISTS plan_limits;

COMMIT;
