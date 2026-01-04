-- Update existing Free plan with rate limit config
UPDATE plans SET config = jsonb_set(
  COALESCE(config, '{}'::jsonb),
  '{rate_limits}',
  '{
    "requests_per_minute": 100,
    "window_seconds": 60,
    "endpoints": {
      "chat": {
        "requests_per_minute": 30,
        "window_seconds": 60
      },
      "sources": {
        "requests_per_minute": 10,
        "window_seconds": 60
      }
    }
  }'::jsonb
)
WHERE code = 'free';

-- Update existing Pro plan with rate limit config
UPDATE plans SET config = jsonb_set(
  COALESCE(config, '{}'::jsonb),
  '{rate_limits}',
  '{
    "requests_per_minute": 500,
    "window_seconds": 60,
    "endpoints": {
      "chat": {
        "requests_per_minute": 100,
        "window_seconds": 60
      },
      "sources": {
        "requests_per_minute": 30,
        "window_seconds": 60
      }
    }
  }'::jsonb
)
WHERE code = 'pro';

-- Add Ultra plan if it doesn't exist
INSERT INTO plans (code, status, billing_cycle, price, currency, trial_days, config)
VALUES (
  'ultra',
  'active',
  'monthly',
  499,
  'TRY',
  7,
  '{
    "scraping": {
      "max_depth": 10,
      "max_pages_per_source": 1000,
      "max_concurrent_requests": 20,
      "timeout_seconds": 60
    },
    "files": {
      "max_file_size_mb": 100,
      "max_files_total": 1000,
      "allowed_types": ["pdf", "docx", "txt", "csv", "xlsx"]
    },
    "chat": {
      "max_messages_per_day": 100000,
      "max_tokens_per_request": 16000,
      "max_context_messages": 50,
      "suggestion_count": 10
    },
    "refresh": {
      "enabled": true,
      "min_interval_hours": 1
    },
    "security": {
      "ip_whitelist_enabled": true,
      "domain_whitelist_enabled": true
    },
    "guardrails": {
      "toxicity_enabled": true,
      "pii_detection_enabled": true,
      "topic_restriction_enabled": true
    },
    "branding": {
      "custom_branding_enabled": true,
      "remove_botla_branding": true
    },
    "rate_limits": {
      "requests_per_minute": 2000,
      "window_seconds": 60,
      "endpoints": {
        "chat": {
          "requests_per_minute": 500,
          "window_seconds": 60
        },
        "sources": {
          "requests_per_minute": 100,
          "window_seconds": 60
        }
      }
    },
    "max_chatbots": 50,
    "max_monthly_ingestions": 10000,
    "max_monthly_embedding_tokens": 100000000,
    "min_readd_cooldown_minutes": 0
  }'::jsonb
)
ON CONFLICT (code) DO UPDATE SET
  config = EXCLUDED.config;
