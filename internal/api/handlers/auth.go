package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/auth"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/langconfig"
	"github.com/onurceri/botla-app/pkg/middleware"
	"github.com/onurceri/botla-app/pkg/policy"
)

// sqlExecer is the subset of *sql.DB / *sql.Tx used for single statements.
type sqlExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// revokeActiveRefreshToken marks the refresh token with the given hash as
// revoked if, and only if, it is currently active. The conditional UPDATE is a
// single atomic statement, so when several requests present the same token
// concurrently exactly one of them observes ok == true.
func revokeActiveRefreshToken(ctx context.Context, db sqlExecer, tokenHash string) (bool, error) {
	res, err := db.ExecContext(ctx,
		"UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1 AND revoked IS NOT TRUE",
		tokenHash)
	if err != nil {
		return false, fmt.Errorf("revoke refresh token: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("revoke refresh token rows affected: %w", err)
	}
	return n == 1, nil
}

// hashToken creates a SHA-256 hash of the token for secure storage
func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// isStrongPassword validates password meets complexity requirements:
// - At least 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
// - At least one special character (@$!%*?&)
func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	specialChars := "@$!%*?&"

	for _, c := range password {
		switch {
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case 'a' <= c && c <= 'z':
			hasLower = true
		case '0' <= c && c <= '9':
			hasDigit = true
		case strings.ContainsRune(specialChars, c):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

type AuthHandlers struct {
	DB               *sql.DB
	Secret           string
	CookieSecure     bool
	CookieDomain     string
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

func (h *AuthHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteErrorCode(w, http.StatusMethodNotAllowed, api.ErrCodeMethodNotAllowed)
		return
	}
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidRequestBody)
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)
	if req.Email == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrEmailRequired)
		return
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidEmailFormat)
		return
	}
	if len(req.Password) < 8 {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrPasswordTooShort)
		return
	}
	if !isStrongPassword(req.Password) {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrPasswordWeak)
		return
	}
	var existing string
	// Match case-insensitively, like login does, so Foo@x.com and foo@x.com
	// cannot become two accounts.
	err := h.DB.QueryRowContext(r.Context(), "SELECT id FROM users WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL LIMIT 1", req.Email).Scan(&existing)
	if err == nil && existing != "" {
		api.WriteErrorCode(w, http.StatusConflict, api.ErrEmailExists)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrDatabaseError)
		return
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrFailedToHashPassword)
		return
	}
	var userID string
	var freePlanID string
	_ = h.DB.QueryRowContext(r.Context(), "SELECT id FROM plans WHERE code=$1", policy.PlanFree.String()).Scan(&freePlanID)
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
		// A concurrent registration can win the race between the email pre-check above
		// and this INSERT; map the unique-violation to 409 instead of a generic 500.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			api.WriteErrorCode(w, http.StatusConflict, api.ErrEmailExists)
			return
		}
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrFailedToCreateUser)
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

	h.generateAndSendTokens(w, r, userID, false, http.StatusCreated)
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
		api.WriteErrorCode(w, http.StatusMethodNotAllowed, api.ErrCodeMethodNotAllowed)
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrInvalidRequestBody)
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		api.WriteErrorCode(w, http.StatusBadRequest, api.ErrEmailAndPasswordReq)
		return
	}
	var userID string
	var hash string
	var isPlatformAdmin bool
	err := h.DB.QueryRowContext(r.Context(), "SELECT id, password_hash, is_platform_admin FROM users WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL", req.Email).Scan(&userID, &hash, &isPlatformAdmin)
	if err == sql.ErrNoRows {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrInvalidCredentials)
		return
	}
	if err != nil {
		api.WriteErrorCode(w, http.StatusInternalServerError, api.ErrDatabaseError)
		return
	}
	if !auth.VerifyPassword(hash, req.Password) {
		api.WriteErrorCode(w, http.StatusUnauthorized, api.ErrInvalidCredentials)
		return
	}

	h.generateAndSendTokens(w, r, userID, isPlatformAdmin, http.StatusOK)
}

func (h *AuthHandlers) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req refreshRequest
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.RefreshToken == "" {
		c, err := r.Cookie("botla_refresh_token")
		if err == nil {
			req.RefreshToken = c.Value
		}
	}

	if req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims, err := auth.VerifyToken(h.Secret, req.RefreshToken)
	if err != nil || claims.TokenType != "refresh" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Rotation: atomically revoke the presented token. Only the request that
	// flips revoked from false to true may mint a new token pair; unknown,
	// already-revoked, or concurrently-rotated tokens are rejected.
	ok, err := revokeActiveRefreshToken(r.Context(), h.DB, hashToken(req.RefreshToken))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.generateAndSendTokens(w, r, claims.UserID, claims.IsPlatformAdmin, http.StatusOK)
}

func (h *AuthHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req refreshRequest
	if r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.RefreshToken == "" {
		c, err := r.Cookie("botla_refresh_token")
		if err == nil {
			req.RefreshToken = c.Value
		}
	}

	if req.RefreshToken == "" {
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

	// Determine SameSite mode based on environment
	sameSite := http.SameSiteLaxMode
	if h.CookieSecure {
		sameSite = http.SameSiteNoneMode
	}

	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "botla_token",
		Value:    "",
		Path:     "/",
		Domain:   h.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: sameSite,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "botla_refresh_token",
		Value:    "",
		Path:     "/",
		Domain:   h.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: sameSite,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) generateAndSendTokens(w http.ResponseWriter, r *http.Request, userID string, isPlatformAdmin bool, status int) {
	accessToken, err := auth.GenerateToken(h.Secret, userID, isPlatformAdmin, "access", 1*time.Hour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.GenerateToken(h.Secret, userID, isPlatformAdmin, "refresh", 7*24*time.Hour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = h.DB.ExecContext(r.Context(), "INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)", userID, hashToken(refreshToken), time.Now().Add(7*24*time.Hour))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Determine SameSite mode based on environment
	// SameSite=None requires Secure=true (production)
	// SameSite=Lax is used in development when Secure=false
	sameSite := http.SameSiteLaxMode
	if h.CookieSecure {
		sameSite = http.SameSiteNoneMode
	}

	// Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "botla_token",
		Value:    accessToken,
		Path:     "/",
		Domain:   h.CookieDomain,
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: sameSite,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "botla_refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   h.CookieDomain,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: sameSite,
	})

	api.WriteJSON(w, status, tokenResponse{Token: accessToken, RefreshToken: refreshToken})
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.UserIDFromContext(r.Context())
	if !ok || id == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	api.WriteJSON(w, http.StatusOK, map[string]any{"user_id": id, "status": "ok"})
}
