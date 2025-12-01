//go:build fitz && ocr

package pdf

import (
    "bytes"
    "strings"

    "github.com/gen2brain/go-fitz"
    "github.com/otiai10/gosseract/v2"
    "github.com/onurceri/botla-co/internal/scraper"
)

func ExtractPDFWithOCR(filePath string) (string, error) {
    doc, err := fitz.New(filePath)
    if err != nil {
        return "", err
    }
    defer doc.Close()

    pages := doc.NumPage()
    if pages < 1 {
        return "", ErrNoSuchFile
    }

    c := gosseract.NewClient()
    defer c.Close()
    _ = c.SetLanguage("tur")
    _ = c.SetVariable("user_defined_dpi", "300")

    var out strings.Builder
    for n := 0; n < pages; n++ {
        png, perr := doc.ImagePNG(n, 300)
        if perr != nil {
            return "", perr
        }
        if err := c.SetImageFromBytes(png); err != nil {
            return "", err
        }
        txt, terr := c.Text()
        if terr != nil {
            return "", terr
        }
        if !scraper.IsValidUTF8([]byte(txt)) {
            txt = string(bytes.ToValidUTF8([]byte(txt), []byte("?")))
        }
        norm, nerr := scraper.NormalizeText(txt)
        if nerr != nil {
            norm = strings.TrimSpace(txt)
        }
        if norm != "" {
            out.WriteString(norm)
        }
        if n < pages-1 {
            out.WriteString("\n\n")
        }
    }
    return out.String(), nil
}
