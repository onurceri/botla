BEGIN;

ALTER TABLE conversations
DROP CONSTRAINT IF EXISTS conversations_chatbot_id_session_id_key;

COMMIT;
