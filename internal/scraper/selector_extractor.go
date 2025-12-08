package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ScrapeConfig holds configuration options for scraping
type ScrapeConfig struct {
	// Selectors is a list of CSS selectors to extract content from
	// If empty, the entire body is scraped (default behavior)
	Selectors []string
}

// ExtractBySelectors extracts visible text from elements matching the given CSS selectors.
// If no content is found from selectors, it falls back to the entire body.
// Each selector's content is extracted and deduplicated.
func ExtractBySelectors(doc *goquery.Selection, selectors []string) string {
	if len(selectors) == 0 {
		return visibleText(doc)
	}

	var parts []string
	seen := make(map[string]bool)

	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}

		doc.Find(sel).Each(func(_ int, s *goquery.Selection) {
			txt := visibleText(s.Clone())
			txt = strings.TrimSpace(txt)
			if txt != "" && !seen[txt] {
				seen[txt] = true
				parts = append(parts, txt)
			}
		})
	}

	// Fallback to full body if no content found
	if len(parts) == 0 {
		return visibleText(doc)
	}

	return strings.Join(parts, "\n\n")
}

// ValidateSelector checks if a CSS selector is valid and safe.
// Returns an error if the selector is invalid or contains dangerous patterns.
func ValidateSelector(selector string) error {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return fmt.Errorf("empty selector")
	}

	// Check length limit (prevent overly complex selectors)
	if len(selector) > 200 {
		return fmt.Errorf("selector too long (max 200 characters)")
	}

	// Validate basic CSS selector syntax using goquery
	// Parse a minimal HTML document to test the selector
	doc, err := goquery.NewDocumentFromReader(strings.NewReader("<html><body></body></html>"))
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}

	// goquery.Find will panic on invalid selectors, so we need to catch this
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("invalid CSS selector: %v", r)
		}
	}()

	_ = doc.Find(selector)
	if err != nil {
		return err
	}

	// Block potentially dangerous patterns (for security)
	dangerousPatterns := []string{
		":contains(",   // Non-standard, potentially problematic
		":has(",        // Some implementations may not support this safely
		"javascript:",  // XSS prevention
		"expression(",  // CSS expression (IE)
	}

	lowerSel := strings.ToLower(selector)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerSel, pattern) {
			return fmt.Errorf("selector contains disallowed pattern: %s", pattern)
		}
	}

	// Disallow selectors that are too generic and could cause issues
	genericPatterns := regexp.MustCompile(`^(\*|script|style|noscript|iframe|link|meta|head)$`)
	if genericPatterns.MatchString(strings.ToLower(selector)) {
		return fmt.Errorf("selector '%s' is too generic or targets disallowed elements", selector)
	}

	return nil
}

// ValidateSelectors validates multiple selectors and returns all errors
func ValidateSelectors(selectors []string) []error {
	var errors []error
	for _, sel := range selectors {
		if err := ValidateSelector(sel); err != nil {
			errors = append(errors, fmt.Errorf("selector '%s': %w", sel, err))
		}
	}
	return errors
}

// NormalizeSelectors cleans and deduplicates a list of selectors
func NormalizeSelectors(selectors []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, sel := range selectors {
		sel = strings.TrimSpace(sel)
		if sel == "" {
			continue
		}
		// Normalize whitespace
		sel = strings.Join(strings.Fields(sel), " ")
		lower := strings.ToLower(sel)
		if !seen[lower] {
			seen[lower] = true
			result = append(result, sel)
		}
	}

	return result
}
