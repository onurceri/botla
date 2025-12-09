package rag

import (
	"strings"
	"testing"
)

func TestChunkText_TurkishBasics(t *testing.T) {
	text := "Dr. Ahmet bugün toplantıya katıldı. Önemli kararlar alındı.\n\nProf. Ayşe yarın sunum yapacak. Detaylar e-posta ile gönderildi, vb. bilgilendirmeler yapıldı."
	chunks, err := ChunkText(text, 50, "tr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatalf("no chunks produced")
	}
	for i, c := range chunks {
		if c.Text == "" || c.TokenCount <= 0 {
			t.Fatalf("invalid chunk at %d", i)
		}
	}
	// ensure overlap between consecutive chunks exists when more than one
	if len(chunks) > 1 {
		prev := chunks[0].Text
		next := chunks[1].Text
		tail := tailString(prev, 20)
		if tail != "" && !strings.Contains(next, tail[:min(len(tail), 10)]) {
			t.Fatalf("expected overlap between chunks")
		}
	}
}

func TestChunkText_Abbreviations(t *testing.T) {
	// Logic migrated from cmd/test_chunker/main.go
	text := "Prof. Dr. Ahmet Bey geldi. Yanında Av. Mehmet de vardı. Bu bir test cümlesidir. Kısaltmalar vb. doğru çalışmalı."

	// We use a small chunkSize to force potential splits if logic is wrong
	chunks, err := ChunkText(text, 50, "tr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatalf("no chunks produced")
	}

	// Check that we didn't split in the middle of a sentence just because of a dot in abbreviation
	// The first chunk should ideally contain the first sentence fully if it fits, or at least not split at "Dr."
	firstChunk := chunks[0].Text

	// "Prof. Dr. Ahmet Bey geldi." is ~26 chars.
	// "Yanında Av. Mehmet de vardı." is ~28 chars.
	// Total ~55 chars. Chunk size 50.
	// So it MIGHT split between sentences, but it MUST NOT split at "Dr." or "Av."

	if strings.Contains(firstChunk, "Prof.") && !strings.Contains(firstChunk, "Dr.") {
		t.Errorf("Split occurred between Prof. and Dr.")
	}
	if strings.Contains(firstChunk, "Av.") && strings.HasSuffix(firstChunk, "Av.") {
		t.Errorf("Split occurred right after Av.")
	}

	// Also verify that the full text is preserved across chunks

	// This simple reconstruction assumes no overlap for verification,
	// but ChunkText DOES produce overlap.
	// So we just check if the critical phrases exist intact in SOME chunk.

	foundDr := false
	foundAv := false

	for _, c := range chunks {
		if strings.Contains(c.Text, "Prof. Dr. Ahmet") {
			foundDr = true
		}
		if strings.Contains(c.Text, "Av. Mehmet") {
			foundAv = true
		}
	}

	if !foundDr {
		t.Errorf("Did not find 'Prof. Dr. Ahmet' intact in any chunk")
	}
	if !foundAv {
		t.Errorf("Did not find 'Av. Mehmet' intact in any chunk")
	}

}

func tailString(s string, n int) string {
	if n <= 0 || s == "" {
		return ""
	}
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestChunkText_EmptyAndInvalidTarget(t *testing.T) {
	ch, err := ChunkText("", 50, "tr")
	if err != nil || ch != nil {
		t.Fatalf("empty should yield nil, got %v", ch)
	}
	_, err = ChunkText("hello", 0, "tr")
	if err == nil {
		t.Fatalf("expected error for invalid targetTokens")
	}
}

func TestChunkText_TurkishSentenceEndings(t *testing.T) {
	// TRK-022, TRK-023, TRK-024
	cases := []struct {
		name     string
		input    string
		expected int // expected minimum chunks (sentences)
	}{
		{"Dot", "Bu birinci cümle. Bu ikinci cümle.", 2},
		{"Question", "Nasılsınız? İyiyim.", 2},
		{"Exclamation", "Dikkat et! Tehlike var.", 2},
		{"Mixed", "Geliyor musun bu akşam? Evet, kesinlikle geleceğim. Harika, o zaman görüşürüz!", 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Use a very small token limit to force splitting if sentences are detected
			chunks, err := ChunkText(tc.input, 5, "tr")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(chunks) < tc.expected {
				t.Errorf("Expected at least %d chunks for input %q, got %d", tc.expected, tc.input, len(chunks))
			}
		})
	}
}

func TestChunkText_MixedLanguage(t *testing.T) {
	// TRK-025
	text := "This is an English sentence. Bu bir Türkçe cümle. Another English one."
	chunks, err := ChunkText(text, 5, "tr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) < 3 {
		t.Errorf("Expected 3 chunks for mixed language, got %d", len(chunks))
	}
}

func TestChunkText_TurkishCharacters(t *testing.T) {
	// TRK-001
	specialChars := "şŞğĞıİöÖüÜçÇ"
	text := "Bu cümlede özel karakterler var: " + specialChars
	chunks, err := ChunkText(text, 100, "tr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatalf("no chunks produced")
	}
	if !strings.Contains(chunks[0].Text, specialChars) {
		t.Errorf("Special characters not preserved. Expected to contain %q, got %q", specialChars, chunks[0].Text)
	}
}
