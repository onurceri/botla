-- Revert changes (Best effort to restore previous state logic)
-- We can't easily "restore" deleted keys exactly as they were without a backup, 
-- but we can re-add the old structure and remove new keys.

-- Free Plan Revert
UPDATE plans 
SET config = jsonb_set(
    (config - 'guardrails' - 'max_chatbots'),
    '{chat,threshold_features}',
    '{"can_customize_thresholds": false, "can_use_smart_fallback": false, "can_use_escalate_fallback": false}'::jsonb
)
WHERE code = 'free';

-- Pro Plan Revert
UPDATE plans 
SET config = jsonb_set(
    (config - 'guardrails' - 'max_chatbots'),
    '{chat,threshold_features}',
    '{"can_customize_thresholds": true, "can_use_smart_fallback": true, "can_use_escalate_fallback": false}'::jsonb
)
WHERE code = 'pro';

-- Ultra Plan Revert
UPDATE plans 
SET config = jsonb_set(
    (config - 'guardrails' - 'max_chatbots'),
    '{chat,threshold_features}',
    '{"can_customize_thresholds": true, "can_use_smart_fallback": true, "can_use_escalate_fallback": true}'::jsonb
)
WHERE code = 'ultra';
