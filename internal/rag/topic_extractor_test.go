package rag

import (
	"context"
	"fmt"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
)

type mockLLM struct{ out string }

func (m mockLLM) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	return &models.CompletionResult{Content: m.out}, nil
}

func (m mockLLM) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{}
}

type errLLM struct{}

func (e errLLM) CreateCompletion(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	return nil, fmt.Errorf("llm error")
}

func (e errLLM) GetModelInfo() models.ModelInfo {
	return models.ModelInfo{}
}

func TestExtractIngestionMetadata_JSONHappyPath(t *testing.T) {
	js := `{
        "capability_summary": "Provides info about products.",
        "suggested_questions": ["What products do you offer?", "How can I purchase?", "Do you ship internationally?"]
    }`
	im, err := ExtractIngestionMetadata(context.Background(), mockLLM{out: js}, "demo content", "en", 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if im.CapabilitySummary == "" {
		t.Fatalf("summary empty")
	}
	if len(im.SuggestedQuestions) != 3 {
		t.Fatalf("expected 3 questions")
	}
}

func TestExtractIngestionMetadata_FallbackFromFence(t *testing.T) {
	fenced := "```json\n{\n  \"capability_summary\": \"Bilgi\",\n  \"suggested_questions\": [\"\", \"  \", \"Kısa bir genel bakış verir misin?\"]\n}\n```"
	im, err := ExtractIngestionMetadata(context.Background(), mockLLM{out: fenced}, "demo", "tr", 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if im.CapabilitySummary == "" {
		t.Fatalf("summary empty")
	}
	if len(im.SuggestedQuestions) == 0 {
		t.Fatalf("expected normalized suggestions")
	}
}

func TestExtractIngestionMetadata_FallbackError(t *testing.T) {
	// invalid JSON triggers fallback, errLLM returns error -> overall error
	_, err := ExtractIngestionMetadata(context.Background(), errLLM{}, "demo", "en", 0)
	if err == nil {
		t.Fatalf("expected error when fallback summary fails")
	}
}

func TestNormalizeSuggestions_DedupeAndCap(t *testing.T) {
	in := []string{"A", "a ", "", "This is a very long question that should be trimmed to the maximum allowed characters to prevent UI overflow and performance issues."}
	out := normalizeSuggestions(in)
	if len(out) < 2 {
		t.Fatalf("expected at least 2 unique items")
	}
	for _, s := range out {
		if len(s) > 120 {
			t.Fatalf("question too long")
		}
	}
}

func TestExtractIngestionMetadata_EmptyQuestionsInJSON(t *testing.T) {
	// Valid JSON but empty questions -> should derive from summary
	js := `{
        "capability_summary": "This text describes the refund policy.",
        "suggested_questions": []
    }`
	im, err := ExtractIngestionMetadata(context.Background(), mockLLM{out: js}, "demo content", "en", 0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if im.CapabilitySummary == "" {
		t.Fatalf("summary empty")
	}
	if len(im.SuggestedQuestions) == 0 {
		t.Fatalf("expected derived questions when JSON list is empty")
	}
}

func TestExtractIngestionMetadata_NilClient(t *testing.T) {
	_, err := ExtractIngestionMetadata(context.Background(), nil, "demo content", "en", 0)
	if err == nil {
		t.Fatalf("expected error for nil client")
	}
	if err.Error() != "LLM client is nil" {
		t.Fatalf("unexpected error: %v", err)
	}
}
