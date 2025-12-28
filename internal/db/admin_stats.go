package db

import (
	"context"
	"database/sql"

	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

type PlatformOverviewStats struct {
	TotalUsers         int `json:"total_users"`
	TotalOrganizations int `json:"total_organizations"`
	TotalChatbots      int `json:"total_chatbots"`
	TotalMessages      int `json:"total_messages"`
}

func GetPlatformOverviewStats(ctx context.Context, pool *sql.DB) (*PlatformOverviewStats, error) {
	stats := &PlatformOverviewStats{}

	// Get total users
	err := pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "count users")
	}

	// Get total organizations
	err = pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations").Scan(&stats.TotalOrganizations)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "count organizations")
	}

	// Get total chatbots
	err = pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM chatbots WHERE deleted_at IS NULL").Scan(&stats.TotalChatbots)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "count chatbots")
	}

	// Get total messages
	err = pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages").Scan(&stats.TotalMessages)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "count messages")
	}

	return stats, nil
}
