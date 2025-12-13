package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/auth"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// hashToken creates a SHA-256 hash of the token for secure storage
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

type AuthHandlers struct {
	DB               *sql.DB
	Secret           string
	OrgService       *services.OrganizationService
	WorkspaceService *services.WorkspaceService
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type tokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// respondError sends a JSON error response
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "Email is required")
		return
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid email format")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters long")
		return
	}
	var existing string
	err := h.DB.QueryRowContext(r.Context(), "SELECT id FROM users WHERE email=$1", req.Email).Scan(&existing)
	if err == nil && existing != "" {
		respondError(w, http.StatusConflict, "Email already exists")
		return
	}
	if err != nil && err != sql.ErrNoRows {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	var userID string
	var freePlanID string
	_ = h.DB.QueryRowContext(r.Context(), "SELECT id FROM plans WHERE code='free'").Scan(&freePlanID)
	err = h.DB.QueryRowContext(
		r.Context(),
		"INSERT INTO users (email, password_hash, full_name, plan_id) VALUES ($1,$2,$3,$4) RETURNING id",
		req.Email, hash, req.FullName, func() interface{} {
			if freePlanID != "" {
				return freePlanID
			}
			return nil
		}(),
	).Scan(&userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Create default personal organization and workspace
	if h.OrgService != nil {
		// Use default language (tr) for localized strings
		cfg := langconfig.Get("tr")
		orgName := cfg.UserMessages.DefaultOrgName
		if req.FullName != "" {
			orgName = fmt.Sprintf(cfg.UserMessages.DefaultOrgNameFormat, req.FullName)
		}
		orgSlug := slugifyEmail(req.Email)
		org, err := h.OrgService.CreateOrganization(r.Context(), orgName, orgSlug, userID)
		if err == nil && org != nil && h.WorkspaceService != nil {
			_, _ = h.WorkspaceService.CreateWorkspace(r.Context(), org.ID, cfg.UserMessages.DefaultWorkspaceName, "default", nil)
		}
	}

	h.generateAndSendTokens(w, r, userID, http.StatusCreated)
}

// slugifyEmail converts email to URL-safe slug
func slugifyEmail(email string) string {
	// Take part before @
	parts := strings.Split(email, "@")
	slug := parts[0]
	// Replace non-alphanumeric with hyphen
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(strings.ToLower(slug), "-")
	// Trim hyphens
	slug = strings.Trim(slug, "-")
	return slug
}

func (h *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Email and password are required")
		return
	}
	var userID string
	var hash string
	err := h.DB.QueryRowContext(r.Context(), "SELECT id, password_hash FROM users WHERE LOWER(email) = LOWER($1)", req.Email).Scan(&userID, &hash)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if !auth.VerifyPassword(hash, req.Password) {
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	h.generateAndSendTokens(w, r, userID, http.StatusOK)
}

func (h *AuthHandlers) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims, err := auth.VerifyToken(h.Secret, req.RefreshToken)
	if err != nil || claims.TokenType != "refresh" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if token exists and is not revoked (lookup by hash)
	var revoked bool
	tokenHash := hashToken(req.RefreshToken)
	err = h.DB.QueryRowContext(r.Context(), "SELECT revoked FROM refresh_tokens WHERE token_hash=$1", tokenHash).Scan(&revoked)
	if err == sql.ErrNoRows || revoked {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Revoke old token (Rotation)
	_, err = h.DB.ExecContext(r.Context(), "UPDATE refresh_tokens SET revoked=true WHERE token_hash=$1", tokenHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.generateAndSendTokens(w, r, claims.UserID, http.StatusOK)
}

func (h *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Just revoke it, don't verify signature necessarily (or verify if you want strictness)
	tokenHash := hashToken(req.RefreshToken)
	_, err := h.DB.ExecContext(r.Context(), "UPDATE refresh_tokens SET revoked=true WHERE token_hash=$1", tokenHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) generateAndSendTokens(w http.ResponseWriter, r *http.Request, userID string, status int) {
	accessToken, err := auth.GenerateToken(h.Secret, userID, "access", 15*time.Minute)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.GenerateToken(h.Secret, userID, "refresh", 7*24*time.Hour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = h.DB.ExecContext(r.Context(), "INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)", userID, hashToken(refreshToken), time.Now().Add(7*24*time.Hour))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err = json.NewEncoder(w).Encode(tokenResponse{Token: accessToken, RefreshToken: refreshToken}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserIDFromContext(r.Context())
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{"user_id": id, "status": "ok"}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
