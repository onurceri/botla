//go:build fitz

package pdf

import (
	"os"
	"testing"
)

func TestExtractPDFText_Sample(t *testing.T) {
	p := os.Getenv("BOTLA_PDF_PATH")
	if p == "" {
		t.Skip("BOTLA_PDF_PATH not set; skipping fitz extraction test")
	}
	s, err := ExtractPDFText(p, "tr")
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if len(s) == 0 {
		t.Fatalf("empty extracted text")
	}
}
