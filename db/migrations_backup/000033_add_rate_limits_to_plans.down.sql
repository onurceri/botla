-- Remove rate_limits from Free plan
UPDATE plans SET config = config - 'rate_limits'
WHERE code = 'free';

-- Remove rate_limits from Pro plan
UPDATE plans SET config = config - 'rate_limits'
WHERE code = 'pro';

-- Delete Ultra plan (if you want to keep it, just remove the rate_limits instead)
DELETE FROM plans WHERE code = 'ultra';
