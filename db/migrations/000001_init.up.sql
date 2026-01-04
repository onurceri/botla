-- Consolidated initial migration

CREATE EXTENSION IF NOT EXISTS btree_gist WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
CREATE OR REPLACE FUNCTION update_suggestion_jobs_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;
CREATE OR REPLACE FUNCTION update_training_jobs_updated_at() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;
CREATE TABLE action_execution_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    action_id uuid NOT NULL,
    conversation_id uuid,
    message_id uuid,
    status character varying(50) NOT NULL,
    request_payload jsonb,
    response_payload jsonb,
    error_message text,
    duration_ms integer,
    created_at timestamp with time zone DEFAULT now()
);
CREATE TABLE admin_audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    admin_user_id uuid,
    action text NOT NULL,
    target_type text NOT NULL,
    target_id uuid,
    details jsonb,
    ip_address inet,
    user_agent text,
    created_at timestamp with time zone DEFAULT now()
);
CREATE TABLE ai_models (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider character varying(50) NOT NULL,
    name character varying(100) NOT NULL,
    max_tokens integer DEFAULT 4096 NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    model_name character varying(100) NOT NULL,
    api_model_id character varying(150) NOT NULL
);
CREATE TABLE analytics (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    analytics_date date NOT NULL,
    total_conversations integer DEFAULT 0,
    total_messages integer DEFAULT 0,
    unanswered_messages integer DEFAULT 0,
    thumbs_up_count integer DEFAULT 0,
    thumbs_down_count integer DEFAULT 0,
    average_tokens_per_message double precision,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    total_tokens_used integer DEFAULT 0,
    handoff_count integer DEFAULT 0,
    avg_response_time_ms integer
);
CREATE TABLE chatbot_actions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    action_type text NOT NULL,
    config jsonb DEFAULT '{}'::jsonb NOT NULL,
    parameters jsonb DEFAULT '{}'::jsonb NOT NULL,
    enabled boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now(),
    tool_name text,
    version integer DEFAULT 1 NOT NULL
);
CREATE TABLE chatbots (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    system_prompt text DEFAULT 'Sen yararlı, kibar ve bilgili bir yapay zeka asistanısın.'::text,
    language_id uuid,
    model character varying(100) DEFAULT 'openai:gpt-4o-mini'::character varying,
    temperature double precision DEFAULT 0.7,
    max_tokens integer DEFAULT 512,
    theme_color character varying(50) DEFAULT 'rgba(255, 174, 0, 1)'::character varying,
    welcome_message text DEFAULT 'Merhaba! Size nasıl yardımcı olabilirim?'::text,
    "position" character varying(20) DEFAULT 'bottom-right'::character varying,
    bot_message_color character varying(50) DEFAULT 'rgba(252, 252, 253, 1)'::character varying,
    user_message_color character varying(50) DEFAULT 'rgba(250, 171, 0, 0.91)'::character varying,
    bot_message_text_color character varying(50) DEFAULT 'rgba(0, 0, 0, 1)'::character varying,
    user_message_text_color character varying(50) DEFAULT 'rgba(255, 255, 255, 1)'::character varying,
    chat_font_family character varying(50) DEFAULT 'Inter, sans-serif'::character varying,
    chat_header_color character varying(50) DEFAULT 'rgba(242, 167, 36, 1)'::character varying,
    chat_header_text_color character varying(50) DEFAULT 'rgba(247, 241, 241, 1)'::character varying,
    chat_background_color character varying(50) DEFAULT 'rgba(255, 245, 230, 1)'::character varying,
    bot_icon character varying(1024),
    bot_display_name character varying(100),
    allowed_domains text,
    embed_secret character varying(255),
    secure_embed_enabled boolean DEFAULT false,
    suggested_questions jsonb,
    suggestions_enabled boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    include_paths text[] DEFAULT '{}'::text[],
    exclude_paths text[] DEFAULT '{}'::text[],
    selector_whitelist text[] DEFAULT '{}'::text[],
    discovery_mode text DEFAULT 'auto'::text,
    refresh_policy text DEFAULT 'manual'::text,
    refresh_frequency text,
    next_refresh_at timestamp with time zone,
    last_refresh_at timestamp with time zone,
    hide_branding boolean DEFAULT false,
    custom_branding jsonb,
    confidence_threshold double precision DEFAULT 0.35,
    fallback_messages jsonb DEFAULT '{}'::jsonb,
    topic_restrictions jsonb DEFAULT '{}'::jsonb,
    handoff_enabled boolean DEFAULT false,
    handoff_type text DEFAULT 'email'::text,
    handoff_config jsonb DEFAULT '{}'::jsonb,
    workspace_id uuid,
    organization_id uuid,
    threshold_config jsonb DEFAULT '{"fallback_mode": "smart", "high_threshold": 0.50, "medium_threshold": 0.30, "show_confidence_warning": true}'::jsonb,
    custom_instruction text DEFAULT ''::text,
    manual_questions jsonb DEFAULT '[]'::jsonb,
    bubble_radius character varying(50) DEFAULT '22px'::character varying NOT NULL,
    input_background_color character varying(50) DEFAULT 'rgba(255, 255, 255, 0.5)'::character varying NOT NULL,
    input_text_color character varying(50) DEFAULT 'rgba(28, 28, 30, 1)'::character varying NOT NULL,
    send_button_color character varying(50) DEFAULT 'rgba(246, 140, 0, 1)'::character varying NOT NULL
);
CREATE TABLE conversations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    session_id character varying(255),
    visitor_name character varying(255),
    visitor_email character varying(255),
    visitor_ip_hash character varying(64),
    user_agent_hash character varying(64),
    message_count integer DEFAULT 0,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE data_exports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    requested_by uuid,
    format text NOT NULL,
    status text DEFAULT 'pending'::text,
    download_url text,
    file_size_bytes bigint,
    expires_at timestamp with time zone,
    error_message text,
    created_at timestamp with time zone DEFAULT now(),
    completed_at timestamp with time zone,
    CONSTRAINT data_exports_format_check CHECK ((format = ANY (ARRAY['json'::text, 'csv'::text]))),
    CONSTRAINT data_exports_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'processing'::text, 'completed'::text, 'failed'::text])))
);
CREATE TABLE data_sources (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    source_type character varying(50) NOT NULL,
    source_url character varying(2048),
    file_path character varying(1024),
    original_filename character varying(255),
    text_content text,
    status character varying(50) DEFAULT 'pending'::character varying,
    error_message text,
    processed_at timestamp without time zone,
    chunk_count integer DEFAULT 0,
    capability_summary text,
    suggested_questions jsonb,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    hash character varying(128),
    size_bytes bigint DEFAULT 0,
    last_refreshed_at timestamp with time zone,
    is_discovered boolean DEFAULT false
);
CREATE TABLE error_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    error_type text NOT NULL,
    message text NOT NULL,
    stack_trace text,
    request_path text,
    request_method text,
    user_id uuid,
    chatbot_id uuid,
    organization_id uuid,
    severity text DEFAULT 'error'::text,
    context jsonb,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT error_logs_severity_check CHECK ((severity = ANY (ARRAY['info'::text, 'warning'::text, 'error'::text, 'critical'::text])))
);
CREATE TABLE handoff_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    conversation_id uuid NOT NULL,
    status text DEFAULT 'pending'::text,
    assigned_to text,
    notes text,
    created_at timestamp without time zone DEFAULT now(),
    resolved_at timestamp without time zone,
    user_email text
);
CREATE TABLE languages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    name text NOT NULL,
    rtl boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE memberships (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text DEFAULT 'member'::text NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);
