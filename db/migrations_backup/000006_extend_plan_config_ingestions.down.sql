-- Remove ingestion-related keys from plans.config JSON
UPDATE plans SET config = (config::jsonb - 'max_monthly_ingestions' - 'max_monthly_embedding_tokens' - 'min_readd_cooldown_minutes')::json;

