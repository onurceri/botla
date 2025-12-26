package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/langconfig"
)

// ErrorResponse is the standard error response format.
// The 'code' field is the machine-readable error code that frontend uses for localization.
// The 'error' field is deprecated and will be removed - use 'code' instead.
type ErrorResponse struct {
	Code    string `json:"code"`
	Status  int    `json:"status"`
	Error   string `json:"error,omitempty"`   // Deprecated: use code for translations
	Details any    `json:"details,omitempty"` // Optional extra info
}

// WriteErrorCode writes a code-only error response. Frontend handles localization.
// This is the preferred method for all new code.
func WriteErrorCode(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Code: code, Status: status})
}

// WriteErrorCodeWithDetails writes a code-only error with additional context.
func WriteErrorCodeWithDetails(w http.ResponseWriter, status int, code string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Code: code, Status: status, Details: details})
}

// WriteError writes a standardized JSON error response.
// Deprecated: Use WriteErrorCode instead. Frontend handles localization.
func WriteError(w http.ResponseWriter, status int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message, Code: code, Status: status})
}

// WriteErrorWithDetails writes a JSON error response with additional details.
// Deprecated: Use WriteErrorCodeWithDetails instead.
func WriteErrorWithDetails(w http.ResponseWriter, status int, message string, code string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message, Code: code, Status: status, Details: details})
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteLocalizedError writes a localized error using langconfig.
// Deprecated: Use WriteErrorCode instead. Frontend handles localization.
// This should only be used for widget/chat responses where backend needs to send translated content.
func WriteLocalizedError(w http.ResponseWriter, status int, code string, cfg langconfig.LanguageConfig) {
	msg := cfg.UserMessages.Errors[code]
	if msg == "" {
		msg = cfg.UserMessages.ErrorMessage
	}
	WriteError(w, status, msg, code)
}

// WriteLocalizedErrorWithDetails writes a localized error with details using langconfig.
// Deprecated: Use WriteErrorCodeWithDetails instead. Frontend handles localization.
func WriteLocalizedErrorWithDetails(w http.ResponseWriter, status int, code string, details any, cfg langconfig.LanguageConfig) {
	msg := cfg.UserMessages.Errors[code]
	if msg == "" {
		msg = cfg.UserMessages.ErrorMessage
	}
	WriteErrorWithDetails(w, status, msg, code, details)
}

// ErrorMapping maps service errors to HTTP responses
type ErrorMapping struct {
	StatusCode int
	ErrorCode  string
}

var handoffErrorMappings = []struct {
	Target  error
	Mapping ErrorMapping
}{
	{Target: services.ErrHandoffExists, Mapping: ErrorMapping{StatusCode: http.StatusConflict, ErrorCode: ErrHandoffExists}},
	{Target: services.ErrHandoffNotFound, Mapping: ErrorMapping{StatusCode: http.StatusNotFound, ErrorCode: ErrHandoffNotFound}},
	{Target: services.ErrHandoffExpired, Mapping: ErrorMapping{StatusCode: http.StatusGone, ErrorCode: ErrHandoffExpired}},
	{Target: services.ErrHandoffClosed, Mapping: ErrorMapping{StatusCode: http.StatusConflict, ErrorCode: ErrHandoffClosed}},
	{Target: services.ErrHandoffRateLimited, Mapping: ErrorMapping{StatusCode: http.StatusTooManyRequests, ErrorCode: ErrHandoffRateLimited}},
	{Target: services.ErrHandoffNotEnabled, Mapping: ErrorMapping{StatusCode: http.StatusBadRequest, ErrorCode: ErrHandoffNotEnabled}},
}

// MapHandoffError maps a service error to HTTP status and error code
func MapHandoffError(err error) (int, string, bool) {
	for _, m := range handoffErrorMappings {
		if errors.Is(err, m.Target) {
			return m.Mapping.StatusCode, m.Mapping.ErrorCode, true
		}
	}
	return http.StatusInternalServerError, ErrCodeInternalError, false
}