CREATE TABLE message_sources (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    message_id uuid NOT NULL,
    source_id uuid NOT NULL,
    chunk_index integer NOT NULL,
    relevance_score double precision NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);
CREATE TABLE messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    conversation_id uuid NOT NULL,
    role character varying(20) NOT NULL,
    content text NOT NULL,
    tokens_used integer,
    thumbs_up boolean,
    feedback_text text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    confidence_score double precision,
    sources_used uuid[] DEFAULT '{}'::uuid[],
    type character varying(20) DEFAULT 'normal'::character varying
);
CREATE TABLE organizations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    owner_id uuid NOT NULL,
    plan_id text DEFAULT 'agency_starter'::text,
    branding jsonb DEFAULT '{}'::jsonb,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);
CREATE TABLE payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    amount numeric(10,2) NOT NULL,
    currency character varying(3) DEFAULT 'TRY'::character varying,
    status character varying(50) DEFAULT 'pending'::character varying,
    payment_method character varying(50),
    iyzico_payment_id character varying(255),
    iyzico_conversation_id character varying(255),
    plan_type character varying(50),
    billing_period_start date,
    billing_period_end date,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE pending_discovered_urls (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    source_id uuid,
    url text NOT NULL,
    discovered_at timestamp without time zone DEFAULT now(),
    status text DEFAULT 'pending'::text
);
CREATE TABLE plan_limits (
    plan_id uuid NOT NULL,
    max_chatbots integer DEFAULT 1 NOT NULL,
    max_monthly_ingestions integer DEFAULT 50 NOT NULL,
    max_monthly_embedding_tokens integer DEFAULT 250000 NOT NULL,
    min_readd_cooldown_minutes integer DEFAULT 60 NOT NULL,
    scraping_dynamic_enabled boolean DEFAULT false NOT NULL,
    scraping_max_urls_per_bot integer DEFAULT 1 NOT NULL,
    scraping_max_pages_per_crawl integer DEFAULT 5 NOT NULL,
    files_max_size_mb integer DEFAULT 5 NOT NULL,
    files_max_files_per_bot integer DEFAULT 1 NOT NULL,
    files_max_files_total integer DEFAULT 5 NOT NULL,
    files_total_storage_mb integer DEFAULT 10 NOT NULL,
    files_max_text_length integer DEFAULT 400000 NOT NULL,
    chat_default_model text DEFAULT 'openai/gpt-4o-mini'::text,
    chat_allowed_models text[] DEFAULT ARRAY['openai/gpt-4o-mini'::text] NOT NULL,
    chat_max_monthly_tokens integer DEFAULT 100000 NOT NULL,
    chat_rag_top_k integer DEFAULT 3 NOT NULL,
    chat_rag_max_context_tokens integer DEFAULT 2000 NOT NULL,
    chat_max_suggested_questions integer DEFAULT 3 NOT NULL,
    chat_max_manual_questions integer DEFAULT 3 NOT NULL,
    chat_min_response_token_limit integer DEFAULT 1 NOT NULL,
    chat_max_response_token_limit integer DEFAULT 4096 NOT NULL,
    refresh_enabled boolean DEFAULT false NOT NULL,
    refresh_max_monthly integer DEFAULT 0 NOT NULL,
    security_secure_embed_enabled boolean DEFAULT false NOT NULL,
    guardrails_can_customize_thresholds boolean DEFAULT false NOT NULL,
    guardrails_can_use_smart_fallback boolean DEFAULT true NOT NULL,
    guardrails_can_use_escalate_fallback boolean DEFAULT false NOT NULL,
    guardrails_can_manage_topics boolean DEFAULT false NOT NULL,
    guardrails_can_customize_messages boolean DEFAULT false NOT NULL,
    branding_can_hide_branding boolean DEFAULT false NOT NULL,
    branding_can_custom_branding boolean DEFAULT false NOT NULL,
    rate_limits_requests_per_minute integer DEFAULT 100 NOT NULL,
    rate_limits_window_seconds integer DEFAULT 60 NOT NULL,
    rate_limits_chat_rpm integer DEFAULT 30 NOT NULL,
    rate_limits_chat_window integer DEFAULT 60 NOT NULL,
    rate_limits_sources_rpm integer DEFAULT 10 NOT NULL,
    rate_limits_sources_window integer DEFAULT 60 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone,
    CONSTRAINT chk_chat_max_manual_questions CHECK ((chat_max_manual_questions >= 0)),
    CONSTRAINT chk_chat_max_monthly_tokens CHECK ((chat_max_monthly_tokens >= 0)),
    CONSTRAINT chk_chat_max_response_token_limit CHECK ((chat_max_response_token_limit >= chat_min_response_token_limit)),
    CONSTRAINT chk_chat_max_suggested_questions CHECK ((chat_max_suggested_questions >= 0)),
    CONSTRAINT chk_chat_min_response_token_limit CHECK ((chat_min_response_token_limit >= 1)),
    CONSTRAINT chk_chat_rag_max_context_tokens CHECK ((chat_rag_max_context_tokens >= 1)),
    CONSTRAINT chk_chat_rag_top_k CHECK ((chat_rag_top_k >= 1)),
    CONSTRAINT chk_files_max_files_per_bot CHECK ((files_max_files_per_bot >= 0)),
    CONSTRAINT chk_files_max_files_total CHECK ((files_max_files_total >= 0)),
    CONSTRAINT chk_files_max_size_mb CHECK ((files_max_size_mb > 0)),
    CONSTRAINT chk_files_max_text_length CHECK ((files_max_text_length >= 0)),
    CONSTRAINT chk_files_total_storage_mb CHECK ((files_total_storage_mb > 0)),
    CONSTRAINT chk_max_chatbots CHECK ((max_chatbots >= 1)),
    CONSTRAINT chk_max_monthly_embedding_tokens CHECK ((max_monthly_embedding_tokens >= 0)),
    CONSTRAINT chk_max_monthly_ingestions CHECK ((max_monthly_ingestions >= 0)),
    CONSTRAINT chk_min_readd_cooldown_minutes CHECK ((min_readd_cooldown_minutes >= 0)),
    CONSTRAINT chk_rate_limits_chat_rpm CHECK ((rate_limits_chat_rpm >= 1)),
    CONSTRAINT chk_rate_limits_chat_window CHECK ((rate_limits_chat_window >= 1)),
    CONSTRAINT chk_rate_limits_requests_per_minute CHECK ((rate_limits_requests_per_minute >= 1)),
    CONSTRAINT chk_rate_limits_sources_rpm CHECK ((rate_limits_sources_rpm >= 1)),
    CONSTRAINT chk_rate_limits_sources_window CHECK ((rate_limits_sources_window >= 1)),
    CONSTRAINT chk_rate_limits_window_seconds CHECK ((rate_limits_window_seconds >= 1)),
    CONSTRAINT chk_refresh_max_monthly CHECK ((refresh_max_monthly >= 0)),
    CONSTRAINT chk_scraping_max_pages_per_crawl CHECK ((scraping_max_pages_per_crawl >= 0)),
    CONSTRAINT chk_scraping_max_urls_per_bot CHECK ((scraping_max_urls_per_bot >= 0))
);
CREATE TABLE plan_translations (
    plan_id uuid NOT NULL,
    language_id uuid NOT NULL,
    name text NOT NULL,
    description text
);
CREATE TABLE plans (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code text NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    billing_cycle text DEFAULT 'monthly'::text NOT NULL,
    price numeric(10,2) DEFAULT 0 NOT NULL,
    currency character varying(3) DEFAULT 'TRY'::character varying NOT NULL,
    trial_days integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE TABLE platform_metrics (
    id bigint NOT NULL,
    metric_name text NOT NULL,
    metric_value bigint NOT NULL,
    recorded_at timestamp with time zone DEFAULT now() NOT NULL,
    dimensions jsonb
);
CREATE SEQUENCE platform_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE platform_metrics_id_seq OWNED BY platform_metrics.id;
CREATE TABLE privacy_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    user_email text NOT NULL,
    request_type text NOT NULL,
    status text DEFAULT 'pending'::text,
    reason text,
    denial_reason text,
    processed_by uuid,
    processed_at timestamp with time zone,
    completed_at timestamp with time zone,
    export_url text,
    export_expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT privacy_requests_request_type_check CHECK ((request_type = ANY (ARRAY['export'::text, 'deletion'::text, 'correction'::text]))),
    CONSTRAINT privacy_requests_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'processing'::text, 'completed'::text, 'denied'::text])))
);
CREATE TABLE refresh_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash character varying(512) NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    revoked boolean DEFAULT false
);
CREATE TABLE suggestion_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    error_message text,
    suggested_questions text[],
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT valid_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'running'::character varying, 'completed'::character varying, 'failed'::character varying])::text[])))
);
CREATE TABLE training_jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    source_id uuid NOT NULL,
    chatbot_id uuid NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    current_step character varying(50),
    progress_percent integer DEFAULT 0,
    error_code character varying(100),
    error_message text,
    failed_step character varying(50),
    retry_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    CONSTRAINT valid_status CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'running'::character varying, 'completed'::character varying, 'failed'::character varying, 'cancelled'::character varying])::text[])))
);
CREATE TABLE unanswered_queries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    chatbot_id uuid NOT NULL,
    query text NOT NULL,
    occurrence_count integer DEFAULT 1,
    last_occurred_at timestamp without time zone DEFAULT now(),
    addressed boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT now()
);
CREATE TABLE usage_ingestions (
    user_id character varying(64) NOT NULL,
    period_month date NOT NULL,
    sources_count integer DEFAULT 0 NOT NULL,
    embedding_tokens integer DEFAULT 0 NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    refresh_count integer DEFAULT 0,
    auto_refresh_count integer DEFAULT 0,
    chat_tokens integer DEFAULT 0 NOT NULL
);
CREATE TABLE user_consents (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid,
    consent_type text NOT NULL,
    granted boolean DEFAULT true,
    granted_at timestamp with time zone DEFAULT now(),
    revoked_at timestamp with time zone,
    ip_address inet,
    user_agent text,
    CONSTRAINT user_consents_consent_type_check CHECK ((consent_type = ANY (ARRAY['marketing'::text, 'analytics'::text, 'personalization'::text, 'third_party'::text])))
);
CREATE TABLE user_subscription_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    plan_id uuid NOT NULL,
    status text NOT NULL,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone,
    auto_renews boolean DEFAULT true,
    payment_id uuid,
    provider_subscription_id text,
    source text,
    reason text,
    request_id uuid,
    CONSTRAINT user_subscription_history_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'active'::text, 'canceled'::text, 'expired'::text, 'suspended'::text])))
);
CREATE TABLE users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    full_name character varying(255),
    avatar_url text,
    is_email_verified boolean DEFAULT false,
    payment_customer_id character varying(255),
    kvkk_accepted boolean DEFAULT false,
    kvkk_accepted_at timestamp without time zone,
    plan_id uuid NOT NULL,
    preferred_language_id uuid,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    onboarding_completed boolean DEFAULT false,
    onboarding_step integer DEFAULT 0,
    onboarding_skipped boolean DEFAULT false,
    onboarding_data jsonb DEFAULT '{}'::jsonb,
    is_platform_admin boolean DEFAULT false
);
CREATE TABLE workspaces (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    organization_id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    client_name text,
    settings jsonb DEFAULT '{}'::jsonb,
    created_at timestamp without time zone DEFAULT now()
);
ALTER TABLE ONLY platform_metrics ALTER COLUMN id SET DEFAULT nextval('platform_metrics_id_seq'::regclass);
ALTER TABLE ONLY action_execution_logs
    ADD CONSTRAINT action_execution_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY admin_audit_logs
    ADD CONSTRAINT admin_audit_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY ai_models
    ADD CONSTRAINT ai_models_api_model_id_key UNIQUE (api_model_id);
