// Package repository provides data access layer implementations.
package repository

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/onurceri/botla-app/internal/models"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
)

// PostgresUserRepo implements UserRepository using PostgreSQL.
type PostgresUserRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresUserRepo implements UserRepository.
var _ UserRepository = (*PostgresUserRepo)(nil)

// NewPostgresUserRepo creates a new PostgresUserRepo instance.
func NewPostgresUserRepo(pool *sql.DB) *PostgresUserRepo {
	return &PostgresUserRepo{pool: pool}
}

// scanUser scans a single user row from the result set.
func (r *PostgresUserRepo) scanUser(rows *sql.Rows) (*models.User, error) {
	var u models.User
	var onboardingDataJSON []byte
	if err := rows.Scan(
		&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.PlanID, &u.PreferredLanguageID,
		&u.CreatedAt, &u.OnboardingCompleted, &u.OnboardingStep, &u.OnboardingSkipped,
		&onboardingDataJSON, &u.IsPlatformAdmin,
	); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan user")
	}
	if len(onboardingDataJSON) > 0 {
		var data models.OnboardingData
		if err := data.Scan(onboardingDataJSON); err == nil {
			u.OnboardingData = &data
		}
	}
	return &u, nil
}

// scanUsers scans multiple user rows from the result set.
func (r *PostgresUserRepo) scanUsers(rows *sql.Rows) ([]*models.User, error) {
	var out []*models.User
	for rows.Next() {
		u, err := r.scanUser(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan users rows")
	}
	return out, nil
}

// GetByID retrieves a user by their unique identifier.
// Returns nil, nil if the user is not found.
func (r *PostgresUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	query, args, err := psql.
		Select(
			"id", "email", "full_name", "avatar_url", "plan_id", "preferred_language_id",
			"created_at", "onboarding_completed", "onboarding_step", "onboarding_skipped",
			"onboarding_data", "is_platform_admin",
		).
		From("users").
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get by id query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query user")
	}
	defer func() { _ = rows.Close() }()

	users, err := r.scanUsers(rows)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

// AdminListUsers returns a paginated list of users for admin views.
func (r *PostgresUserRepo) AdminListUsers(ctx context.Context, filter UserFilter, limit, offset int) ([]*models.User, int, error) {
	limit64, offset64, err := ValidatePagination(limit, offset)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "validate pagination")
	}

	query := psql.
		Select(
			"id", "email", "full_name", "avatar_url", "plan_id", "preferred_language_id",
			"created_at", "onboarding_completed", "onboarding_step", "onboarding_skipped",
			"onboarding_data", "is_platform_admin",
		).
		From("users").
		Where(sq.Eq{"deleted_at": nil})

	if filter.Email != nil {
		query = query.Where(sq.ILike{"email": "%" + *filter.Email + "%"})
	}
	if filter.IsPlatformAdmin != nil {
		query = query.Where(sq.Eq{"is_platform_admin": *filter.IsPlatformAdmin})
	}
	if filter.PlanID != nil {
		query = query.Where(sq.Eq{"plan_id": *filter.PlanID})
	}

	// Get total count
	countQuery, countArgs, err := query.ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build count query")
	}
	countQuery = fmt.Sprintf("SELECT COUNT(*) FROM (%s) as sub", countQuery)

	var total int
	if err := r.pool.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "count users")
	}

	// Add pagination
	query = query.OrderBy("created_at DESC").Limit(limit64).Offset(offset64)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build list query")
	}

	rows, err := r.pool.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "query users")
	}
	defer func() { _ = rows.Close() }()

	users, err := r.scanUsers(rows)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// AdminUpdateUser updates a user's fields for admin operations.
func (r *PostgresUserRepo) AdminUpdateUser(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	query := psql.Update("users")
	for k, v := range updates {
		switch k {
		case "full_name":
			query = query.Set("full_name", v)
		case "plan_id":
			query = query.Set("plan_id", v)
		case "is_platform_admin":
			query = query.Set("is_platform_admin", v)
		}
	}
	query = query.Set("updated_at", sq.Expr("NOW())"))
	query = query.Where(sq.Eq{"id": id}).Where(sq.Eq{"deleted_at": nil})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update query")
	}

	_, err = r.pool.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update user")
	}
	return nil
}

// GetByEmail retrieves a user by their email address.
func (r *PostgresUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query, args, err := psql.
		Select(
			"id", "email", "full_name", "avatar_url", "plan_id", "preferred_language_id",
			"created_at", "onboarding_completed", "onboarding_step", "onboarding_skipped",
			"onboarding_data", "is_platform_admin",
		).
		From("users").
		Where(sq.Eq{"email": email}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "build get by email query")
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query user by email")
	}
	defer func() { _ = rows.Close() }()

	users, err := r.scanUsers(rows)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	return users[0], nil
}

// UpdateOnboardingState updates the user's onboarding progress.
func (r *PostgresUserRepo) UpdateOnboardingState(ctx context.Context, userID string, step int, data *models.OnboardingData) error {
	dataJSON, err := data.Value()
	if err != nil {
		return pkgerrors.Wrapf(err, "serialize onboarding data")
	}

	query, args, err := psql.
		Update("users").
		Set("onboarding_step", step).
		Set("onboarding_data", dataJSON).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build update onboarding query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "update onboarding state")
	}
	return nil
}

// SkipOnboarding marks the user's onboarding as skipped.
func (r *PostgresUserRepo) SkipOnboarding(ctx context.Context, userID string) error {
	query, args, err := psql.
		Update("users").
		Set("onboarding_skipped", true).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build skip onboarding query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "skip onboarding")
	}
	return nil
}

// CompleteOnboarding marks the user's onboarding as completed.
func (r *PostgresUserRepo) CompleteOnboarding(ctx context.Context, userID, botID string) error {
	// Get current onboarding data
	var dataJSON []byte
	err := r.pool.QueryRowContext(ctx, "SELECT onboarding_data FROM users WHERE id=$1", userID).Scan(&dataJSON)
	if err != nil && err != sql.ErrNoRows {
		return pkgerrors.Wrapf(err, "get onboarding data")
	}

	var data models.OnboardingData
	if err == nil && len(dataJSON) > 0 {
		_ = data.Scan(dataJSON)
	}
	data.CreatedBotID = botID

	updatedDataJSON, err := data.Value()
	if err != nil {
		return pkgerrors.Wrapf(err, "serialize updated onboarding data")
	}

	query, args, err := psql.
		Update("users").
		Set("onboarding_completed", true).
		Set("onboarding_step", 4).
		Set("onboarding_data", updatedDataJSON).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": userID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return pkgerrors.Wrapf(err, "build complete onboarding query")
	}

	_, err = r.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return pkgerrors.Wrapf(err, "complete onboarding")
	}
	return nil
}

// GetTotalUsers returns the total count of non-deleted users.
func (r *PostgresUserRepo) GetTotalUsers(ctx context.Context) (int, error) {
	query, args, err := psql.
		Select("COUNT(*)").
		From("users").
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "build count users query")
	}

	var count int
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, pkgerrors.Wrapf(err, "count users")
	}
	return count, nil
}

// GetTotalMessages returns the total count of messages.
func (r *PostgresUserRepo) GetTotalMessages(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM messages"
	var count int
	if err := r.pool.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, pkgerrors.Wrapf(err, "count messages")
	}
	return count, nil
}
