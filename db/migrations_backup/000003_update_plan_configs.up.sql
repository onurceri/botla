BEGIN;

-- Insert Ultra plan if it doesn't exist
INSERT INTO plans (code, price, currency, trial_days, config)
VALUES ('ultra', 1000.00, 'TRY', 0, '{}')
ON CONFLICT (code) DO NOTHING;

-- Update Free Plan Config
UPDATE plans 
SET config = '{
  "scraping": {
    "dynamic_enabled": false,
    "max_urls_per_bot": 1,
    "max_pages_per_crawl": 0
  },
  "files": {
    "ocr_enabled": false,
    "max_size_mb": 5,
    "max_files_per_bot": 1,
    "total_storage_mb": 10
  },
  "chat": {
    "allowed_models": ["gpt-4o-mini"],
    "max_monthly_tokens": 100000,
    "rag": {
      "top_k": 3,
      "max_context_tokens": 2000
    }
  }
}'::jsonb
WHERE code = 'free';

-- Update Pro Plan Config
UPDATE plans 
SET config = '{
  "scraping": {
    "dynamic_enabled": true,
    "max_urls_per_bot": 10,
    "max_pages_per_crawl": 10
  },
  "files": {
    "ocr_enabled": true,
    "max_size_mb": 20,
    "max_files_per_bot": 20,
    "total_storage_mb": 500
  },
  "chat": {
    "allowed_models": ["gpt-4o-mini", "gpt-4o"],
    "max_monthly_tokens": 1000000,
    "rag": {
      "top_k": 5,
      "max_context_tokens": 4000
    }
  }
}'::jsonb
WHERE code = 'pro';

-- Update Ultra Plan Config
UPDATE plans 
SET config = '{
  "scraping": {
    "dynamic_enabled": true,
    "max_urls_per_bot": 50,
    "max_pages_per_crawl": 100
  },
  "files": {
    "ocr_enabled": true,
    "max_size_mb": 50,
    "max_files_per_bot": 100,
    "total_storage_mb": 2000
  },
  "chat": {
    "allowed_models": ["gpt-4o-mini", "gpt-4o", "claude-3-5-sonnet"],
    "max_monthly_tokens": 5000000,
    "rag": {
      "top_k": 10,
      "max_context_tokens": 8000
    }
  }
}'::jsonb
WHERE code = 'ultra';

COMMIT;
