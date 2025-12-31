package services

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
	"github.com/onurceri/botla-co/pkg/logger"
)

type OrganizationService struct {
	DB  *sql.DB
	Log *logger.Logger
}

// RBAC role constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// isValidRole checks if the given role is valid
func isValidRole(role string) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleMember:
		return true
	}
	return false
}

// roleWeight returns the weight of a role for comparison
func roleWeight(role string) int {
	switch role {
	case RoleMember:
		return 1
	case RoleAdmin:
		return 2
	case RoleOwner:
		return 3
	}
	return 0
}

// hasHigherRole checks if newRole is higher than currentRole
func hasHigherRole(newRole, currentRole string) bool {
	return roleWeight(newRole) > roleWeight(currentRole)
}

func NewOrganizationService(db *sql.DB, log *logger.Logger) *OrganizationService {
	return &OrganizationService{
		DB:  db,
		Log: log,
	}
}

// CreateOrganization creates a new organization with the owner
func (s *OrganizationService) CreateOrganization(ctx context.Context, name, slug, ownerID string) (*models.Organization, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "starting transaction")
	}
	defer func() { _ = tx.Rollback() }()

	// Check if slug exists
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1)", slug).Scan(&exists)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "checking slug existence")
	}
	if exists {
		return nil, ErrOrgSlugExists
	}

	org := &models.Organization{
		Name:    name,
		Slug:    slug,
		OwnerID: ownerID,
		PlanID:  "agency_starter",
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO organizations (name, slug, owner_id, plan_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, org.Name, org.Slug, org.OwnerID, org.PlanID).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		return nil, pkgerrors.Wrapf(err, "inserting organization")
	}

	// Add owner as member with 'owner' role
	_, err = tx.ExecContext(ctx, `
		INSERT INTO memberships (organization_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, org.ID, ownerID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "adding owner membership")
	}

	if err = tx.Commit(); err != nil {
		return nil, pkgerrors.Wrapf(err, "committing transaction")
	}

	return org, nil
}

// UpdateOrganization updates organization details
func (s *OrganizationService) UpdateOrganization(ctx context.Context, id, name, slug string) error {
	var exists bool
	err := s.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM organizations WHERE slug = $1 AND id != $2)", slug, id).Scan(&exists)
	if err != nil {
		return pkgerrors.Wrapf(err, "checking slug existence")
	}
	if exists {
		return ErrOrgSlugExists
	}

	_, err = s.DB.ExecContext(ctx, "UPDATE organizations SET name = $1, slug = $2, updated_at = NOW() WHERE id = $3", name, slug, id)
	if err != nil {
		return pkgerrors.Wrapf(err, "updating organization")
	}
	return nil
}

// GetOrganization returns an organization by ID
func (s *OrganizationService) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	var org models.Organization
	var brandingBytes []byte
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, name, slug, owner_id, plan_id, branding, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`, id).Scan(
		&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.PlanID, &brandingBytes,
		&org.CreatedAt, &org.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "getting organization")
	}
	if len(brandingBytes) > 0 {
		var cb models.CustomBranding
		if err := json.Unmarshal(brandingBytes, &cb); err == nil {
			org.Branding = &cb
		}
	}
	return &org, nil
}

// DeleteOrganization deletes an organization
func (s *OrganizationService) DeleteOrganization(ctx context.Context, id string) error {
	// Get owner_id
	var ownerID string
	err := s.DB.QueryRowContext(ctx, "SELECT owner_id FROM organizations WHERE id = $1", id).Scan(&ownerID)
	if err != nil {
		return pkgerrors.Wrapf(err, "getting organization owner")
	}

	// Check total organizations count for the owner
	var count int
	err = s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations WHERE owner_id = $1", ownerID).Scan(&count)
	if err != nil {
		return pkgerrors.Wrapf(err, "counting owner organizations")
	}

	if count <= 1 {
		return ErrLastOrganization
	}

	_, err = s.DB.ExecContext(ctx, "DELETE FROM organizations WHERE id = $1", id)
	if err != nil {
		return pkgerrors.Wrapf(err, "deleting organization")
	}
	return nil
}

// GetUserOrganizations returns all organizations a user belongs to
func (s *OrganizationService) GetUserOrganizations(ctx context.Context, userID string) ([]*models.Organization, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT o.id, o.name, o.slug, o.owner_id, o.plan_id, o.branding, o.created_at, o.updated_at, m.role
		FROM organizations o
		JOIN memberships m ON o.id = m.organization_id
		WHERE m.user_id = $1
		ORDER BY o.name
	`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "querying user organizations")
	}
	defer func() { _ = rows.Close() }()

	var orgs []*models.Organization
	for rows.Next() {
		var org models.Organization
		var brandingBytes []byte // To handle JSONB
		err := rows.Scan(
			&org.ID, &org.Name, &org.Slug, &org.OwnerID, &org.PlanID, &brandingBytes,
			&org.CreatedAt, &org.UpdatedAt, &org.Role,
		)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "scanning organization row")
		}

		if len(brandingBytes) > 0 {
			var cb models.CustomBranding
			if err := json.Unmarshal(brandingBytes, &cb); err == nil {
				org.Branding = &cb
			}
		}
		orgs = append(orgs, &org)
	}
	// MI-006: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "iterating organization rows")
	}
	return orgs, nil
}

// CheckMembership checks if user is a member of the organization
func (s *OrganizationService) CheckMembership(ctx context.Context, userID, orgID string) (*models.Membership, error) {
	var m models.Membership
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, organization_id, user_id, role, created_at
		FROM memberships
		WHERE organization_id = $1 AND user_id = $2
	`, orgID, userID).Scan(&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, pkgerrors.Wrapf(err, "checking membership")
	}
	return &m, nil
}