ALTER TABLE ONLY ai_models
    ADD CONSTRAINT ai_models_pkey PRIMARY KEY (id);
ALTER TABLE ONLY analytics
    ADD CONSTRAINT analytics_chatbot_id_analytics_date_key UNIQUE (chatbot_id, analytics_date);
ALTER TABLE ONLY analytics
    ADD CONSTRAINT analytics_pkey PRIMARY KEY (id);
ALTER TABLE ONLY chatbot_actions
    ADD CONSTRAINT chatbot_actions_chatbot_id_name_key UNIQUE (chatbot_id, name);
ALTER TABLE ONLY chatbot_actions
    ADD CONSTRAINT chatbot_actions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY chatbots
    ADD CONSTRAINT chatbots_pkey PRIMARY KEY (id);
ALTER TABLE ONLY conversations
    ADD CONSTRAINT conversations_chatbot_id_session_id_key UNIQUE (chatbot_id, session_id);
ALTER TABLE ONLY conversations
    ADD CONSTRAINT conversations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY data_exports
    ADD CONSTRAINT data_exports_pkey PRIMARY KEY (id);
ALTER TABLE ONLY data_sources
    ADD CONSTRAINT data_sources_pkey PRIMARY KEY (id);
ALTER TABLE ONLY error_logs
    ADD CONSTRAINT error_logs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY handoff_requests
    ADD CONSTRAINT handoff_requests_pkey PRIMARY KEY (id);
