package config

import "time"

// Default values used throughout the application
const (
	DefaultLanguage        = "tr"
	DefaultMaxChars        = 1000
	DefaultChunkSize       = 512
	DefaultMaxRetries      = 4
	DefaultRetryBaseDelay  = 200 * time.Millisecond
	DefaultFilesPerBot     = 5
	DefaultURLsPerBot      = 5
	DefaultMaxMonthlyIngest = 50
	DefaultFileSizeMB      = 10
	DefaultChatTimeout     = 20 * time.Second
)
