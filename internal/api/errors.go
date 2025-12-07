package api

import (
	"encoding/json"
	"net/http"
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
