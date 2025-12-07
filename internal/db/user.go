package db

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/models"
)

func GetUserByID(ctx context.Context, pool *sql.DB, id string) (*models.User, error) {
	var u models.User
	err := pool.QueryRowContext(ctx, `
        SELECT id, email, full_name, avatar_url, plan_id, preferred_language_id
        FROM users WHERE id=$1 AND deleted_at IS NULL`, id).Scan(
		&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.PlanID, &u.PreferredLanguageID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