ALTER TABLE ONLY languages
    ADD CONSTRAINT languages_code_key UNIQUE (code);
ALTER TABLE ONLY languages
    ADD CONSTRAINT languages_pkey PRIMARY KEY (id);
ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_organization_id_user_id_key UNIQUE (organization_id, user_id);
ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_pkey PRIMARY KEY (id);
ALTER TABLE ONLY message_sources
    ADD CONSTRAINT message_sources_message_id_source_id_chunk_index_key UNIQUE (message_id, source_id, chunk_index);
ALTER TABLE ONLY message_sources
    ADD CONSTRAINT message_sources_pkey PRIMARY KEY (id);
ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);
ALTER TABLE ONLY organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);
ALTER TABLE ONLY organizations
    ADD CONSTRAINT organizations_slug_key UNIQUE (slug);
ALTER TABLE ONLY payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY pending_discovered_urls
    ADD CONSTRAINT pending_discovered_urls_chatbot_id_url_key UNIQUE (chatbot_id, url);
ALTER TABLE ONLY pending_discovered_urls
    ADD CONSTRAINT pending_discovered_urls_pkey PRIMARY KEY (id);
ALTER TABLE ONLY plan_limits
    ADD CONSTRAINT plan_limits_pkey PRIMARY KEY (plan_id);
