package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/pkg/langconfig"
)

func TestWriteLocalizedError_TR(t *testing.T) {
	rr := httptest.NewRecorder()
	cfg := langconfig.Get("tr")
	WriteLocalizedError(rr, http.StatusPaymentRequired, ErrMonthlyTokensExceeded, cfg)
	if rr.Code != http.StatusPaymentRequired {
		t.Fatalf("status code mismatch: %d", rr.Code)
	}
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if resp.Error != cfg.UserMessages.Errors[ErrMonthlyTokensExceeded] {
		t.Fatalf("unexpected error text: %q", resp.Error)
	}
	if resp.Code != ErrMonthlyTokensExceeded {
		t.Fatalf("unexpected code: %q", resp.Code)
	}
}

func TestWriteLocalizedError_EN(t *testing.T) {
	rr := httptest.NewRecorder()
	cfg := langconfig.Get("en")
	WriteLocalizedError(rr, http.StatusBadRequest, ErrInvalidRequestBody, cfg)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status code mismatch: %d", rr.Code)
	}
	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json error: %v", err)
	}
	if resp.Error != cfg.UserMessages.Errors[ErrInvalidRequestBody] {
		t.Fatalf("unexpected error text: %q", resp.Error)
	}
	if resp.Code != ErrInvalidRequestBody {
		t.Fatalf("unexpected code: %q", resp.Code)
	}
}
