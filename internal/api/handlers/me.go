package handlers

import (
	"context"
	"encoding/json"
	"html"
	"net/http"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/middleware"
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
	UserRepo   repository.UserRepository
	OrgService *services.OrganizationService
}

// NewMeHandlers creates a new MeHandlers instance
func NewMeHandlers(userRepo repository.UserRepository, orgService *services.OrganizationService) *MeHandlers {
	return &MeHandlers{
		UserRepo:   userRepo,
		OrgService: orgService,
	}
}

// Me handles GET /me endpoint
func (h *MeHandlers) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u, err := h.UserRepo.GetByID(r.Context(), uid)
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
	orgs, err := h.OrgService.GetUserOrganizations(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]Organization, len(orgs))
	for i, org := range orgs {
		result[i] = Organization{
			ID:   org.ID,
			Name: org.Name,
			Role: org.Role,
		}
	}
	return result, nil
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
