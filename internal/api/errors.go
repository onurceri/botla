package api

import (
	"encoding/json"
	"net/http"

	"github.com/onurceri/botla-co/pkg/langconfig"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

// WriteError writes a standardized JSON error response
func WriteError(w http.ResponseWriter, status int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message, Code: code})
}

// WriteErrorWithDetails writes a JSON error response with additional details
func WriteErrorWithDetails(w http.ResponseWriter, status int, message string, code string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message, Code: code, Details: details})
}

func WriteLocalizedError(w http.ResponseWriter, status int, code string, cfg langconfig.LanguageConfig) {
	msg := cfg.ResponseTemplates.Errors[code]
	if msg == "" {
		msg = cfg.ResponseTemplates.ErrorMessage
	}
	WriteError(w, status, msg, code)
}

func WriteLocalizedErrorWithDetails(w http.ResponseWriter, status int, code string, details any, cfg langconfig.LanguageConfig) {
	msg := cfg.ResponseTemplates.Errors[code]
	if msg == "" {
		msg = cfg.ResponseTemplates.ErrorMessage
	}
	WriteErrorWithDetails(w, status, msg, code, details)
}

// Common error codes for consistent API responses
const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeTooManyRequests     = "TOO_MANY_REQUESTS"
	ErrCodePaymentRequired     = "PAYMENT_REQUIRED"
	ErrCodeInternalError       = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	ErrCodeRequestEntityTooBig = "REQUEST_ENTITY_TOO_LARGE"
)

// Domain-specific error codes
const (
	ErrMonthlyTokensExceeded     = "ERR_MONTHLY_TOKENS_EXCEEDED" //nolint:gosec
	ErrNameAndActionTypeRequired = "ERR_NAME_AND_ACTION_TYPE_REQUIRED"
	ErrPdfLimitReached           = "ERR_PDF_LIMIT_REACHED"
	ErrFileTooLarge              = "ERR_FILE_TOO_LARGE"
	ErrReaddCooldownActive       = "ERR_READD_COOLDOWN_ACTIVE"
	ErrDuplicateURL              = "ERR_DUPLICATE_URL"
	ErrOnlyURLRefresh            = "ERR_ONLY_URL_REFRESH"
	ErrSourceAlreadyProcessing   = "ERR_SOURCE_ALREADY_PROCESSING"
	ErrPlanRefreshUnavailable    = "ERR_PLAN_REFRESH_UNAVAILABLE"
	ErrMonthlyRefreshExceeded    = "ERR_MONTHLY_REFRESH_EXCEEDED"
	ErrRefreshCooldownActive     = "ERR_REFRESH_COOLDOWN_ACTIVE"
	ErrInvalidRequestBody        = "ERR_INVALID_REQUEST_BODY"
	ErrNoURLsProvided            = "ERR_NO_URLS_PROVIDED"
	ErrURLLimitReached           = "ERR_URL_LIMIT_REACHED"
	ErrMonthlyIngestionExceeded  = "ERR_MONTHLY_INGESTION_EXCEEDED"
	ErrSitemapParseFailed        = "ERR_SITEMAP_PARSE_FAILED"
	ErrInvalidStatus             = "ERR_INVALID_STATUS"
	ErrMaxChatbotsExceeded       = "ERR_MAX_CHATBOTS_EXCEEDED"
)
