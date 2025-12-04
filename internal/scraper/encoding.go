package scraper

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
)

func NormalizeText(rawHTML string) (string, error) {
	r, err := charset.NewReader(strings.NewReader(rawHTML), "text/html")
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		b = b[3:]
	}
	if !utf8.Valid(b) {
		b = bytes.ToValidUTF8(b, []byte(" "))
	}
	s := string(b)
	s = strings.ReplaceAll(s, "\uFFFD", " ")
	s = strings.TrimSpace(s)
	return s, nil
}

func IsValidUTF8(data []byte) bool {
	return utf8.Valid(data)
}
