package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"html"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// MeResponse represents the /me endpoint response
type MeResponse struct {
	ID              string         `json:"id"`
	Email           string         `json:"email"`
	CreatedAt       time.Time      `json:"created_at"`
	FullName        *string        `json:"full_name,omitempty"`
	AvatarURL       *string        `json:"avatar_url,omitempty"`
	IsPlatformAdmin bool           `json:"is_platform_admin"`
	Organizations   []Organization `json:"organizations,omitempty"`
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// MeHandlers handles user profile endpoints
type MeHandlers struct {
	DB *sql.DB
}

// Me handles GET /me endpoint
func (h *MeHandlers) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u, err := db.GetUserByID(r.Context(), h.DB, uid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if u == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	orgs, err := h.getUserOrganizations(r.Context(), u.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := h.buildMeResponse(u, orgs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (h *MeHandlers) getUserOrganizations(ctx context.Context, userID string) ([]Organization, error) {
	rows, err := h.DB.QueryContext(ctx, `
		SELECT o.id, o.name, m.role
		FROM organizations o
		JOIN memberships m ON o.id = m.organization_id
		WHERE m.user_id = $1
		ORDER BY o.created_at
	`, userID)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "query user organizations")
	}
	defer func() { _ = rows.Close() }()

	var orgs []Organization
	for rows.Next() {
		var org Organization
		if err = rows.Scan(&org.ID, &org.Name, &org.Role); err != nil {
			return nil, pkgerrors.Wrapf(err, "scan user organization")
		}
		orgs = append(orgs, org)
	}
	if err = rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "user organizations rows err")
	}
	return orgs, nil
}

// buildMeResponse constructs the response from user and orgs data
func (h *MeHandlers) buildMeResponse(u *models.User, orgs []Organization) MeResponse {
	var sanitizedFullName *string
	if u.FullName != nil {
		escaped := html.EscapeString(*u.FullName)
		sanitizedFullName = &escaped
	}

	return MeResponse{
		ID:              u.ID,
		Email:           u.Email,
		CreatedAt:       u.CreatedAt,
		FullName:        sanitizedFullName,
		AvatarURL:       u.AvatarURL,
		IsPlatformAdmin: u.IsPlatformAdmin,
		Organizations:   orgs,
	}
}
