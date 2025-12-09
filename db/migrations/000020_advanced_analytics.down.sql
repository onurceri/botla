DROP TABLE IF EXISTS unanswered_queries;

ALTER TABLE messages
DROP COLUMN IF EXISTS sources_used,
DROP COLUMN IF EXISTS confidence_score;

DROP INDEX IF EXISTS idx_analytics_date_range;

ALTER TABLE analytics 
DROP COLUMN IF EXISTS avg_response_time_ms,
DROP COLUMN IF EXISTS handoff_count,
DROP COLUMN IF EXISTS total_tokens_used;
