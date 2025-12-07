-- Extend plans.config JSON with new ingestion-related defaults
-- Defaults chosen conservatively; update per plan codes via application logic if needed
UPDATE plans SET config = jsonb_set(config::jsonb, '{max_monthly_ingestions}', to_jsonb(50), true);
UPDATE plans SET config = jsonb_set(config::jsonb, '{max_monthly_embedding_tokens}', to_jsonb(250000), true);
UPDATE plans SET config = jsonb_set(config::jsonb, '{min_readd_cooldown_minutes}', to_jsonb(60), true);

