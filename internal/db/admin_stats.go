package db

import (
	"context"
	"database/sql"
	"fmt"
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
		return nil, fmt.Errorf("count users: %w", err)
	}

	// Get total organizations
	err = pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations").Scan(&stats.TotalOrganizations)
	if err != nil {
		return nil, fmt.Errorf("count organizations: %w", err)
	}

	// Get total chatbots
	err = pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM chatbots WHERE deleted_at IS NULL").Scan(&stats.TotalChatbots)
	if err != nil {
		return nil, fmt.Errorf("count chatbots: %w", err)
	}

	// Get total messages
	err = pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages").Scan(&stats.TotalMessages)
	if err != nil {
		return nil, fmt.Errorf("count messages: %w", err)
	}

	return stats, nil
}
