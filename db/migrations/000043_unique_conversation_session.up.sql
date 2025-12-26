BEGIN;

-- 1. Remove duplicate conversations, keeping the one with the most messages, or latest updated_at
DELETE FROM conversations c1
USING conversations c2
WHERE c1.id < c2.id 
  AND c1.chatbot_id = c2.chatbot_id 
  AND c1.session_id = c2.session_id
  AND (c1.message_count < c2.message_count OR (c1.message_count = c2.message_count AND c1.updated_at < c2.updated_at));

-- 2. Add unique constraint
ALTER TABLE conversations
ADD CONSTRAINT conversations_chatbot_id_session_id_key UNIQUE (chatbot_id, session_id);

COMMIT;
