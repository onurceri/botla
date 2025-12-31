package models

import (
	"encoding/json"
	"time"
)

// JobStatus represents the status of a training job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// String returns the string representation of JobStatus
func (s JobStatus) String() string {
	return string(s)
}

// IsValid checks if the JobStatus value is valid
func (s JobStatus) IsValid() bool {
	switch s {
	case JobStatusPending, JobStatusRunning, JobStatusCompleted, JobStatusFailed, JobStatusCancelled:
		return true
	}
	return false
}

// TrainingStep represents a step in the training pipeline
type TrainingStep string

const (
	StepFetchSource  TrainingStep = "fetch_source"
	StepParseContent TrainingStep = "parse_content"
	StepChunkText    TrainingStep = "chunk_text"
	StepEmbedChunks  TrainingStep = "embed_chunks"
	StepStoreVectors TrainingStep = "store_vectors"
)

// String returns the string representation of TrainingStep
func (s TrainingStep) String() string {
	return string(s)
}

// IsValid checks if the TrainingStep value is valid
func (s TrainingStep) IsValid() bool {
	switch s {
	case StepFetchSource, StepParseContent, StepChunkText, StepEmbedChunks, StepStoreVectors:
		return true
	}
	return false
}

// TrainingJob represents a training job for processing data sources
type TrainingJob struct {
	ID              string          `json:"id"`
	SourceID        string          `json:"source_id"`
	ChatbotID       string          `json:"chatbot_id"`
	Status          JobStatus       `json:"status"`
	CurrentStep     *TrainingStep   `json:"current_step,omitempty"`
	ProgressPercent int             `json:"progress_percent"`
	ErrorCode       *string         `json:"error_code,omitempty"`
	ErrorMessage    *string         `json:"error_message,omitempty"`
	FailedStep      *TrainingStep   `json:"failed_step,omitempty"`
	RetryCount      int             `json:"retry_count"`
	CreatedAt       time.Time       `json:"created_at"`
	StartedAt       *time.Time      `json:"started_at,omitempty"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
}

// StepProgress maps training steps to their progress percentages
var StepProgress = map[TrainingStep]int{
	StepFetchSource:  10,
	StepParseContent: 30,
	StepChunkText:    50,
	StepEmbedChunks:  80,
	StepStoreVectors: 100,
}

// StepOrder defines the chronological order of training steps
var StepOrder = []TrainingStep{
	StepFetchSource,
	StepParseContent,
	StepChunkText,
	StepEmbedChunks,
	StepStoreVectors,
}

// IsStepAtOrAfter returns true if step is at or after the target step in the pipeline
func IsStepAtOrAfter(step, target TrainingStep) bool {
	stepIdx := -1
	targetIdx := -1
	for i, s := range StepOrder {
		if s == step {
			stepIdx = i
		}
		if s == target {
			targetIdx = i
		}
	}
	return stepIdx >= targetIdx && stepIdx != -1 && targetIdx != -1
}

// IsTerminal returns true if the job is in a terminal state
func (j *TrainingJob) IsTerminal() bool {
	return j.Status == JobStatusCompleted || j.Status == JobStatusFailed || j.Status == JobStatusCancelled
}

// CanRetry returns true if the job can be retried
func (j *TrainingJob) CanRetry(maxRetries int) bool {
	return j.Status == JobStatusFailed && j.RetryCount < maxRetries
}
