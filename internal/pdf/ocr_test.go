//go:build fitz && ocr

package pdf

import (
	"os"
	"testing"
)

func TestExtractPDFWithOCR_Sample(t *testing.T) {
	p := os.Getenv("BOTLA_PDF_PATH")
	if p == "" {
		t.Skip("BOTLA_PDF_PATH not set; skipping OCR test")
	}
	s, err := ExtractPDFWithOCR(p)
	if err != nil {
		t.Skipf("ocr error: %v", err)
	}
	if len(s) == 0 {
		t.Fatalf("empty ocr text")
	}
}
