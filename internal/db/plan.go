package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/onurceri/botla-co/internal/models"
)

func GetPlanByUserID(ctx context.Context, pool *sql.DB, userID string) (*models.Plan, error) {
	var p models.Plan
	err := pool.QueryRowContext(ctx, `
		SELECT p.id, p.code, p.status, p.billing_cycle, p.price, p.currency, p.trial_days, p.config, p.created_at, p.updated_at
		FROM plans p
		JOIN users u ON u.plan_id = p.id
		WHERE u.id = $1 AND u.deleted_at IS NULL AND p.deleted_at IS NULL
	`, userID).Scan(
		&p.ID, &p.Code, &p.Status, &p.BillingCycle, &p.Price, &p.Currency, &p.TrialDays, &p.Config, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get plan by user id: %w", err)
	}
	return &p, nil
}
