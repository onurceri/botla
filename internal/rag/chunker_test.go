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

func tailString(s string, n int) string {
    if n <= 0 || s == "" {
        return ""
    }
    if len(s) <= n {
        return s
    }
    return s[len(s)-n:]
}

func min(a, b int) int { if a < b { return a }; return b }
