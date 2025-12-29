package models

import (
	"time"
)

type SuggestionJobStatus string

const (
	SuggestionJobStatusPending   SuggestionJobStatus = "pending"
	SuggestionJobStatusRunning   SuggestionJobStatus = "running"
	SuggestionJobStatusCompleted SuggestionJobStatus = "completed"
	SuggestionJobStatusFailed    SuggestionJobStatus = "failed"
)

func (s SuggestionJobStatus) String() string {
	return string(s)
}

func (s SuggestionJobStatus) IsValid() bool {
	switch s {
	case SuggestionJobStatusPending, SuggestionJobStatusRunning, SuggestionJobStatusCompleted, SuggestionJobStatusFailed:
		return true
	}
	return false
}

type SuggestionJob struct {
	ID                 string              `json:"id"`
	ChatbotID          string              `json:"chatbot_id"`
	Status             SuggestionJobStatus `json:"status"`
	ErrorMessage       *string             `json:"error_message,omitempty"`
	SuggestedQuestions []string            `json:"suggested_questions,omitempty"`
	CreatedAt          time.Time           `json:"created_at"`
	StartedAt          *time.Time          `json:"started_at,omitempty"`
	CompletedAt        *time.Time          `json:"completed_at,omitempty"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

func (j *SuggestionJob) IsTerminal() bool {
	return j.Status == SuggestionJobStatusCompleted || j.Status == SuggestionJobStatusFailed
}
