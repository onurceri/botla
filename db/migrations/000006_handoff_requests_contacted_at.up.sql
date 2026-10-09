-- Add contacted_at to handoff_requests. The widget's "leave your email" step
-- (POST /public/chatbots/{id}/handoff/{requestId}/contact) stores the visitor's
-- address in the existing user_email column and records when it was submitted
-- here. Matches the table's other timestamps (timestamp without time zone).
-- IF NOT EXISTS keeps this idempotent if the column was added out-of-band.
ALTER TABLE handoff_requests ADD COLUMN IF NOT EXISTS contacted_at timestamp without time zone;
