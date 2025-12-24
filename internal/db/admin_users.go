package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/onurceri/botla-co/internal/models"
)

type UserFilter struct {
	Email           *string
	IsPlatformAdmin *bool
	PlanID          *string
}

func AdminListUsers(ctx context.Context, pool *sql.DB, filter UserFilter, limit, offset int) ([]models.User, int, error) {
	query := `
		SELECT id, email, full_name, avatar_url, plan_id, preferred_language_id, 
		       created_at, onboarding_completed, onboarding_step, onboarding_skipped, 
		       onboarding_data, is_platform_admin, COUNT(*) OVER() as total_count
		FROM users
		WHERE deleted_at IS NULL
	`
	args := []any{}
	argIdx := 1

	if filter.Email != nil {
		query += fmt.Sprintf(" AND email ILIKE $%d", argIdx)
		args = append(args, "%"+*filter.Email+"%")
		argIdx++
	}
	if filter.IsPlatformAdmin != nil {
		query += fmt.Sprintf(" AND is_platform_admin = $%d", argIdx)
		args = append(args, *filter.IsPlatformAdmin)
		argIdx++
	}
	if filter.PlanID != nil {
		query += fmt.Sprintf(" AND plan_id = $%d", argIdx)
		args = append(args, *filter.PlanID)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query users: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var users []models.User
	var totalCount int

	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.FullName,
			&u.AvatarURL,
			&u.PlanID,
			&u.PreferredLanguageID,
			&u.CreatedAt,
			&u.OnboardingCompleted,
			&u.OnboardingStep,
			&u.OnboardingSkipped,
			&u.OnboardingData,
			&u.IsPlatformAdmin,
			&totalCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}

		users = append(users, u)
	}

	return users, totalCount, nil
}

func AdminUpdateUser(ctx context.Context, pool *sql.DB, userID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE users SET "
	args := []any{}
	argIdx := 1

	setParts := []string{}
	for k, v := range updates {
		// Basic validation of keys to prevent SQL injection (though we should use a proper builder)
		allowedKeys := map[string]bool{
			"full_name":         true,
			"plan_id":           true,
			"is_platform_admin": true,
		}
		if !allowedKeys[k] {
			continue
		}

		setParts = append(setParts, fmt.Sprintf("%s = $%d", k, argIdx))
		args = append(args, v)
		argIdx++
	}

	if len(setParts) == 0 {
		return nil
	}

	query += strings.Join(setParts, ", ")
	query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", argIdx)
	args = append(args, userID)

	_, err := pool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}
