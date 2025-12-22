package rag

import (
	"context"
	"testing"
)

func TestExtractTopics_UsesLLM(t *testing.T) {
	m := mockLLM{out: " summary "}
	s, err := extractTopicsFallback(context.Background(), m, "content", "en")
	if err != nil || s == "" {
		t.Fatalf("extract topics err: %v", err)
	}
}

func TestDeriveQuestionsFromSummary_EN(t *testing.T) {
	qs := deriveQuestionsFromSummary("hello", "en")
	if len(qs) == 0 {
		t.Fatalf("no qs")
	}
}

func TestDeriveQuestionsFromSummary_TR(t *testing.T) {
	qs := deriveQuestionsFromSummary("merhaba", "tr")
	if len(qs) == 0 {
		t.Fatalf("no qs")
	}
}
