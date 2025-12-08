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
