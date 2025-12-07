package models

import "database/sql"

// User represents a user in the system.
type User struct {
	ID                  string
	Email               string
	FullName            sql.NullString
	AvatarURL           sql.NullString
	PlanID              sql.NullString
	PreferredLanguageID sql.NullString
}
