//go:build fitz && !ocr

package pdf

func ExtractPDFWithOCRCompat(filePath string, langCode string) (string, error) {
	return ExtractPDFWithOCR(filePath)
}
