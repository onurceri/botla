package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/langconfig"
)

func TestWriteErrorCodeWithDetails(t *testing.T) {
	t.Run("writes_error_with_details", func(t *testing.T) {
		w := httptest.NewRecorder()
		details := map[string]string{"field": "value"}

		WriteErrorCodeWithDetails(w, http.StatusBadRequest, "ERR_VALIDATION", details)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}

		var resp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Code != "ERR_VALIDATION" {
			t.Errorf("expected code ERR_VALIDATION, got %s", resp.Code)
		}
		if resp.Status != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.Status)
		}
		if resp.Details == nil {
			t.Error("expected details to be set")
		}
	})

	t.Run("writes_error_with_array_details", func(t *testing.T) {
		w := httptest.NewRecorder()
		details := []string{"error1", "error2"}

		WriteErrorCodeWithDetails(w, http.StatusUnprocessableEntity, "ERR_INVALID", details)

		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
		}

		var resp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Code != "ERR_INVALID" {
			t.Errorf("expected code ERR_INVALID, got %s", resp.Code)
		}
	})
}

func TestWriteErrorWithDetails(t *testing.T) {
	t.Run("writes_deprecated_error_with_details", func(t *testing.T) {
		w := httptest.NewRecorder()
		details := map[string]any{"validation_errors": []string{"email required"}}

		WriteErrorWithDetails(w, http.StatusBadRequest, "validation failed", "ERR_VALIDATION", details)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		var resp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Error != "validation failed" {
			t.Errorf("expected error 'validation failed', got %s", resp.Error)
		}
		if resp.Code != "ERR_VALIDATION" {
			t.Errorf("expected code ERR_VALIDATION, got %s", resp.Code)
		}
		if resp.Details == nil {
			t.Error("expected details to be set")
		}
	})
}

func TestWriteLocalizedErrorWithDetails(t *testing.T) {
	t.Run("writes_localized_error_with_details", func(t *testing.T) {
		w := httptest.NewRecorder()
		cfg := langconfig.LanguageConfig{
			UserMessages: langconfig.UserMessages{
				ErrorMessage: "An error occurred",
				Errors: map[string]string{
					"ERR_TEST": "This is a test error",
				},
			},
		}
		details := map[string]string{"info": "additional"}

		WriteLocalizedErrorWithDetails(w, http.StatusForbidden, "ERR_TEST", details, cfg)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}

		var resp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Error != "This is a test error" {
			t.Errorf("expected localized error message, got %s", resp.Error)
		}
		if resp.Code != "ERR_TEST" {
			t.Errorf("expected code ERR_TEST, got %s", resp.Code)
		}
	})

	t.Run("falls_back_to_default_message_when_code_not_found", func(t *testing.T) {
		w := httptest.NewRecorder()
		cfg := langconfig.LanguageConfig{
			UserMessages: langconfig.UserMessages{
				ErrorMessage: "Default error message",
				Errors:       map[string]string{},
			},
		}

		WriteLocalizedErrorWithDetails(w, http.StatusInternalServerError, "ERR_UNKNOWN", nil, cfg)

		var resp ErrorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if resp.Error != "Default error message" {
			t.Errorf("expected default error message, got %s", resp.Error)
		}
	})
}

func TestMapHandoffError_AllMappings(t *testing.T) {
	testCases := []struct {
		name           string
		err            error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "maps_Exists_error",
			err:            services.ErrHandoffExists,
			expectedCode:   ErrHandoffExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "maps_NotFound_error",
			err:            services.ErrHandoffNotFound,
			expectedCode:   ErrHandoffNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "maps_Expired_error",
			err:            services.ErrHandoffExpired,
			expectedCode:   ErrHandoffExpired,
			expectedStatus: http.StatusGone,
		},
		{
			name:           "maps_Closed_error",
			err:            services.ErrHandoffClosed,
			expectedCode:   ErrHandoffClosed,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "maps_RateLimited_error",
			err:            services.ErrHandoffRateLimited,
			expectedCode:   ErrHandoffRateLimited,
			expectedStatus: http.StatusTooManyRequests,
		},
		{
			name:           "maps_NotEnabled_error",
			err:            services.ErrHandoffNotEnabled,
			expectedCode:   ErrHandoffNotEnabled,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, code, found := MapHandoffError(tc.err)
			if !found {
				t.Error("expected to find error mapping")
			}
			if status != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, status)
			}
			if code != tc.expectedCode {
				t.Errorf("expected code %s, got %s", tc.expectedCode, code)
			}
		})
	}

	t.Run("returns_internal_error_for_unknown_error", func(t *testing.T) {
		status, code, found := MapHandoffError(errors.New("unknown error"))
		if found {
			t.Error("expected not to find error mapping")
		}
		if status != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, status)
		}
		if code != ErrCodeInternalError {
			t.Errorf("expected code %s, got %s", ErrCodeInternalError, code)
		}
	})

	t.Run("returns_internal_error_for_nil", func(t *testing.T) {
		status, code, found := MapHandoffError(nil)
		if found {
			t.Error("expected not to find error mapping")
		}
		if status != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, status)
		}
		if code != ErrCodeInternalError {
			t.Errorf("expected code %s, got %s", ErrCodeInternalError, code)
		}
	})
}
