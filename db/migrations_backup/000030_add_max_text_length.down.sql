-- Remove max_text_length from plan files config

-- Remove from Free plan
UPDATE plans SET config = config #- '{files,max_text_length}'
WHERE code = 'free';

-- Remove from Pro plan
UPDATE plans SET config = config #- '{files,max_text_length}'
WHERE code = 'pro';

-- Remove from Ultra plan
UPDATE plans SET config = config #- '{files,max_text_length}'
WHERE code = 'ultra';
