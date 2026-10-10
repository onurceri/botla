-- Make GPT-6 Luna (gpt-6-luna, OpenRouter openai/gpt-6-luna) the default chat
-- model. gpt-4o-mini, gpt-4o and gpt-5 stay registered and allowed so existing
-- bots keep working; Luna is added to the seeded plans and becomes the Free
-- plan's default (Pro and Ultra keep gpt-4o as their default).

INSERT INTO ai_models (id, provider, name, max_tokens, is_active, model_name, api_model_id)
SELECT '5b1e6c2a-7f3d-4e8a-9c41-2d6f0b8a3e17', 'OpenRouter', 'GPT-6 Luna', 1050000, true, 'gpt-6-luna', 'openai/gpt-6-luna'
WHERE NOT EXISTS (SELECT 1 FROM ai_models WHERE model_name = 'gpt-6-luna');

-- Seeded plans: allow Luna (first, so it is the fallback when a bot's model is
-- not allowed). Plans that already list it are left alone.
UPDATE plan_limits
SET chat_allowed_models = array_prepend('gpt-6-luna'::text, chat_allowed_models)
WHERE plan_id IN (SELECT id FROM plans WHERE code IN ('free', 'pro', 'ultra'))
  AND NOT ('gpt-6-luna' = ANY (chat_allowed_models));

-- Free plan default, only if it is still the old default (keeps admin edits).
UPDATE plan_limits
SET chat_default_model = 'gpt-6-luna'
WHERE plan_id IN (SELECT id FROM plans WHERE code = 'free')
  AND chat_default_model = 'gpt-4o-mini';

-- Column defaults for new rows (bare model names, as the rows use).
ALTER TABLE plan_limits ALTER COLUMN chat_default_model SET DEFAULT 'gpt-6-luna';
ALTER TABLE plan_limits ALTER COLUMN chat_allowed_models SET DEFAULT ARRAY['gpt-6-luna'::text, 'gpt-4o-mini'::text];
ALTER TABLE chatbots ALTER COLUMN model SET DEFAULT 'gpt-6-luna';
