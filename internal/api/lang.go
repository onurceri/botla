package api

import (
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/pkg/langconfig"
)

func BaseLang(code string) string {
	s := strings.TrimSpace(code)
	if s == "" {
		return "tr"
	}
	if i := strings.Index(s, "-"); i > 0 {
		s = s[:i]
	}
	return s
}

func LangFromRequest(r *http.Request) string {
	q := strings.TrimSpace(r.URL.Query().Get("lang"))
	if q != "" {
		return BaseLang(q)
	}
	al := r.Header.Get("Accept-Language")
	if al != "" {
		parts := strings.Split(al, ",")
		if len(parts) > 0 {
			return BaseLang(parts[0])
		}
	}
	return "tr"
}

func ConfigFromBase(base string) langconfig.LanguageConfig {
	return langconfig.Get(base)
}
