//go:build !fitz

package pdf

import "fmt"

func ExtractPDFText(filePath string) (string, error) {
	return "", fmt.Errorf("pdf: extractor unavailable (build with '-tags fitz' and install MuPDF)")
}
