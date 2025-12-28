package db

import (
	"context"
	"database/sql"
	"time"

	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

type QueueStats struct {
	QueueName       string     `json:"queue_name"`
	PendingCount    int        `json:"pending_count"`
	ProcessingCount int        `json:"processing_count"`
	FailedCount     int        `json:"failed_count"`
	OldestPending   *time.Time `json:"oldest_pending"`
}

type StuckJob struct {
	ID            string    `json:"id"`
	QueueName     string    `json:"queue_name"`
	SourceID      string    `json:"source_id,omitempty"`
	ChatbotID     string    `json:"chatbot_id,omitempty"`
	Status        string    `json:"status"`
	StartedAt     time.Time `json:"started_at"`
	StuckDuration string    `json:"stuck_duration"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// GetQueueStats returns stats for scraping and processing queues
func GetQueueStats(ctx context.Context, db *sql.DB) ([]QueueStats, error) {
	var stats []QueueStats

	// 1. Data Sources Queue
	var dsStats QueueStats
	dsStats.QueueName = "source_processing"

	err := db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'processing'),
			COUNT(*) FILTER (WHERE status = 'error'),
			MIN(created_at) FILTER (WHERE status = 'pending')
		FROM data_sources
		WHERE deleted_at IS NULL AND status IN ('pending', 'processing', 'error')
	`).Scan(&dsStats.PendingCount, &dsStats.ProcessingCount, &dsStats.FailedCount, &dsStats.OldestPending)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "scan source queue stats")
	}
	stats = append(stats, dsStats)

	// 2. Discoverer Queue (pending_discovered_urls)
	var discoveryStats QueueStats
	discoveryStats.QueueName = "url_discovery"
	err = db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'pending'),
			MIN(discovered_at) FILTER (WHERE status = 'pending')
		FROM pending_discovered_urls
		WHERE status = 'pending'
	`).Scan(&discoveryStats.PendingCount, &discoveryStats.OldestPending)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "scan discovery queue stats")
	}
	stats = append(stats, discoveryStats)

	return stats, nil
}

// GetStuckJobs returns jobs that have been processing for too long
func GetStuckJobs(ctx context.Context, db *sql.DB, threshold time.Duration) ([]StuckJob, error) {
	// threshold is in milliseconds in the query for simplicity
	rows, err := db.QueryContext(ctx, `
		SELECT 
			id, 
			'source_processing' as queue_name,
			id as source_id,
			chatbot_id,
			status,
			COALESCE(last_refreshed_at, created_at) as started_at,
			EXTRACT(EPOCH FROM (NOW() - COALESCE(last_refreshed_at, created_at))) as duration_seconds,
			error_message
		FROM data_sources
		WHERE status = 'processing' 
		  AND COALESCE(last_refreshed_at, created_at) < NOW() - ($1 * INTERVAL '1 millisecond')
		  AND deleted_at IS NULL
	`, threshold.Milliseconds())
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query stuck jobs")
	}
	defer func() { _ = rows.Close() }()

	var stuckJobs []StuckJob
	for rows.Next() {
		var job StuckJob
		var durationSec float64
		var errMsg sql.NullString
		err := rows.Scan(
			&job.ID, &job.QueueName, &job.SourceID, &job.ChatbotID,
			&job.Status, &job.StartedAt, &durationSec, &errMsg,
		)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "scan stuck job")
		}
		if errMsg.Valid {
			job.ErrorMessage = errMsg.String
		}
		job.StuckDuration = formatSeconds(durationSec)
		stuckJobs = append(stuckJobs, job)
	}

	return stuckJobs, nil
}

func formatSeconds(seconds float64) string {
	d := time.Duration(seconds) * time.Second
	if d.Hours() >= 1 {
		return d.Round(time.Minute).String()
	}
	if d.Minutes() >= 1 {
		return d.Round(time.Second).String()
	}
	return d.Round(time.Second).String()
}