ALTER TABLE ONLY plan_translations
    ADD CONSTRAINT plan_translations_plan_id_language_id_key UNIQUE (plan_id, language_id);
ALTER TABLE ONLY plans
    ADD CONSTRAINT plans_code_key UNIQUE (code);
ALTER TABLE ONLY plans
    ADD CONSTRAINT plans_pkey PRIMARY KEY (id);
ALTER TABLE ONLY platform_metrics
    ADD CONSTRAINT platform_metrics_pkey PRIMARY KEY (id);
ALTER TABLE ONLY privacy_requests
    ADD CONSTRAINT privacy_requests_pkey PRIMARY KEY (id);
ALTER TABLE ONLY refresh_tokens
    ADD CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id);
ALTER TABLE ONLY suggestion_jobs
    ADD CONSTRAINT suggestion_jobs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY training_jobs
    ADD CONSTRAINT training_jobs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY unanswered_queries
    ADD CONSTRAINT unanswered_queries_chatbot_id_query_key UNIQUE (chatbot_id, query);
ALTER TABLE ONLY unanswered_queries
    ADD CONSTRAINT unanswered_queries_pkey PRIMARY KEY (id);
ALTER TABLE ONLY usage_ingestions
    ADD CONSTRAINT usage_ingestions_pkey PRIMARY KEY (user_id, period_month);
ALTER TABLE ONLY user_consents
    ADD CONSTRAINT user_consents_pkey PRIMARY KEY (id);
ALTER TABLE ONLY user_subscription_history
    ADD CONSTRAINT user_subscription_history_pkey PRIMARY KEY (id);
ALTER TABLE ONLY user_subscription_history
    ADD CONSTRAINT user_subscription_history_request_id_key UNIQUE (request_id);
ALTER TABLE ONLY users
    ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY workspaces
    ADD CONSTRAINT workspaces_organization_id_slug_key UNIQUE (organization_id, slug);
ALTER TABLE ONLY workspaces
    ADD CONSTRAINT workspaces_pkey PRIMARY KEY (id);
