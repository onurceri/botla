package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/middleware"
)

type MeResponse struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	FullName        *string `json:"full_name,omitempty"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	PlanID          string  `json:"plan_id"`
	PlanCode        string  `json:"plan_code"`
	PlanName        *string `json:"plan_name,omitempty"`
	PlanDescription *string `json:"plan_description,omitempty"`
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
	var planCode string
	var planName *string
	var planDesc *string
	langID := u.PreferredLanguageID
	if !langID.Valid {
		_ = h.DB.QueryRow(`SELECT id FROM languages WHERE code='tr-TR'`).Scan(&langID.String)
		if langID.String != "" {
			langID.Valid = true
		}
	}
	var name sql.NullString
	var desc sql.NullString
	var planID string
	if !u.PlanID.Valid || u.PlanID.String == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	planID = u.PlanID.String
	err = h.DB.QueryRow(`
        SELECT p.code, pt.name, pt.description
        FROM plans p
        LEFT JOIN plan_translations pt ON pt.plan_id=p.id AND pt.language_id=$2
        WHERE p.id=$1
    `, planID, func() interface{} {
		if langID.Valid {
			return langID.String
		}
		return nil
	}()).Scan(&planCode, &name, &desc)
	if err != nil {
		planCode = "free"
	}
	if name.Valid {
		s := name.String
		planName = &s
	}
	if desc.Valid {
		s := desc.String
		planDesc = &s
	}
	res := MeResponse{ID: u.ID, Email: u.Email, FullName: fullName, AvatarURL: avatar, PlanID: planID, PlanCode: planCode, PlanName: planName, PlanDescription: planDesc}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
