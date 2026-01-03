//go:build !fitz

package pdf

import "fmt"

func ExtractPDFText(_ string, _ string, _ bool) (string, error) {
	return "", fmt.Errorf("pdf support not enabled (build with -tags fitz)")
}
