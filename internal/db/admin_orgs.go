package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/onurceri/botla-co/internal/models"
)

type OrganizationFilter struct {
	Name   *string
	PlanID *string
}

func AdminListOrganizations(ctx context.Context, pool *sql.DB, filter OrganizationFilter, limit, offset int) ([]models.Organization, int, error) {
	query := `
		SELECT 
			o.id, o.name, o.slug, o.owner_id, o.plan_id, o.created_at, o.updated_at, 
			COUNT(*) OVER() as total_count,
			(SELECT COUNT(*) FROM memberships m WHERE m.organization_id = o.id) as user_count,
			(SELECT COUNT(*) FROM chatbots c WHERE c.organization_id = o.id) as chatbot_count
		FROM organizations o
		WHERE 1=1
	`
	args := []any{}
	argIdx := 1

	if filter.Name != nil {
		query += fmt.Sprintf(" AND o.name ILIKE $%d", argIdx)
		args = append(args, "%"+*filter.Name+"%")
		argIdx++
	}
	if filter.PlanID != nil {
		query += fmt.Sprintf(" AND o.plan_id = $%d", argIdx)
		args = append(args, *filter.PlanID)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := pool.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query organizations: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var orgs []models.Organization
	var totalCount int

	for rows.Next() {
		var o models.Organization
		err := rows.Scan(
			&o.ID,
			&o.Name,
			&o.Slug,
			&o.OwnerID,
			&o.PlanID,
			&o.CreatedAt,
			&o.UpdatedAt,
			&totalCount,
			&o.UserCount,
			&o.ChatbotCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan organization: %w", err)
		}

		orgs = append(orgs, o)
	}

	return orgs, totalCount, nil
}

func GetOrganizationByID(ctx context.Context, pool *sql.DB, id string) (*models.Organization, error) {
	var o models.Organization
	query := `
		SELECT 
			o.id, o.name, o.slug, o.owner_id, o.plan_id, o.created_at, o.updated_at,
			(SELECT COUNT(*) FROM memberships m WHERE m.organization_id = o.id) as user_count,
			(SELECT COUNT(*) FROM chatbots c WHERE c.organization_id = o.id) as chatbot_count
		FROM organizations o
		WHERE o.id = $1
	`
	err := pool.QueryRowContext(ctx, query, id).Scan(
		&o.ID,
		&o.Name,
		&o.Slug,
		&o.OwnerID,
		&o.PlanID,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.UserCount,
		&o.ChatbotCount,
	)
	if err != nil {
		return nil, err
	}

	return &o, nil
}
