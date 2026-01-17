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

// PostgresOrganizationRepo implements OrganizationRepository using PostgreSQL.
type PostgresOrganizationRepo struct {
	pool *sql.DB
}

// Compile-time check that PostgresOrganizationRepo implements OrganizationRepository.
var _ OrganizationRepository = (*PostgresOrganizationRepo)(nil)

// NewPostgresOrganizationRepo creates a new PostgresOrganizationRepo instance.
func NewPostgresOrganizationRepo(pool *sql.DB) *PostgresOrganizationRepo {
	return &PostgresOrganizationRepo{pool: pool}
}

// scanOrganization scans an organization from rows.
func (r *PostgresOrganizationRepo) scanOrganization(rows *sql.Rows) (*models.Organization, error) {
	var o models.Organization
	if err := rows.Scan(
		&o.ID, &o.Name, &o.Slug, &o.OwnerID, &o.PlanID,
		&o.CreatedAt, &o.UpdatedAt, &o.UserCount, &o.ChatbotCount,
	); err != nil {
		return nil, pkgerrors.Wrapf(err, "scan organization")
	}
	return &o, nil
}

// GetByID retrieves an organization by its unique identifier.
func (r *PostgresOrganizationRepo) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	query := `
		SELECT 
			o.id, o.name, o.slug, o.owner_id, o.plan_id, o.created_at, o.updated_at,
			(SELECT COUNT(*) FROM memberships m WHERE m.organization_id = o.id) as user_count,
			(SELECT COUNT(*) FROM chatbots c WHERE c.organization_id = o.id) as chatbot_count
		FROM organizations o
		WHERE o.id = $1
	`
	row := r.pool.QueryRowContext(ctx, query, id)

	var o models.Organization
	err := row.Scan(
		&o.ID, &o.Name, &o.Slug, &o.OwnerID, &o.PlanID,
		&o.CreatedAt, &o.UpdatedAt, &o.UserCount, &o.ChatbotCount,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "get organization by id")
	}

	return &o, nil
}

// AdminList returns a paginated list of organizations for admin views.
func (r *PostgresOrganizationRepo) AdminList(ctx context.Context, filter OrganizationFilter, limit, offset int) ([]*models.Organization, int, error) {
	limit64, offset64, err := ValidatePagination(limit, offset)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "validate pagination")
	}

	query := psql.
		Select(
			"o.id", "o.name", "o.slug", "o.owner_id", "o.plan_id", "o.created_at", "o.updated_at",
			"(SELECT COUNT(*) FROM memberships m WHERE m.organization_id = o.id) as user_count",
			"(SELECT COUNT(*) FROM chatbots c WHERE c.organization_id = o.id) as chatbot_count",
		).
		From("organizations o").
		Where(sq.Expr("1=1"))

	if filter.Name != nil {
		query = query.Where(sq.ILike{"o.name": "%" + *filter.Name + "%"})
	}
	if filter.PlanID != nil {
		query = query.Where(sq.Eq{"o.plan_id": *filter.PlanID})
	}

	// Get total count
	countQuery, countArgs, err := query.ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build count query")
	}
	countQuery = fmt.Sprintf("SELECT COUNT(*) FROM (%s) as sub", countQuery)

	var total int
	if err := r.pool.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "count organizations")
	}

	// Add pagination
	query = query.OrderBy("o.created_at DESC").Limit(limit64).Offset(offset64)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "build list query")
	}

	rows, err := r.pool.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "query organizations")
	}
	defer func() { _ = rows.Close() }()

	var orgs []*models.Organization
	for rows.Next() {
		o, err := r.scanOrganization(rows)
		if err != nil {
			return nil, 0, err
		}
		orgs = append(orgs, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, pkgerrors.Wrapf(err, "scan organizations")
	}

	return orgs, total, nil
}

// GetPlatformOverviewStats returns aggregated platform statistics.
func (r *PostgresOrganizationRepo) GetPlatformOverviewStats(ctx context.Context) (*PlatformOverviewStats, error) {
	stats := &PlatformOverviewStats{}

	// Get total users
	err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "count users")
	}

	// Get total organizations
	err = r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations").Scan(&stats.TotalOrganizations)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "count organizations")
	}

	// Get total chatbots
	err = r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM chatbots WHERE deleted_at IS NULL").Scan(&stats.TotalChatbots)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "count chatbots")
	}

	// Get total messages
	err = r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM messages").Scan(&stats.TotalMessages)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "count messages")
	}

	return stats, nil
}

// GetTotalOrganizations returns the total count of organizations.
func (r *PostgresOrganizationRepo) GetTotalOrganizations(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations").Scan(&count)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "count organizations")
	}
	return count, nil
}

// GetTotalChatbots returns the total count of non-deleted chatbots.
func (r *PostgresOrganizationRepo) GetTotalChatbots(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRowContext(ctx, "SELECT COUNT(*) FROM chatbots WHERE deleted_at IS NULL").Scan(&count)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "count chatbots")
	}
	return count, nil
}
