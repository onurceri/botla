package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strings"
    "time"

    "github.com/onurceri/botla-co/internal/auth"
    "github.com/onurceri/botla-co/pkg/middleware"
)

type AuthHandlers struct {
    DB     *sql.DB
    Secret string
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
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    var req registerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    req.Email = strings.TrimSpace(req.Email)
    req.FullName = strings.TrimSpace(req.FullName)
    if req.Email == "" || req.Password == "" {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    var existing string
    err := h.DB.QueryRowContext(r.Context(), "SELECT id FROM users WHERE email=$1", req.Email).Scan(&existing)
    if err == nil && existing != "" {
        w.WriteHeader(http.StatusConflict)
        return
    }
    if err != nil && err != sql.ErrNoRows {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    hash, err := auth.HashPassword(req.Password)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    var userID string
    err = h.DB.QueryRowContext(
        r.Context(),
        "INSERT INTO users (email, password_hash, full_name) VALUES ($1,$2,$3) RETURNING id",
        req.Email, hash, req.FullName,
    ).Scan(&userID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    
    h.generateAndSendTokens(w, r, userID, http.StatusCreated)
}

func (h *AuthHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    var req loginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    req.Email = strings.TrimSpace(req.Email)
    if req.Email == "" || req.Password == "" {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    var userID string
    var hash string
    err := h.DB.QueryRowContext(r.Context(), "SELECT id, password_hash FROM users WHERE email=$1", req.Email).Scan(&userID, &hash)
    if err == sql.ErrNoRows {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    if !auth.VerifyPassword(hash, req.Password) {
        w.WriteHeader(http.StatusUnauthorized)
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
    
    // Check if token exists and is not revoked
    var revoked bool
    err = h.DB.QueryRowContext(r.Context(), "SELECT revoked FROM refresh_tokens WHERE token=$1", req.RefreshToken).Scan(&revoked)
    if err == sql.ErrNoRows || revoked {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    
    // Revoke old token (Rotation)
    _, err = h.DB.ExecContext(r.Context(), "UPDATE refresh_tokens SET revoked=true WHERE token=$1", req.RefreshToken)
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
    _, err := h.DB.ExecContext(r.Context(), "UPDATE refresh_tokens SET revoked=true WHERE token=$1", req.RefreshToken)
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
    
    _, err = h.DB.ExecContext(r.Context(), "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)", userID, refreshToken, time.Now().Add(7*24*time.Hour))
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(tokenResponse{Token: accessToken, RefreshToken: refreshToken})
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
    id, ok := middleware.UserIDFromContext(r.Context())
    if !ok || id == "" {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]any{"user_id": id, "status": "ok"})
}

