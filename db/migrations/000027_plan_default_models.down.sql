-- Remove default_model from plan chat config

-- Remove from Free plan
UPDATE plans SET config = config #- '{chat,default_model}'
WHERE code = 'free';

-- Remove from Pro plan
UPDATE plans SET config = config #- '{chat,default_model}'
WHERE code = 'pro';

-- Remove from Ultra plan
UPDATE plans SET config = config #- '{chat,default_model}'
WHERE code = 'ultra';
