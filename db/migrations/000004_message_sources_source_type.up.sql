-- Add source_type to message_sources. SaveMessageSources writes the source type of
-- each cited chunk (e.g. pdf/text/url), but the column was missing from the original
-- schema, so those inserts failed (the error was only logged and message sources were
-- silently not persisted). IF NOT EXISTS keeps this idempotent if the column was
-- already added out-of-band.
ALTER TABLE message_sources ADD COLUMN IF NOT EXISTS source_type text;
