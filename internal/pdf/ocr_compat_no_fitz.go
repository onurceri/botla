//go:build !fitz

package pdf

import "fmt"

func ExtractPDFWithOCRCompat(filePath string, langCode string) (string, error) {
	return "", fmt.Errorf("pdf: ocr unavailable (build with '-tags fitz' to enable)")
}