// GetMembers returns all members of an organization
func (s *OrganizationService) GetMembers(ctx context.Context, orgID string) ([]*models.MembershipWithUser, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT m.id, m.organization_id, m.user_id, m.role, m.created_at,
		       u.id, u.email, u.full_name, u.avatar_url
		FROM memberships m
		JOIN users u ON m.user_id = u.id
		WHERE m.organization_id = $1
		ORDER BY m.created_at DESC
	`, orgID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "querying members")
	}
	defer func() { _ = rows.Close() }()

	var members []*models.MembershipWithUser
	for rows.Next() {
		var m models.MembershipWithUser
		err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt,
			&m.User.ID, &m.User.Email, &m.User.FullName, &m.User.AvatarURL,
		)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "scanning member row")
		}
		members = append(members, &m)
	}
	// MI-006: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "iterating member rows")
	}
	return members, nil
}

// AddMember adds a user to the organization
func (s *OrganizationService) AddMember(ctx context.Context, orgID, userID, role string) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO memberships (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role = $3
	`, orgID, userID, role)
	if err != nil {
		return pkgerrors.Wrapf(err, "adding member")
	}
	return nil
}

// RemoveMember removes a user from the organization with RBAC validation
func (s *OrganizationService) RemoveMember(ctx context.Context, orgID, callerID, targetUserID string) error {
	// 1. Get target's current membership
	targetMembership, err := s.CheckMembership(ctx, targetUserID, orgID)
	if err != nil {
		return pkgerrors.Wrapf(err, "checking target membership")
	}
	if targetMembership == nil {
		return ErrNotMember
	}

	// 2. Prevent removing the last owner
	if targetMembership.Role == RoleOwner {
		ownerCount, err2 := s.countOwnersInOrg(ctx, orgID)
		if err2 != nil {
			return err2
		}
		if ownerCount <= 1 {
			return ErrLastOwner
		}
	}

	// 3. Execute removal
	_, err = s.DB.ExecContext(ctx, "DELETE FROM memberships WHERE organization_id = $1 AND user_id = $2", orgID, targetUserID)
	if err != nil {
		return pkgerrors.Wrapf(err, "removing member")
	}
	return nil
}

// UpdateMemberRole updates a member's role with comprehensive RBAC validation
func (s *OrganizationService) UpdateMemberRole(ctx context.Context, orgID, callerID, targetUserID, newRole string) error {
	// 1. Validate role
	if !isValidRole(newRole) {
		return ErrInvalidRole
	}

	// 2. Get caller's membership
	callerMembership, err := s.CheckMembership(ctx, callerID, orgID)
	if err != nil {
		return pkgerrors.Wrapf(err, "checking caller membership")
	}
	if callerMembership == nil {
		return ErrNotMember
	}

	// 3. Get target's current membership
	targetMembership, err := s.CheckMembership(ctx, targetUserID, orgID)
	if err != nil {
		return pkgerrors.Wrapf(err, "checking target membership")
	}
	if targetMembership == nil {
		return ErrNotMember
	}

	// 4. Prevent self-promotion (member→admin, admin→owner)
	if callerID == targetUserID && hasHigherRole(newRole, callerMembership.Role) {
		return ErrCannotPromoteSelf
	}

	// 5. Prevent owner demotion if last owner
	if targetMembership.Role == RoleOwner && newRole != RoleOwner {
		ownerCount, err2 := s.countOwnersInOrg(ctx, orgID)
		if err2 != nil {
			return pkgerrors.Wrapf(err2, "checking owner count")
		}
		if ownerCount <= 1 {
			return ErrCannotDemoteOwner
		}
	}

	// 6. Only owners can assign owner role
	if newRole == RoleOwner && callerMembership.Role != RoleOwner {
		return ErrOnlyOwnersCanGrant
	}

	// 7. Execute update
	_, err = s.DB.ExecContext(ctx, "UPDATE memberships SET role = $1 WHERE organization_id = $2 AND user_id = $3", newRole, orgID, targetUserID)
	if err != nil {
		return pkgerrors.Wrapf(err, "updating member role")
	}
	return nil
}

// countOwnersInOrg returns the number of owners in an organization
func (s *OrganizationService) countOwnersInOrg(ctx context.Context, orgID string) (int, error) {
	var count int
	err := s.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM memberships WHERE organization_id = $1 AND role = 'owner'",
		orgID).Scan(&count)
	if err != nil {
		return 0, pkgerrors.Wrapf(err, "counting owners")
	}
	return count, nil
}
