-- Remove is_discovered column
ALTER TABLE data_sources DROP COLUMN IF EXISTS is_discovered;
