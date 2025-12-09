package models

import (
	"time"
)

type Organization struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	OwnerID   string          `json:"owner_id"`
	PlanID    string          `json:"plan_id"`
	Branding  *CustomBranding `json:"branding,omitempty"`
	Role      string          `json:"role,omitempty"` // Role of the current user in this organization (only populated in list context)
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Membership struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Role           string    `json:"role"` // owner, admin, member
	CreatedAt      time.Time `json:"created_at"`
}

type MembershipWithUser struct {
	Membership
	User User `json:"user"`
}

type Workspace struct {
	ID             string             `json:"id"`
	OrganizationID string             `json:"organization_id"`
	Name           string             `json:"name"`
	Slug           string             `json:"slug"`
	ClientName     *string            `json:"client_name,omitempty"`
	Settings       *WorkspaceSettings `json:"settings,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
}

type WorkspaceSettings struct {
	// Add settings fields here as needed
}
