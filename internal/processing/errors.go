package processing

// Error codes for source processing failures
// These codes are returned to the frontend for user-friendly error messages
const (
	// URL-specific errors
	ErrCodeEmptyURL              = "ERR_EMPTY_URL"
	ErrCodeEmptyContent          = "ERR_EMPTY_CONTENT"
	ErrCodeScrapeFailedNetwork   = "ERR_SCRAPE_NETWORK"
	ErrCodeScrapeFailedTimeout   = "ERR_SCRAPE_TIMEOUT"
	ErrCodeScrapeFailedForbidden = "ERR_SCRAPE_FORBIDDEN"
	ErrCodeInvalidURL            = "ERR_INVALID_URL"
	ErrCodeDynamicRequired       = "ERR_DYNAMIC_REQUIRED" // Site requires JS rendering, upgrade needed

	// PDF-specific errors
	ErrCodeEmptyFilePath     = "ERR_EMPTY_FILE_PATH"
	ErrCodePDFDownloadFailed = "ERR_PDF_DOWNLOAD_FAILED"
	ErrCodePDFParseFailed    = "ERR_PDF_PARSE_FAILED"

	// Text-specific errors
	ErrCodeStorageRequired = "ERR_STORAGE_REQUIRED"

	// Common processing errors
	ErrCodeUnknownSourceType = "ERR_UNKNOWN_SOURCE_TYPE"
	ErrCodeChunkingFailed    = "ERR_CHUNKING_FAILED"
	ErrCodeEmbeddingFailed   = "ERR_EMBEDDING_FAILED"
	ErrCodeLLMNotSupported   = "ERR_LLM_NOT_SUPPORTED"
)

// ProcessingError represents a processing error with a structured error code
type ProcessingError struct {
	Msg string
}

func (e *ProcessingError) Error() string { return e.Msg }
