package models

// User represents a user in the system.
type User struct {
	ID                  string  `json:"id"`
	Email               string  `json:"email"`
	FullName            *string `json:"full_name"`
	AvatarURL           *string `json:"avatar_url"`
	PlanID              *string `json:"plan_id"`
	PreferredLanguageID *string `json:"preferred_language_id"`
}
