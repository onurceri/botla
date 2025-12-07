//go:build !fitz

package pdf

import "fmt"

func ExtractPDFText(filePath string, langCode string, allowOCR bool) (string, error) {
	return "", fmt.Errorf("pdf support not enabled (build with -tags fitz)")
}
