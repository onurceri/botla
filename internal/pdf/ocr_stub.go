//go:build !ocr

package pdf

import "fmt"

func ExtractPDFWithOCR(filePath string) (string, error) {
	return "", fmt.Errorf("pdf: ocr unavailable (build with '-tags ocr,fitz' and install tesseract)")
}
