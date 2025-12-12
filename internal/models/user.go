package models

import "time"

type User struct {
	ID                  string    `json:"id"`
	Email               string    `json:"email"`
	FullName            *string   `json:"full_name"`
	AvatarURL           *string   `json:"avatar_url"`
	PlanID              *string   `json:"plan_id"`
	PreferredLanguageID *string   `json:"preferred_language_id"`
	CreatedAt           time.Time `json:"created_at"`
}
