package pdf

import "testing"

func TestExtractPDFText_NoFile(t *testing.T) {
    if _, err := ExtractPDFText("/tmp/does-not-exist.pdf"); err == nil {
        t.Fatalf("expected error for missing file")
    }
}

