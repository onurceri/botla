package processing

import (
	"testing"
)

func TestAggregateWithWeightedSelection_EmptyInputs(t *testing.T) {
	result := AggregateWithWeightedSelection(nil, nil, 6)
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d items", len(result))
	}
}

func TestAggregateWithWeightedSelection_ExistingQuestionsFirst(t *testing.T) {
	existing := []string{"Existing Q1", "Existing Q2"}
	sources := []SourceQuestions{
		{Questions: []string{"Source Q1", "Source Q2"}, ChunkCount: 100},
	}
	result := AggregateWithWeightedSelection(sources, existing, 6)

	if len(result) != 4 {
		t.Fatalf("expected 4 questions, got %d", len(result))
	}
	// Existing questions should come first
	if result[0] != "Existing Q1" || result[1] != "Existing Q2" {
		t.Fatalf("existing questions should be first, got: %v", result)
	}
}

func TestAggregateWithWeightedSelection_LargerSourceFirst(t *testing.T) {
	sources := []SourceQuestions{
		{Questions: []string{"Small Q"}, ChunkCount: 10},
		{Questions: []string{"Large Q"}, ChunkCount: 100},
	}
	result := AggregateWithWeightedSelection(sources, nil, 6)

	if len(result) != 2 {
		t.Fatalf("expected 2 questions, got %d", len(result))
	}
	// Larger source question should come first
	if result[0] != "Large Q" {
		t.Fatalf("expected Large Q first, got: %s", result[0])
	}
}

func TestAggregateWithWeightedSelection_RespectsLimit(t *testing.T) {
	sources := []SourceQuestions{
		{Questions: []string{"Q1", "Q2", "Q3", "Q4", "Q5"}, ChunkCount: 100},
		{Questions: []string{"Q6", "Q7", "Q8"}, ChunkCount: 50},
	}
	result := AggregateWithWeightedSelection(sources, nil, 3)

	if len(result) != 3 {
		t.Fatalf("expected 3 questions (limit), got %d", len(result))
	}
}

func TestAggregateWithWeightedSelection_Deduplication(t *testing.T) {
	sources := []SourceQuestions{
		{Questions: []string{"Hello?", "hello?"}, ChunkCount: 100}, // duplicates (case-insensitive)
		{Questions: []string{"World?"}, ChunkCount: 50},
	}
	result := AggregateWithWeightedSelection(sources, nil, 6)

	if len(result) != 2 {
		t.Fatalf("expected 2 unique questions, got %d: %v", len(result), result)
	}
}

func TestAggregateWithWeightedSelection_ExistingAtLimit(t *testing.T) {
	existing := []string{"Q1", "Q2", "Q3"}
	sources := []SourceQuestions{
		{Questions: []string{"Q4", "Q5"}, ChunkCount: 100},
	}
	result := AggregateWithWeightedSelection(sources, existing, 3)

	if len(result) != 3 {
		t.Fatalf("expected 3 questions (limit), got %d", len(result))
	}
	// Should only have existing questions
	if result[0] != "Q1" || result[1] != "Q2" || result[2] != "Q3" {
		t.Fatalf("expected only existing questions, got: %v", result)
	}
}

func TestAggregateWithWeightedSelection_LongQuestionTruncated(t *testing.T) {
	longQ := "This is a very long question that exceeds the maximum allowed characters to prevent UI overflow and performance issues. It should be truncated to 120 characters."
	sources := []SourceQuestions{
		{Questions: []string{longQ}, ChunkCount: 100},
	}
	result := AggregateWithWeightedSelection(sources, nil, 6)

	if len(result) != 1 {
		t.Fatalf("expected 1 question, got %d", len(result))
	}
	if len(result[0]) != 120 {
		t.Fatalf("expected question length 120, got %d", len(result[0]))
	}
}

func TestAggregateWithWeightedSelection_EmptyQuestionsSkipped(t *testing.T) {
	sources := []SourceQuestions{
		{Questions: []string{"", "  ", "Valid?"}, ChunkCount: 100},
	}
	result := AggregateWithWeightedSelection(sources, nil, 6)

	if len(result) != 1 {
		t.Fatalf("expected 1 valid question, got %d: %v", len(result), result)
	}
	if result[0] != "Valid?" {
		t.Fatalf("expected 'Valid?', got: %s", result[0])
	}
}

func TestNormalizeQuestion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"", ""},
		{"   ", ""},
		{"short", "short"},
	}

	for _, tc := range tests {
		got := normalizeQuestion(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeQuestion(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestSlicesEqual(t *testing.T) {
	tests := []struct {
		a, b     []string
		expected bool
	}{
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"a", "b"}, []string{"b", "a"}, false},
		{[]string{"a"}, []string{"a", "b"}, false},
		{nil, nil, true},
		{[]string{}, []string{}, true},
	}

	for _, tc := range tests {
		got := slicesEqual(tc.a, tc.b)
		if got != tc.expected {
			t.Errorf("slicesEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.expected)
		}
	}
}
