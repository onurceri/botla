package services

import (
	"context"
	"database/sql"
	"strings"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/logger"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

type WorkspaceService struct {
	DB  *sql.DB
	Log *logger.Logger
}

func NewWorkspaceService(db *sql.DB, log *logger.Logger) *WorkspaceService {
	return &WorkspaceService{
		DB:  db,
		Log: log,
	}
}

// CreateWorkspace creates a new workspace
func (s *WorkspaceService) CreateWorkspace(ctx context.Context, orgID, name, slug string, clientName *string) (*models.Workspace, error) {
	ws := &models.Workspace{
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		ClientName:     clientName,
	}

	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO workspaces (organization_id, name, slug, client_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, ws.OrganizationID, ws.Name, ws.Slug, ws.ClientName).Scan(&ws.ID, &ws.CreatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrWorkspaceSlugExists
		}
		return nil, pkgerrors.Wrapf(err, "creating workspace")
	}

	return ws, nil
}

// UpdateWorkspace updates workspace details
func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, wsID, name, slug string, clientName *string) error {
	_, err := s.DB.ExecContext(ctx, `
		UPDATE workspaces 
		SET name = $1, slug = $2, client_name = $3 
		WHERE id = $4
	`, name, slug, clientName, wsID)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return ErrWorkspaceSlugExists
		}
		return pkgerrors.Wrapf(err, "updating workspace")
	}
	return nil
}

// DeleteWorkspace deletes a workspace
func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, wsID string) error {
	// MI-003: Use transaction for atomicity
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return pkgerrors.Wrapf(err, "starting transaction")
	}
	defer func() { _ = tx.Rollback() }()

	// Get organization_id first
	var orgID string
	err = tx.QueryRowContext(ctx, "SELECT organization_id FROM workspaces WHERE id = $1", wsID).Scan(&orgID)
	if err != nil {
		return pkgerrors.Wrapf(err, "getting workspace org")
	}

	// Check total workspaces count
	var count int
	err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM workspaces WHERE organization_id = $1", orgID).Scan(&count)
	if err != nil {
		return pkgerrors.Wrapf(err, "counting workspaces")
	}

	if count <= 1 {
		return ErrLastWorkspace
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM workspaces WHERE id = $1", wsID)
	if err != nil {
		return pkgerrors.Wrapf(err, "deleting workspace")
	}

	if err := tx.Commit(); err != nil {
		return pkgerrors.Wrapf(err, "committing transaction")
	}
	return nil
}

// GetWorkspaces returns all workspaces of an organization
func (s *WorkspaceService) GetWorkspaces(ctx context.Context, orgID string) ([]*models.Workspace, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, organization_id, name, slug, client_name, created_at
		FROM workspaces
		WHERE organization_id = $1
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "querying workspaces")
	}
	defer func() { _ = rows.Close() }()

	var workspaces []*models.Workspace
	for rows.Next() {
		var ws models.Workspace
		if err := rows.Scan(&ws.ID, &ws.OrganizationID, &ws.Name, &ws.Slug, &ws.ClientName, &ws.CreatedAt); err != nil {
			return nil, pkgerrors.Wrapf(err, "scanning workspace row")
		}
		workspaces = append(workspaces, &ws)
	}
	// MI-006: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "iterating workspace rows")
	}
	return workspaces, nil
}

// GetWorkspace returns a workspace by ID
func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string) (*models.Workspace, error) {
	var ws models.Workspace
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, organization_id, name, slug, client_name, created_at
		FROM workspaces
		WHERE id = $1
	`, id).Scan(&ws.ID, &ws.OrganizationID, &ws.Name, &ws.Slug, &ws.ClientName, &ws.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "getting workspace")
	}
	return &ws, nil
}

// CheckAccess checks if user has access to the workspace via organization membership
func (s *WorkspaceService) CheckAccess(ctx context.Context, userID, workspaceID string) (*models.Workspace, error) {
	// 1. Get workspace
	ws, err := s.GetWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "checking workspace existence")
	}
	if ws == nil {
		return nil, nil // Workspace not found
	}

	// 2. Check organization membership directly
	var role string
	err = s.DB.QueryRowContext(ctx, `
		SELECT role FROM memberships 
		WHERE user_id = $1 AND organization_id = $2
	`, userID, ws.OrganizationID).Scan(&role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not a member
		}
		return nil, pkgerrors.Wrapf(err, "checking membership role")
	}

	return ws, nil
}
