-- Reverse 000007. Bots already on gpt-6-luna are moved back to gpt-4o-mini
-- before the model row is removed, so none is left on an unknown model.
ALTER TABLE chatbots ALTER COLUMN model SET DEFAULT 'openai:gpt-4o-mini';
ALTER TABLE plan_limits ALTER COLUMN chat_allowed_models SET DEFAULT ARRAY['openai/gpt-4o-mini'::text];
ALTER TABLE plan_limits ALTER COLUMN chat_default_model SET DEFAULT 'openai/gpt-4o-mini';

UPDATE plan_limits SET chat_default_model = 'gpt-4o-mini' WHERE chat_default_model = 'gpt-6-luna';
UPDATE plan_limits SET chat_allowed_models = array_remove(chat_allowed_models, 'gpt-6-luna');
UPDATE chatbots SET model = 'gpt-4o-mini' WHERE model = 'gpt-6-luna';

DELETE FROM ai_models WHERE model_name = 'gpt-6-luna';
