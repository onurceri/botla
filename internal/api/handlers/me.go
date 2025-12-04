package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type MeResponse struct {
	ID               string  `json:"id"`
	Email            string  `json:"email"`
	FullName         *string `json:"full_name,omitempty"`
	AvatarURL        *string `json:"avatar_url,omitempty"`
	SubscriptionPlan string  `json:"subscription_plan"`
}

type MeHandlers struct{ DB *sql.DB }

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
	var fullName *string
	if u.FullName.Valid {
		fullName = &u.FullName.String
	}
	var avatar *string
	if u.AvatarURL.Valid {
		avatar = &u.AvatarURL.String
	}
	plan := "free"
	if u.SubscriptionPlan.Valid && u.SubscriptionPlan.String != "" {
		plan = u.SubscriptionPlan.String
	}
	res := MeResponse{ID: u.ID, Email: u.Email, FullName: fullName, AvatarURL: avatar, SubscriptionPlan: plan}
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    if err := json.NewEncoder(w).Encode(res); err != nil {
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
    }
}
