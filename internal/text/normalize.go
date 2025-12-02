package text

import (
	"strings"
	"unicode"
)

func NormalizeTR(s string) string {
	if s == "" {
		return ""
	}
	r := strings.ReplaceAll(s, "\u00A0", " ")
	r = strings.ReplaceAll(r, "\u200B", " ")
	r = strings.ReplaceAll(r, "\t", " ")
	r = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' {
			return -1
		}
		return r
	}, r)
	r = strings.TrimSpace(r)
	r = strings.Join(strings.Fields(r), " ")
	return r
}
