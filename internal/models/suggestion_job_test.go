package models

import (
	"testing"
	"time"
)

func TestSuggestionJobStatus_String(t *testing.T) {
	tests := []struct {
		status   SuggestionJobStatus
		expected string
	}{
		{SuggestionJobStatusPending, "pending"},
		{SuggestionJobStatusRunning, "running"},
		{SuggestionJobStatusCompleted, "completed"},
		{SuggestionJobStatusFailed, "failed"},
	}

	for _, tc := range tests {
		if got := tc.status.String(); got != tc.expected {
			t.Errorf("SuggestionJobStatus(%v).String() = %q, want %q", tc.status, got, tc.expected)
		}
	}
}

func TestSuggestionJobStatus_IsValid(t *testing.T) {
	tests := []struct {
		status   SuggestionJobStatus
		expected bool
	}{
		{SuggestionJobStatusPending, true},
		{SuggestionJobStatusRunning, true},
		{SuggestionJobStatusCompleted, true},
		{SuggestionJobStatusFailed, true},
		{SuggestionJobStatus("invalid"), false},
		{SuggestionJobStatus(""), false},
	}

	for _, tc := range tests {
		if got := tc.status.IsValid(); got != tc.expected {
			t.Errorf("SuggestionJobStatus(%v).IsValid() = %v, want %v", tc.status, got, tc.expected)
		}
	}
}

func TestSuggestionJob_IsTerminal(t *testing.T) {
	now := time.Now()
	completedJob := &SuggestionJob{
		ID:          "test-id",
		ChatbotID:   "chatbot-id",
		Status:      SuggestionJobStatusCompleted,
		CompletedAt: &now,
	}

	failedJob := &SuggestionJob{
		ID:          "test-id",
		ChatbotID:   "chatbot-id",
		Status:      SuggestionJobStatusFailed,
		CompletedAt: &now,
	}

	runningJob := &SuggestionJob{
		ID:        "test-id",
		ChatbotID: "chatbot-id",
		Status:    SuggestionJobStatusRunning,
		StartedAt: &now,
	}

	pendingJob := &SuggestionJob{
		ID:        "test-id",
		ChatbotID: "chatbot-id",
		Status:    SuggestionJobStatusPending,
	}

	if !completedJob.IsTerminal() {
		t.Errorf("completed job should be terminal")
	}

	if !failedJob.IsTerminal() {
		t.Errorf("failed job should be terminal")
	}

	if runningJob.IsTerminal() {
		t.Errorf("running job should not be terminal")
	}

	if pendingJob.IsTerminal() {
		t.Errorf("pending job should not be terminal")
	}
}
