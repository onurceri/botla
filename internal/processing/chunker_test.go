package processing

import "testing"

func TestChunkText_Basic(t *testing.T) {
    text := "Paragraf 1.\n\nParagraf 2 uzun metin burada devam eder. Kelimeler ve cümleler.\n\nParagraf 3 daha kısa."
    chunks := ChunkText(text, 50, 10)
    if len(chunks) == 0 {
        t.Fatalf("no chunks produced")
    }
    for _, c := range chunks {
        if len(c) == 0 {
            t.Fatalf("empty chunk")
        }
    }
}

