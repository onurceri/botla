ALTER TABLE chatbots 
ADD COLUMN IF NOT EXISTS include_paths TEXT[] DEFAULT '{}',
ADD COLUMN IF NOT EXISTS exclude_paths TEXT[] DEFAULT '{}';

COMMENT ON COLUMN chatbots.include_paths IS 'Glob patterns for paths to include (empty = all)';
COMMENT ON COLUMN chatbots.exclude_paths IS 'Glob patterns for paths to exclude';
