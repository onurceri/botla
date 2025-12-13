package langconfig

import "testing"

func TestGet_DefaultsToTR(t *testing.T) {
	c := Get("")
	if c.Code != "tr" {
		t.Fatalf("expected default code 'tr', got %q", c.Code)
	}
}

func TestUserMessages_ENandTR(t *testing.T) {
	en := Get("en").UserMessages
	tr := Get("tr").UserMessages
	if en.NoInfoFound == "" || tr.NoInfoFound == "" {
		t.Fatalf("NoInfoFound messages must not be empty")
	}
	if en.ErrorMessage == "" || tr.ErrorMessage == "" {
		t.Fatalf("ErrorMessage must not be empty")
	}
	if en.WelcomeMessage == "" || tr.WelcomeMessage == "" {
		t.Fatalf("WelcomeMessage must not be empty")
	}
	if en.Errors["CHAT_TIMEOUT_OR_INCOMPLETE"] == "" || tr.Errors["CHAT_TIMEOUT_OR_INCOMPLETE"] == "" {
		t.Fatalf("chat timeout/incomplete messages must not be empty")
	}
	if en.Errors["ERR_INVALID_REQUEST_BODY"] == "" || tr.Errors["ERR_INVALID_REQUEST_BODY"] == "" {
		t.Fatalf("invalid request body messages must not be empty")
	}
}

func TestLanguageConfig_HasName(t *testing.T) {
	en := Get("en")
	tr := Get("tr")
	if en.Name != "English" {
		t.Errorf("Expected English Name to be 'English', got %q", en.Name)
	}
	if tr.Name != "Turkish" {
		t.Errorf("Expected Turkish Name to be 'Turkish', got %q", tr.Name)
	}
}

func TestTurkishErrorMessages(t *testing.T) {
	tr := Get("tr").UserMessages

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

	// TRK-013: Verify all key error codes have Turkish translations
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
