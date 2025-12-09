package langconfig

import "testing"

func TestGet_DefaultsToTR(t *testing.T) {
	c := Get("")
	if c.Code != "tr" {
		t.Fatalf("expected default code 'tr', got %q", c.Code)
	}
}

func TestTemplates_ENandTR(t *testing.T) {
	en := Get("en").ResponseTemplates
	tr := Get("tr").ResponseTemplates
	if en.DefaultSystemPrompt == "" || tr.DefaultSystemPrompt == "" {
		t.Fatalf("default system prompts must not be empty")
	}
	if en.NoInfoFound == "" || tr.NoInfoFound == "" {
		t.Fatalf("NoInfoFound templates must not be empty")
	}
	if EN_TopicExtractionSystemPrompt == "" || TR_TopicExtractionSystemPrompt == "" {
		t.Fatalf("topic extraction system prompts must not be empty")
	}
	if EN_TopicExtractionUserPrompt == "" || TR_TopicExtractionUserPrompt == "" {
		t.Fatalf("topic extraction user prompts must not be empty")
	}
	if en.Errors["CHAT_TIMEOUT_OR_INCOMPLETE"] == "" || tr.Errors["CHAT_TIMEOUT_OR_INCOMPLETE"] == "" {
		t.Fatalf("chat timeout/incomplete messages must not be empty")
	}
	if en.Errors["ERR_INVALID_REQUEST_BODY"] == "" || tr.Errors["ERR_INVALID_REQUEST_BODY"] == "" {
		t.Fatalf("invalid request body messages must not be empty")
	}
}

func TestTurkishErrorMessages(t *testing.T) {
	// TRK-010 to TRK-015
	tr := Get("tr").ResponseTemplates

	expectedErrors := map[string]string{
		"ERR_MONTHLY_TOKENS_EXCEEDED": "Aylık token sınırı aşıldı",
		"ERR_DUPLICATE_URL":           "Yinelenen URL",
		"CHAT_TIMEOUT_OR_INCOMPLETE":  "İşlem tamamlanamadı veya çok uzun sürdü.",
	}

	for k, v := range expectedErrors {
		if tr.Errors[k] != v {
			t.Errorf("Expected error %q to be %q, got %q", k, v, tr.Errors[k])
		}
	}

	if tr.NoInfoFound != "Yeterli bilgi bulamadım." {
		t.Errorf("Expected NoInfoFound to be 'Yeterli bilgi bulamadım.', got %q", tr.NoInfoFound)
	}

	if tr.DefaultSystemPrompt != "Her zaman Türkçe yanıt ver ve sadece verilen bağlamı kullan." {
		t.Errorf("Expected DefaultSystemPrompt to be 'Her zaman Türkçe yanıt ver ve sadece verilen bağlamı kullan.', got %q", tr.DefaultSystemPrompt)
	}

	// TRK-013: Verify all 21 error codes have Turkish translations
	// This is a rough count based on the list in the markdown file
	expectedCount := 21
	// We might have more, but we should check if we have at least these many
	if len(tr.Errors) < expectedCount {
		t.Logf("Warning: Expected at least %d error messages, got %d", expectedCount, len(tr.Errors))
	}

	// Check for presence of key ones mentioned in the doc
	keysToCheck := []string{
		"ERR_MONTHLY_TOKENS_EXCEEDED",
		"ERR_NAME_AND_ACTION_TYPE_REQUIRED",
		"ERR_PDF_LIMIT_REACHED",
		"ERR_FILE_TOO_LARGE",
		"ERR_READD_COOLDOWN_ACTIVE",
		"ERR_DUPLICATE_URL",
		"ERR_ONLY_URL_REFRESH",
		"ERR_SOURCE_ALREADY_PROCESSING",
		"ERR_PLAN_REFRESH_UNAVAILABLE",
		"ERR_MONTHLY_REFRESH_EXCEEDED",
		"ERR_REFRESH_COOLDOWN_ACTIVE",
		"ERR_INVALID_REQUEST_BODY",
		"ERR_NO_URLS_PROVIDED",
		"ERR_URL_LIMIT_REACHED",
		"ERR_MONTHLY_INGESTION_EXCEEDED",
		"ERR_SITEMAP_PARSE_FAILED",
		"CHAT_TIMEOUT_OR_INCOMPLETE",
		"HANDOFF_NOT_ENABLED",
		"HANDOFF_CREATE_FAILED",
		"HANDOFF_EMAIL_NOT_CONFIGURED",
		"ERR_INVALID_STATUS",
	}

	for _, k := range keysToCheck {
		if _, ok := tr.Errors[k]; !ok {
			t.Errorf("Missing translation for error code: %s", k)
		}
	}
}
