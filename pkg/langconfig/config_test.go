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
}
