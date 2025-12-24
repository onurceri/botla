package db

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/models"
)

func GetUserByID(ctx context.Context, pool *sql.DB, id string) (*models.User, error) {
	var u models.User
	var onboardingDataJSON []byte
	err := pool.QueryRowContext(ctx, `
        SELECT id, email, full_name, avatar_url, plan_id, preferred_language_id, created_at,
               onboarding_completed, onboarding_step, onboarding_skipped, onboarding_data,
               is_platform_admin
        FROM users WHERE id=$1 AND deleted_at IS NULL`, id).Scan(
		&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.PlanID, &u.PreferredLanguageID, &u.CreatedAt,
		&u.OnboardingCompleted, &u.OnboardingStep, &u.OnboardingSkipped, &onboardingDataJSON,
		&u.IsPlatformAdmin,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if len(onboardingDataJSON) > 0 {
		var data models.OnboardingData
		if err := data.Scan(onboardingDataJSON); err == nil {
			u.OnboardingData = &data
		}
	}
	return &u, nil
}

func GetUserByEmail(ctx context.Context, pool *sql.DB, email string) (*models.User, error) {
	var u models.User
	var onboardingDataJSON []byte
	err := pool.QueryRowContext(ctx, `
        SELECT id, email, full_name, avatar_url, plan_id, preferred_language_id, created_at,
               onboarding_completed, onboarding_step, onboarding_skipped, onboarding_data,
               is_platform_admin
        FROM users WHERE email=$1 AND deleted_at IS NULL`, email).Scan(
		&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.PlanID, &u.PreferredLanguageID, &u.CreatedAt,
		&u.OnboardingCompleted, &u.OnboardingStep, &u.OnboardingSkipped, &onboardingDataJSON,
		&u.IsPlatformAdmin,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if len(onboardingDataJSON) > 0 {
		var data models.OnboardingData
		if err := data.Scan(onboardingDataJSON); err == nil {
			u.OnboardingData = &data
		}
	}
	return &u, nil
}

// UpdateOnboardingState updates the user's onboarding progress
func UpdateOnboardingState(ctx context.Context, pool *sql.DB, userID string, step int, data *models.OnboardingData) error {
	dataJSON, err := data.Value()
	if err != nil {
		return err
	}
	_, err = pool.ExecContext(ctx, `
		UPDATE users 
		SET onboarding_step = $2, onboarding_data = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, step, dataJSON)
	return err
}

// SkipOnboarding marks the user's onboarding as skipped
func SkipOnboarding(ctx context.Context, pool *sql.DB, userID string) error {
	_, err := pool.ExecContext(ctx, `
		UPDATE users 
		SET onboarding_skipped = true, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, userID)
	return err
}

// CompleteOnboarding marks the user's onboarding as completed
func CompleteOnboarding(ctx context.Context, pool *sql.DB, userID, botID string) error {
	// Update onboarding data with created bot ID
	var data models.OnboardingData
	var dataJSON []byte
	err := pool.QueryRowContext(ctx, `
		SELECT onboarding_data FROM users WHERE id = $1
	`, userID).Scan(&dataJSON)
	if err == nil && len(dataJSON) > 0 {
		_ = data.Scan(dataJSON)
	}
	data.CreatedBotID = botID
	updatedDataJSON, err := data.Value()
	if err != nil {
		return err
	}

	_, err = pool.ExecContext(ctx, `
		UPDATE users 
		SET onboarding_completed = true, 
		    onboarding_step = 4,
		    onboarding_data = $2,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, userID, updatedDataJSON)
	return err
}