CREATE INDEX idx_action_logs_action_created ON action_execution_logs USING btree (action_id, created_at DESC);
CREATE INDEX idx_action_logs_chatbot_created ON action_execution_logs USING btree (chatbot_id, created_at DESC);
CREATE INDEX idx_admin_audit_logs_action ON admin_audit_logs USING btree (action);
CREATE INDEX idx_admin_audit_logs_admin ON admin_audit_logs USING btree (admin_user_id);
CREATE INDEX idx_admin_audit_logs_created ON admin_audit_logs USING btree (created_at DESC);
CREATE INDEX idx_admin_audit_logs_target ON admin_audit_logs USING btree (target_type, target_id);
CREATE INDEX idx_ai_models_model_name ON ai_models USING btree (model_name);
CREATE INDEX idx_analytics_chatbot_date ON analytics USING btree (chatbot_id, analytics_date);
CREATE INDEX idx_analytics_date_range ON analytics USING btree (chatbot_id, analytics_date DESC);
CREATE INDEX idx_chatbot_actions_bot ON chatbot_actions USING btree (chatbot_id) WHERE (enabled = true);
CREATE UNIQUE INDEX idx_chatbot_actions_tool_name_unique ON chatbot_actions USING btree (chatbot_id, tool_name) WHERE ((enabled = true) AND (tool_name IS NOT NULL));
CREATE INDEX idx_chatbots_next_refresh ON chatbots USING btree (next_refresh_at) WHERE ((refresh_policy = 'auto'::text) AND (deleted_at IS NULL));
CREATE INDEX idx_chatbots_user_id ON chatbots USING btree (user_id);
CREATE INDEX idx_chatbots_workspace ON chatbots USING btree (workspace_id);
CREATE INDEX idx_conversations_chatbot_id ON conversations USING btree (chatbot_id);
CREATE INDEX idx_data_exports_status ON data_exports USING btree (status);
CREATE INDEX idx_data_exports_user ON data_exports USING btree (user_id);
CREATE INDEX idx_data_sources_chatbot_id ON data_sources USING btree (chatbot_id);
CREATE INDEX idx_data_sources_status ON data_sources USING btree (status);
CREATE INDEX idx_error_logs_chatbot ON error_logs USING btree (chatbot_id) WHERE (chatbot_id IS NOT NULL);
CREATE INDEX idx_error_logs_created ON error_logs USING btree (created_at DESC);
CREATE INDEX idx_error_logs_severity ON error_logs USING btree (severity);
CREATE INDEX idx_error_logs_type ON error_logs USING btree (error_type);
CREATE INDEX idx_handoff_requests_chatbot ON handoff_requests USING btree (chatbot_id);
CREATE INDEX idx_handoff_requests_created ON handoff_requests USING btree (created_at DESC);
CREATE INDEX idx_handoff_requests_status ON handoff_requests USING btree (status);
CREATE INDEX idx_memberships_user ON memberships USING btree (user_id);
CREATE INDEX idx_message_sources_created_at ON message_sources USING btree (created_at);
CREATE INDEX idx_message_sources_message_id ON message_sources USING btree (message_id);
CREATE INDEX idx_message_sources_source_id ON message_sources USING btree (source_id);
CREATE INDEX idx_messages_conversation_id ON messages USING btree (conversation_id);
CREATE INDEX idx_payments_status ON payments USING btree (status);
CREATE INDEX idx_payments_user_id ON payments USING btree (user_id);
CREATE INDEX idx_pending_urls_chatbot ON pending_discovered_urls USING btree (chatbot_id, status);
CREATE INDEX idx_platform_metrics_name_time ON platform_metrics USING btree (metric_name, recorded_at DESC);
CREATE INDEX idx_privacy_requests_created ON privacy_requests USING btree (created_at DESC);
CREATE INDEX idx_privacy_requests_status ON privacy_requests USING btree (status);
CREATE INDEX idx_privacy_requests_user ON privacy_requests USING btree (user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens USING btree (token_hash);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens USING btree (token_hash);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens USING btree (user_id);
CREATE INDEX idx_suggestion_jobs_chatbot_id ON suggestion_jobs USING btree (chatbot_id);
CREATE INDEX idx_suggestion_jobs_created_at ON suggestion_jobs USING btree (created_at DESC);
CREATE INDEX idx_suggestion_jobs_status ON suggestion_jobs USING btree (status);
CREATE INDEX idx_training_jobs_chatbot_id ON training_jobs USING btree (chatbot_id);
CREATE INDEX idx_training_jobs_created_at ON training_jobs USING btree (created_at DESC);
CREATE INDEX idx_training_jobs_retry ON training_jobs USING btree (status, retry_count) WHERE (((status)::text = 'failed'::text) AND (retry_count < 3));
CREATE INDEX idx_training_jobs_source_id ON training_jobs USING btree (source_id);
CREATE INDEX idx_training_jobs_status ON training_jobs USING btree (status);
CREATE INDEX idx_unanswered_chatbot ON unanswered_queries USING btree (chatbot_id, addressed);
CREATE UNIQUE INDEX idx_user_consents_unique ON user_consents USING btree (user_id, consent_type);
CREATE INDEX idx_user_sub_hist_active_open ON user_subscription_history USING btree (user_id) WHERE ((status = 'active'::text) AND (ends_at IS NULL));
CREATE INDEX idx_user_sub_hist_user_starts ON user_subscription_history USING btree (user_id, starts_at DESC);
CREATE INDEX idx_user_sub_hist_user_status ON user_subscription_history USING btree (user_id, status);
CREATE INDEX idx_users_email ON users USING btree (email);
CREATE INDEX idx_users_is_platform_admin ON users USING btree (is_platform_admin) WHERE (is_platform_admin = true);
CREATE INDEX idx_users_onboarding ON users USING btree (onboarding_completed, onboarding_skipped) WHERE (deleted_at IS NULL);
CREATE INDEX idx_workspaces_org ON workspaces USING btree (organization_id);
CREATE UNIQUE INDEX ux_data_sources_chatbot_hash ON data_sources USING btree (chatbot_id, hash) WHERE ((hash IS NOT NULL) AND (deleted_at IS NULL));
CREATE TRIGGER suggestion_jobs_updated_at BEFORE UPDATE ON suggestion_jobs FOR EACH ROW EXECUTE FUNCTION update_suggestion_jobs_updated_at();
CREATE TRIGGER training_jobs_updated_at BEFORE UPDATE ON training_jobs FOR EACH ROW EXECUTE FUNCTION update_training_jobs_updated_at();
ALTER TABLE ONLY action_execution_logs
    ADD CONSTRAINT action_execution_logs_action_id_fkey FOREIGN KEY (action_id) REFERENCES chatbot_actions(id) ON DELETE CASCADE;
ALTER TABLE ONLY action_execution_logs
    ADD CONSTRAINT action_execution_logs_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY action_execution_logs
    ADD CONSTRAINT action_execution_logs_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE SET NULL;
ALTER TABLE ONLY action_execution_logs
    ADD CONSTRAINT action_execution_logs_message_id_fkey FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE SET NULL;
ALTER TABLE ONLY admin_audit_logs
    ADD CONSTRAINT admin_audit_logs_admin_user_id_fkey FOREIGN KEY (admin_user_id) REFERENCES users(id);
ALTER TABLE ONLY analytics
    ADD CONSTRAINT analytics_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY chatbot_actions
    ADD CONSTRAINT chatbot_actions_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY chatbots
    ADD CONSTRAINT chatbots_language_id_fkey FOREIGN KEY (language_id) REFERENCES languages(id);
ALTER TABLE ONLY chatbots
    ADD CONSTRAINT chatbots_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE ONLY chatbots
    ADD CONSTRAINT chatbots_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY chatbots
    ADD CONSTRAINT chatbots_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
ALTER TABLE ONLY conversations
    ADD CONSTRAINT conversations_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY data_exports
    ADD CONSTRAINT data_exports_requested_by_fkey FOREIGN KEY (requested_by) REFERENCES users(id);
ALTER TABLE ONLY data_exports
    ADD CONSTRAINT data_exports_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE ONLY data_sources
    ADD CONSTRAINT data_sources_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY handoff_requests
    ADD CONSTRAINT handoff_requests_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY handoff_requests
    ADD CONSTRAINT handoff_requests_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE;
ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE ONLY memberships
    ADD CONSTRAINT memberships_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY message_sources
    ADD CONSTRAINT message_sources_message_id_fkey FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE;
ALTER TABLE ONLY message_sources
    ADD CONSTRAINT message_sources_source_id_fkey FOREIGN KEY (source_id) REFERENCES data_sources(id) ON DELETE CASCADE;
ALTER TABLE ONLY messages
    ADD CONSTRAINT messages_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE;
ALTER TABLE ONLY organizations
    ADD CONSTRAINT organizations_owner_id_fkey FOREIGN KEY (owner_id) REFERENCES users(id);
ALTER TABLE ONLY payments
    ADD CONSTRAINT payments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY pending_discovered_urls
    ADD CONSTRAINT pending_discovered_urls_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY pending_discovered_urls
    ADD CONSTRAINT pending_discovered_urls_source_id_fkey FOREIGN KEY (source_id) REFERENCES data_sources(id) ON DELETE CASCADE;
ALTER TABLE ONLY plan_limits
    ADD CONSTRAINT plan_limits_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE;
ALTER TABLE ONLY plan_translations
    ADD CONSTRAINT plan_translations_language_id_fkey FOREIGN KEY (language_id) REFERENCES languages(id) ON DELETE CASCADE;
ALTER TABLE ONLY plan_translations
    ADD CONSTRAINT plan_translations_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE;
ALTER TABLE ONLY privacy_requests
    ADD CONSTRAINT privacy_requests_processed_by_fkey FOREIGN KEY (processed_by) REFERENCES users(id);
ALTER TABLE ONLY privacy_requests
    ADD CONSTRAINT privacy_requests_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE ONLY refresh_tokens
    ADD CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY suggestion_jobs
    ADD CONSTRAINT suggestion_jobs_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY training_jobs
    ADD CONSTRAINT training_jobs_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY training_jobs
    ADD CONSTRAINT training_jobs_source_id_fkey FOREIGN KEY (source_id) REFERENCES data_sources(id) ON DELETE CASCADE;
ALTER TABLE ONLY unanswered_queries
    ADD CONSTRAINT unanswered_queries_chatbot_id_fkey FOREIGN KEY (chatbot_id) REFERENCES chatbots(id) ON DELETE CASCADE;
ALTER TABLE ONLY user_consents
    ADD CONSTRAINT user_consents_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY user_subscription_history
    ADD CONSTRAINT user_subscription_history_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES plans(id);
ALTER TABLE ONLY user_subscription_history
    ADD CONSTRAINT user_subscription_history_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE ONLY users
    ADD CONSTRAINT users_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES plans(id);
ALTER TABLE ONLY users
    ADD CONSTRAINT users_preferred_language_id_fkey FOREIGN KEY (preferred_language_id) REFERENCES languages(id);
ALTER TABLE ONLY workspaces
    ADD CONSTRAINT workspaces_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE;

-- Seed data
INSERT INTO ai_models VALUES ('dbea1a08-e801-44d6-a08f-2f15ec6b2042', 'OpenRouter', 'GPT-4o Mini', 128000, true, '2025-12-28 23:16:55.235451+00', '2025-12-28 23:16:55.235451+00', 'gpt-4o-mini', 'openai/gpt-4o-mini');
INSERT INTO ai_models VALUES ('0dc3a903-a684-4ce4-85fe-dae7a4c81a32', 'OpenRouter', 'GPT-4o', 128000, true, '2025-12-28 23:16:55.235451+00', '2025-12-28 23:16:55.235451+00', 'gpt-4o', 'openai/gpt-4o');
INSERT INTO ai_models VALUES ('ae4f435b-0f2b-4c07-a7f7-d1ec3ca14d36', 'OpenRouter', 'GPT-5', 128000, true, '2025-12-28 23:16:55.235451+00', '2025-12-28 23:16:55.235451+00', 'gpt-5', 'openai/gpt-5');
INSERT INTO languages VALUES ('1228a1b3-ad14-4899-a435-57c1042b52ed', 'tr-TR', 'Turkish (Türkiye)', false, '2025-12-28 23:16:54.708516+00', NULL, NULL);
INSERT INTO languages VALUES ('bd1c8677-7f03-425a-b5aa-4baf47e77cf8', 'en-US', 'English (United States)', false, '2025-12-28 23:16:54.708516+00', NULL, NULL);
INSERT INTO plans VALUES ('1eb60c59-8c7c-4d35-ae9b-5ff43831b64b', 'ultra', 'active', 'monthly', 999.00, 'TRY', 0, '2025-12-28 23:16:54.819028+00', NULL, NULL);
INSERT INTO plans VALUES ('64127d68-aaf8-4cd2-bba8-fff82f5babf6', 'free', 'active', 'lifetime', 0.00, 'TRY', 0, '2025-12-28 23:16:54.708516+00', NULL, NULL);
INSERT INTO plans VALUES ('fd67ede1-692a-4d56-af69-0b7e062c7232', 'pro', 'active', 'monthly', 199.00, 'TRY', 7, '2025-12-28 23:16:54.708516+00', NULL, NULL);
INSERT INTO plan_limits VALUES ('1eb60c59-8c7c-4d35-ae9b-5ff43831b64b', 100, 10000, 100000000, 0, true, 50, 200, 50, 100, 1000, 2000, 400000, 'gpt-4o', '{gpt-4o-mini,gpt-4o,gpt-5}', 5000000, 10, 8000, 10, 10, 1, 4096, true, 100, true, true, true, true, true, true, true, true, 2000, 60, 500, 60, 100, 60, '2026-01-03 21:33:18.892709+00', NULL);
INSERT INTO plan_limits VALUES ('64127d68-aaf8-4cd2-bba8-fff82f5babf6', 1, 50, 250000, 60, false, 1, 5, 5, 1, 5, 10, 400000, 'gpt-4o-mini', '{gpt-4o-mini}', 100000, 3, 2000, 3, 3, 1, 4096, false, 0, false, false, true, false, false, false, false, false, 100, 60, 30, 60, 10, 60, '2026-01-03 21:33:18.892709+00', NULL);
INSERT INTO plan_limits VALUES ('fd67ede1-692a-4d56-af69-0b7e062c7232', 10, 500, 2500000, 30, true, 10, 50, 20, 20, 100, 500, 400000, 'gpt-4o', '{gpt-4o-mini,gpt-4o}', 1000000, 5, 4000, 6, 6, 1, 4096, true, 5, true, true, true, false, true, true, true, false, 500, 60, 100, 60, 30, 60, '2026-01-03 21:33:18.892709+00', NULL);
