package storage

import (
    "strings"
    "testing"
)

func TestGenerateKey(t *testing.T) {
    k := GenerateKey("sources", "a/b\\c.pdf")
    if !strings.HasPrefix(k, "sources/") {
        t.Fatalf("prefix missing in key: %q", k)
    }
    parts := strings.SplitN(k, "_", 2)
    if len(parts) != 2 {
        t.Fatalf("expected timestamp separator '_' in key: %q", k)
    }
    if !strings.HasSuffix(k, "c.pdf") {
        t.Fatalf("basename not applied: %q", k)
    }
}
