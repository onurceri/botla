package pdf

import (
	"errors"
	"testing"
)

func TestPDFError(t *testing.T) {
	inner := errors.New("file not found")
	err := &PDFError{Op: "open", Err: inner}

	if err.Error() != "pdf open: file not found" {
		t.Errorf("unexpected error string: %s", err.Error())
	}

	if !errors.Is(err, inner) {
		t.Error("expected error to wrap inner")
	}
}
