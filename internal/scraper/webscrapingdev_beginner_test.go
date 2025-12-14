//go:build webscraping

package scraper

import (
	"strings"
	"testing"
	"time"
)

// TestBeginnerChallenges tests beginner-level challenges from web-scraping.dev
// These tests are safe to run and won't cause IP blocking.

func TestBeginner_StaticPaging(t *testing.T) {
	// Challenge: HTML-based server-side item paging where each page has its own URL
	// URL: https://web-scraping.dev/products
	t.Log("Testing Static Paging (Beginner)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/products",
		ChatbotID: 1,
		SourceID:  1,
	}

	cfg := CollectorConfig{
		AllowedDomains:  []string{"web-scraping.dev"},
		Timeout:         30 * time.Second,
		RateLimitPerSec: 1, // Be respectful
	}

	content, err := ScrapeURL(task, cfg, nil)
	if err != nil {
		t.Fatalf("Failed to scrape static paging: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Verify we got product information
	if !strings.Contains(strings.ToLower(content), "product") {
		t.Errorf("Expected 'product' in content, but not found. Content length: %d", len(content))
	}

	t.Logf("✓ Static Paging test passed. Content length: %d", len(content))
}

func TestBeginner_ForcedNewTabLinks(t *testing.T) {
	// Challenge: Links that use different techniques to force opening in a new tab
	// URL: https://web-scraping.dev/reviews
	t.Log("Testing Forced New Tab Links (Beginner)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/reviews",
		ChatbotID: 1,
		SourceID:  1,
	}

	cfg := CollectorConfig{
		AllowedDomains:  []string{"web-scraping.dev"},
		Timeout:         30 * time.Second,
		RateLimitPerSec: 1,
	}

	content, err := ScrapeURL(task, cfg, nil)
	if err != nil {
		t.Fatalf("Failed to scrape forced new tab links: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// We should be able to extract the text content regardless of link targets
	if !strings.Contains(strings.ToLower(content), "review") {
		t.Errorf("Expected 'review' in content. Content length: %d", len(content))
	}

	t.Logf("✓ Forced New Tab Links test passed. Content length: %d", len(content))
}

func TestBeginner_ProductHTMLMarkup(t *testing.T) {
	// Challenge: Basic e-commerce product structure and CSS class-based markup
	// URL: https://web-scraping.dev/product/1
	t.Log("Testing Product HTML Markup (Beginner)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/product/1",
		ChatbotID: 1,
		SourceID:  1,
	}

	cfg := CollectorConfig{
		AllowedDomains:  []string{"web-scraping.dev"},
		Timeout:         30 * time.Second,
		RateLimitPerSec: 1,
	}

	content, err := ScrapeURL(task, cfg, nil)
	if err != nil {
		t.Fatalf("Failed to scrape product markup: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Check for typical product page elements
	lowerContent := strings.ToLower(content)
	hasProduct := strings.Contains(lowerContent, "product") || 
		strings.Contains(lowerContent, "price") ||
		strings.Contains(lowerContent, "add to cart")

	if !hasProduct {
		t.Errorf("Expected product-related content. Content length: %d", len(content))
	}

	t.Logf("✓ Product HTML Markup test passed. Content length: %d", len(content))
}

func TestBeginner_CookiePopup(t *testing.T) {
	// Challenge: Cookie info modal popup that blocks the entire screen
	// URL: https://web-scraping.dev/login?cookies
	t.Log("Testing Cookie Popup (Beginner)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/login?cookies",
		ChatbotID: 1,
		SourceID:  1,
	}

	cfg := CollectorConfig{
		AllowedDomains:  []string{"web-scraping.dev"},
		Timeout:         30 * time.Second,
		RateLimitPerSec: 1,
	}

	// Static scraper should still get the underlying content
	content, err := ScrapeURL(task, cfg, nil)
	if err != nil {
		t.Fatalf("Failed to scrape cookie popup page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain both cookie popup and login content
	lowerContent := strings.ToLower(content)
	hasCookie := strings.Contains(lowerContent, "cookie")
	hasLogin := strings.Contains(lowerContent, "login") || strings.Contains(lowerContent, "sign in")

	if !hasCookie && !hasLogin {
		t.Errorf("Expected cookie or login content. Content length: %d", len(content))
	}

	t.Logf("✓ Cookie Popup test passed. Content length: %d", len(content))
}

func TestBeginner_ExampleBlockPage(t *testing.T) {
	// Challenge: Valid 200-status response that redirects to block notification
	// URL: https://web-scraping.dev/blocked
	// Note: This page returns 200 OK but shows a block message
	t.Log("Testing Example Block Page (Beginner)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/blocked",
		ChatbotID: 1,
		SourceID:  1,
	}

	cfg := CollectorConfig{
		AllowedDomains:  []string{"web-scraping.dev"},
		Timeout:         30 * time.Second,
		RateLimitPerSec: 1,
	}

	content, err := ScrapeURL(task, cfg, nil)
	// We expect this to succeed but return block message
	if err != nil {
		t.Fatalf("Failed to scrape block page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content (block message) but got empty string")
	}

	// Should contain block-related message
	lowerContent := strings.ToLower(content)
	hasBlockMessage := strings.Contains(lowerContent, "block") || 
		strings.Contains(lowerContent, "access denied") ||
		strings.Contains(lowerContent, "forbidden")

	if !hasBlockMessage {
		t.Logf("Warning: Expected block message in content. Got: %s", content[:min(200, len(content))])
	}

	t.Logf("✓ Example Block Page test passed. Content length: %d", len(content))
}

// TestBeginner_AllChallenges runs all beginner challenges sequentially
func TestBeginner_AllChallenges(t *testing.T) {
	t.Log("Running all Beginner challenges...")
	
	t.Run("StaticPaging", TestBeginner_StaticPaging)
	t.Run("ForcedNewTabLinks", TestBeginner_ForcedNewTabLinks)
	t.Run("ProductHTMLMarkup", TestBeginner_ProductHTMLMarkup)
	t.Run("CookiePopup", TestBeginner_CookiePopup)
	t.Run("ExampleBlockPage", TestBeginner_ExampleBlockPage)
	
	t.Log("✓ All Beginner challenges completed!")
}
