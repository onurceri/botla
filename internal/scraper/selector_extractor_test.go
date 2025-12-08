package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractBySelectors(t *testing.T) {
	html := `
	<html>
	<body>
		<nav>Menu items</nav>
		<main>
			<article class="content">
				<p>Important content here</p>
			</article>
		</main>
		<aside class="sidebar">Sidebar content</aside>
		<footer>Footer info</footer>
	</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	testCases := []struct {
		name        string
		selectors   []string
		contains    []string
		notContains []string
	}{
		{
			name:        "single selector - class",
			selectors:   []string{".content"},
			contains:    []string{"Important content here"},
			notContains: []string{"Menu items", "Footer info", "Sidebar content"},
		},
		{
			name:        "single selector - tag",
			selectors:   []string{"main"},
			contains:    []string{"Important content here"},
			notContains: []string{"Menu items", "Footer info", "Sidebar content"},
		},
		{
			name:        "multiple selectors",
			selectors:   []string{"main", "footer"},
			contains:    []string{"Important content here", "Footer info"},
			notContains: []string{"Menu items", "Sidebar content"},
		},
		{
			name:        "empty selectors - fallback to body",
			selectors:   []string{},
			contains:    []string{"Menu items", "Important content here", "Footer info"},
			notContains: []string{},
		},
		{
			name:        "non-matching selector - fallback to body",
			selectors:   []string{".non-existent"},
			contains:    []string{"Menu items", "Important content here", "Footer info"},
			notContains: []string{},
		},
		{
			name:        "nested selector",
			selectors:   []string{"main article p"},
			contains:    []string{"Important content here"},
			notContains: []string{"Menu items", "Footer info", "Sidebar content"},
		},
		{
			name:        "ID selector",
			selectors:   []string{".sidebar"},
			contains:    []string{"Sidebar content"},
			notContains: []string{"Menu items", "Footer info", "Important content here"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractBySelectors(doc.Find("body"), tc.selectors)

			for _, want := range tc.contains {
				if !strings.Contains(result, want) {
					t.Errorf("Expected result to contain %q, got: %s", want, result)
				}
			}

			for _, notWant := range tc.notContains {
				if strings.Contains(result, notWant) {
					t.Errorf("Expected result NOT to contain %q, got: %s", notWant, result)
				}
			}
		})
	}
}

func TestExtractBySelectors_Deduplication(t *testing.T) {
	html := `
	<html>
	<body>
		<div class="content">Same content</div>
		<div class="content">Same content</div>
		<div class="content">Different content</div>
	</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := ExtractBySelectors(doc.Find("body"), []string{".content"})

	// Should contain both unique contents
	if !strings.Contains(result, "Same content") {
		t.Error("Expected 'Same content' in result")
	}
	if !strings.Contains(result, "Different content") {
		t.Error("Expected 'Different content' in result")
	}

	// Duplicate should only appear once
	count := strings.Count(result, "Same content")
	if count != 1 {
		t.Errorf("Expected 'Same content' to appear once, got %d", count)
	}
}

func TestValidateSelector(t *testing.T) {
	testCases := []struct {
		name      string
		selector  string
		expectErr bool
	}{
		{"valid class selector", ".content", false},
		{"valid id selector", "#main", false},
		{"valid tag selector", "article", false},
		{"valid nested selector", "main .content p", false},
		{"valid attribute selector", "[data-id='test']", false},
		{"valid complex selector", "main > article.content > p:first-child", false},

		{"empty selector", "", true},
		{"whitespace only", "   ", true},
		{"too long selector", strings.Repeat("a", 201), true},
		{"dangerous script tag", "script", true},
		{"dangerous style tag", "style", true},
		{"dangerous noscript tag", "noscript", true},
		{"dangerous iframe tag", "iframe", true},
		{"universal selector alone", "*", true},
		{"contains pattern", ":contains(text)", true},
		{"javascript pattern", "javascript:alert(1)", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSelector(tc.selector)
			if tc.expectErr && err == nil {
				t.Errorf("Expected error for selector %q, got nil", tc.selector)
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error for selector %q: %v", tc.selector, err)
			}
		})
	}
}

func TestValidateSelectors(t *testing.T) {
	// Mix of valid and invalid
	selectors := []string{".valid", "script", "#ok", ""}
	errors := ValidateSelectors(selectors)

	// Should have 2 errors (script and empty)
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
}

func TestNormalizeSelectors(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "remove duplicates (case insensitive)",
			input:    []string{".Content", ".content", ".CONTENT"},
			expected: []string{".Content"},
		},
		{
			name:     "trim whitespace",
			input:    []string{"  .content  ", "\t\n.sidebar\t"},
			expected: []string{".content", ".sidebar"},
		},
		{
			name:     "remove empty strings",
			input:    []string{"", ".content", "   ", ".sidebar"},
			expected: []string{".content", ".sidebar"},
		},
		{
			name:     "normalize internal whitespace",
			input:    []string{"main   .content    p"},
			expected: []string{"main .content p"},
		},
		{
			name:     "preserve order",
			input:    []string{".first", ".second", ".third"},
			expected: []string{".first", ".second", ".third"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeSelectors(tc.input)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d selectors, got %d: %v", len(tc.expected), len(result), result)
				return
			}

			for i, want := range tc.expected {
				if result[i] != want {
					t.Errorf("At index %d: expected %q, got %q", i, want, result[i])
				}
			}
		})
	}
}

func TestExtractBySelectors_ScriptsRemoved(t *testing.T) {
	html := `
	<html>
	<body>
		<main>
			<script>var evil = 'code';</script>
			<p>Good content</p>
			<style>.hidden{display:none}</style>
		</main>
	</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Failed to parse HTML: %v", err)
	}

	result := ExtractBySelectors(doc.Find("body"), []string{"main"})

	if !strings.Contains(result, "Good content") {
		t.Error("Expected 'Good content' in result")
	}
	if strings.Contains(result, "evil") {
		t.Error("Script content should be removed")
	}
	if strings.Contains(result, ".hidden") {
		t.Error("Style content should be removed")
	}
}
