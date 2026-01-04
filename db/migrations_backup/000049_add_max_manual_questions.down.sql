-- Remove max_manual_questions from plan configs
UPDATE plans SET config = config #- '{chat,max_manual_questions}' WHERE code IN ('free', 'pro', 'ultra');
