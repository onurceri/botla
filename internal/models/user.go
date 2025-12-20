package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type User struct {
	ID                  string           `json:"id"`
	Email               string           `json:"email"`
	FullName            *string          `json:"full_name"`
	AvatarURL           *string          `json:"avatar_url"`
	PlanID              *string          `json:"plan_id"`
	PreferredLanguageID *string          `json:"preferred_language_id"`
	CreatedAt           time.Time        `json:"created_at"`
	OnboardingCompleted bool             `json:"onboarding_completed"`
	OnboardingStep      int              `json:"onboarding_step"`
	OnboardingSkipped   bool             `json:"onboarding_skipped"`
	OnboardingData      *OnboardingData  `json:"onboarding_data,omitempty"`
}

type OnboardingData struct {
	BotName        string `json:"bot_name,omitempty"`
	SourceType     string `json:"source_type,omitempty"`
	TextContent    string `json:"text_content,omitempty"`
	URLContent     string `json:"url_content,omitempty"`
	SystemPrompt   string `json:"system_prompt,omitempty"`
	WelcomeMessage string `json:"welcome_message,omitempty"`
	CreatedBotID   string `json:"created_bot_id,omitempty"`
}

// Value implements driver.Valuer for OnboardingData
func (o OnboardingData) Value() (driver.Value, error) {
	return json.Marshal(o)
}

// Scan implements sql.Scanner for OnboardingData
func (o *OnboardingData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, o)
}
