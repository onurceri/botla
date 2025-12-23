package handlers

import (
	"strings"
)

// --- Helper functions ---

// normalizeSuggestions deduplicates and truncates suggestions
func normalizeSuggestions(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, q := range in {
		t := strings.TrimSpace(q)
		if t == "" {
			continue
		}
		if len(t) > 120 {
			t = t[:120]
		}
		k := strings.ToLower(t)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, t)
		if len(out) >= 6 {
			break
		}
	}
	return out
}

func defaultString(p *string, d string) string {
	if p != nil {
		s := strings.TrimSpace(*p)
		if s != "" {
			return s
		}
	}
	return d
}

func defaultInt(p *int, d int) int {
	if p != nil {
		return *p
	}
	return d
}

func defaultFloat32(p *float32, d float32) float32 {
	if p != nil {
		return *p
	}
	return d
}

func defaultFloat64(p *float64, d float64) float64 {
	if p != nil {
		return *p
	}
	return d
}

func boolValue(p *bool, d bool) bool {
	if p != nil {
		return *p
	}
	return d
}

func suggestionsValue(p *[]string) []string {
	if p != nil {
		return normalizeSuggestions(*p)
	}
	return nil
}

func pathsValue(p *[]string) []string {
	if p != nil {
		return normalizePaths(*p)
	}
	return nil
}

func selectorsValue(p *[]string) []string {
	if p != nil {
		return normalizeSelectors(*p)
	}
	return nil
}

func normalizePaths(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, path := range in {
		t := strings.TrimSpace(path)
		if t == "" {
			continue
		}
		// Ensure path starts with /
		if !strings.HasPrefix(t, "/") {
			t = "/" + t
		}
		k := strings.ToLower(t)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, t)
	}
	return out
}

// normalizeSelectors cleans and deduplicates CSS selectors
func normalizeSelectors(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, sel := range in {
		t := strings.TrimSpace(sel)
		if t == "" {
			continue
		}
		// Normalize internal whitespace
		t = strings.Join(strings.Fields(t), " ")
		k := strings.ToLower(t)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, t)
	}
	return out
}

func normalizeLocale(code string) string {
	if code == "" {
		return "tr-TR"
	}
	s := strings.TrimSpace(code)
	switch s {
	case "tr":
		return "tr-TR"
	case "en":
		return "en-US"
	}
	return s
}

func baseLangCode(code string) string {
	s := strings.TrimSpace(code)
	if s == "" {
		return "tr"
	}
	if i := strings.Index(s, "-"); i > 0 {
		s = s[:i]
	}
	return s
}
