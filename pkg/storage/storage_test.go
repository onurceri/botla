package storage

import (
	"strings"
	"testing"
)

func TestGenerateKey_Format(t *testing.T) {
	k := GenerateKey("p", "f.txt")
	if !strings.HasPrefix(k, "p/") || !strings.HasSuffix(k, "_f.txt") {
		t.Fatalf("bad key: %s", k)
	}
}
