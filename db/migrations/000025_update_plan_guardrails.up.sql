-- Free Plan Update
UPDATE plans 
SET config = jsonb_set(
    jsonb_set(
        (config - 'chat' || jsonb_build_object('chat', (config->'chat') - 'threshold_features')),
        '{max_chatbots}', 
        '1'::jsonb
    ),
    '{guardrails}',
    '{
        "can_manage_topics": false,
        "can_customize_messages": false,
        "can_customize_thresholds": false,
        "can_use_smart_fallback": false,
        "can_use_escalate_fallback": false
    }'::jsonb
)
WHERE code = 'free';

-- Pro Plan Update
UPDATE plans 
SET config = jsonb_set(
    jsonb_set(
        (config - 'chat' || jsonb_build_object('chat', (config->'chat') - 'threshold_features')),
        '{max_chatbots}', 
        '10'::jsonb
    ),
    '{guardrails}',
    '{
        "can_manage_topics": true,
        "can_customize_messages": true,
        "can_customize_thresholds": true,
        "can_use_smart_fallback": true,
        "can_use_escalate_fallback": false
    }'::jsonb
)
WHERE code = 'pro';

-- Ultra Plan Update
UPDATE plans 
SET config = jsonb_set(
    jsonb_set(
        (config - 'chat' || jsonb_build_object('chat', (config->'chat') - 'threshold_features')),
        '{max_chatbots}', 
        '100'::jsonb
    ),
    '{guardrails}',
    '{
        "can_manage_topics": true,
        "can_customize_messages": true,
        "can_customize_thresholds": true,
        "can_use_smart_fallback": true,
        "can_use_escalate_fallback": true
    }'::jsonb
)
WHERE code = 'ultra';
