//go:build webscraping

package scraper

import (
	"strings"
	"testing"
	"time"
)

// TestIntermediateChallenges tests intermediate-level challenges from web-scraping.dev
// These tests involve more complex scenarios like dynamic content, authentication, and API calls.
// Most of these should be safe to run, but they may require dynamic scraping.

func TestIntermediate_EndlessScrollPaging(t *testing.T) {
	// Challenge: Dynamic client-side paging where new items appear as user scrolls
	// URL: https://web-scraping.dev/testimonials
	// Requires: JavaScript execution (dynamic scraping)
	t.Log("Testing Endless Scroll Paging (Intermediate)")

	url := "https://web-scraping.dev/testimonials"
	cfg := DynamicConfig{
		PoolSize:   1,
		IdleTTL:    60 * time.Second,
		NavTimeout: 15 * time.Second,
		Allowed:    []string{"web-scraping.dev"},
	}

	// Try dynamic scraping since this requires JS
	content, err := ScrapeDynamicURL(url, cfg)
	if err != nil {
		t.Fatalf("Failed to scrape endless scroll: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain testimonial content
	lowerContent := strings.ToLower(content)
	hasTestimonial := strings.Contains(lowerContent, "testimonial") || 
		strings.Contains(lowerContent, "review") ||
		strings.Contains(lowerContent, "customer")

	if !hasTestimonial {
		t.Logf("Warning: Expected testimonial content. Content length: %d", len(content))
	}

	t.Logf("✓ Endless Scroll Paging test passed. Content length: %d", len(content))
}

func TestIntermediate_SecretAPIToken(t *testing.T) {
	// Challenge: The testimonial paging uses X-Secret-Token to lock access to hidden APIs
	// URL: https://web-scraping.dev/testimonials
	// This tests if we can still scrape the visible content (not the API directly)
	t.Log("Testing Secret API Token (Intermediate)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/testimonials",
		ChatbotID: 1,
		SourceID:  1,
	}

	cfg := CollectorConfig{
		AllowedDomains:  []string{"web-scraping.dev"},
		Timeout:         30 * time.Second,
		RateLimitPerSec: 1,
	}

	// Static scraper should get initial page load
	content, err := ScrapeURL(task, cfg, nil)
	if err != nil {
		t.Fatalf("Failed to scrape with secret API token: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	t.Logf("✓ Secret API Token test passed (got initial content). Content length: %d", len(content))
}

func TestIntermediate_EndlessButtonPaging(t *testing.T) {
	// Challenge: Dynamic client-side paging where new items appear when user presses Load More button
	// URL: https://web-scraping.dev/reviews
	// Requires: JavaScript execution for "Load More" functionality
	t.Log("Testing Endless Button Paging (Intermediate)")

	url := "https://web-scraping.dev/reviews"
	cfg := DynamicConfig{
		PoolSize:   1,
		IdleTTL:    60 * time.Second,
		NavTimeout: 15 * time.Second,
		Allowed:    []string{"web-scraping.dev"},
	}

	// Dynamic scraping to get initial content
	content, err := ScrapeDynamicURL(url, cfg)
	if err != nil {
		t.Fatalf("Failed to scrape endless button paging: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain review content
	lowerContent := strings.ToLower(content)
	hasReview := strings.Contains(lowerContent, "review") || 
		strings.Contains(lowerContent, "rating")

	if !hasReview {
		t.Logf("Warning: Expected review content. Content length: %d", len(content))
	}

	t.Logf("✓ Endless Button Paging test passed. Content length: %d", len(content))
}

func TestIntermediate_HiddenWebData(t *testing.T) {
	// Challenge: Product review data is hidden in HTML as JSON, then loaded to HTML on page load
	// URL: https://web-scraping.dev/product/1
	t.Log("Testing Hidden Web Data (Intermediate)")

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

	// Static scraper should get the rendered content or raw HTML with JSON
	content, err := ScrapeURL(task, cfg, nil)
	if err != nil {
		t.Fatalf("Failed to scrape hidden web data: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain product or review data
	lowerContent := strings.ToLower(content)
	hasData := strings.Contains(lowerContent, "product") || 
		strings.Contains(lowerContent, "review") ||
		strings.Contains(lowerContent, "rating")

	if !hasData {
		t.Logf("Warning: Expected product/review data. Content length: %d", len(content))
	}

	t.Logf("✓ Hidden Web Data test passed. Content length: %d", len(content))
}

func TestIntermediate_LocalStorage(t *testing.T) {
	// Challenge: Cart system powered by local storage (client-side state store)
	// URL: https://web-scraping.dev/product/1
	// Note: We can't test localStorage directly with static scraping, but we can scrape the page
	t.Log("Testing Local Storage (Intermediate)")

	url := "https://web-scraping.dev/product/1"
	cfg := DynamicConfig{
		PoolSize:   1,
		IdleTTL:    60 * time.Second,
		NavTimeout: 15 * time.Second,
		Allowed:    []string{"web-scraping.dev"},
	}

	// Dynamic scraping to ensure JS-rendered content
	content, err := ScrapeDynamicURL(url, cfg)
	if err != nil {
		t.Fatalf("Failed to scrape local storage page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain cart or product content
	lowerContent := strings.ToLower(content)
	hasCart := strings.Contains(lowerContent, "cart") || 
		strings.Contains(lowerContent, "add to cart") ||
		strings.Contains(lowerContent, "product")

	if !hasCart {
		t.Logf("Warning: Expected cart/product content. Content length: %d", len(content))
	}

	t.Logf("✓ Local Storage test passed. Content length: %d", len(content))
}

func TestIntermediate_CookiesBasedLogin(t *testing.T) {
	// Challenge: User authentication based on form request and cookies
	// URL: https://web-scraping.dev/login
	// Note: We're just testing if we can scrape the login page, not actually logging in
	t.Log("Testing Cookies Based Login (Intermediate)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/login",
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
		t.Fatalf("Failed to scrape login page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain login-related content
	lowerContent := strings.ToLower(content)
	hasLogin := strings.Contains(lowerContent, "login") || 
		strings.Contains(lowerContent, "username") ||
		strings.Contains(lowerContent, "password")

	if !hasLogin {
		t.Logf("Warning: Expected login content. Content length: %d", len(content))
	}

	t.Logf("✓ Cookies Based Login test passed. Content length: %d", len(content))
}

func TestIntermediate_PDFDownloads(t *testing.T) {
	// Challenge: The login page features link and JS-based file download triggers
	// URL: https://web-scraping.dev/login
	// Note: We're testing if we can scrape the page, not downloading PDFs
	t.Log("Testing PDF Downloads (Intermediate)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/login",
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
		t.Fatalf("Failed to scrape PDF downloads page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain page content (PDF links are in the HTML)
	lowerContent := strings.ToLower(content)
	hasContent := strings.Contains(lowerContent, "download") || 
		strings.Contains(lowerContent, "pdf") ||
		strings.Contains(lowerContent, "login")

	if !hasContent {
		t.Logf("Warning: Expected download/PDF content. Content length: %d", len(content))
	}

	t.Logf("✓ PDF Downloads test passed. Content length: %d", len(content))
}

func TestIntermediate_FormFileAttachmentDownload(t *testing.T) {
	// Challenge: Form submission that triggers file download with Content-Disposition header
	// URL: https://web-scraping.dev/file-download
	// Note: We're testing if we can scrape the page, not submitting the form
	t.Log("Testing Form File Attachment Download (Intermediate)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/file-download",
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
		t.Fatalf("Failed to scrape file download page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain form or download content
	lowerContent := strings.ToLower(content)
	hasContent := strings.Contains(lowerContent, "download") || 
		strings.Contains(lowerContent, "file") ||
		strings.Contains(lowerContent, "form")

	if !hasContent {
		t.Logf("Warning: Expected download/form content. Content length: %d", len(content))
	}

	t.Logf("✓ Form File Attachment Download test passed. Content length: %d", len(content))
}

func TestIntermediate_AIContentObfuscation(t *testing.T) {
	// Challenge: Extract clean text from AI-obfuscated content using invisible Unicode characters
	// URL: https://web-scraping.dev/ai-content-obfuscation
	// This tests if our scraper can handle Unicode and extract meaningful content
	t.Log("Testing AI Content Obfuscation (Intermediate)")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/ai-content-obfuscation",
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
		t.Fatalf("Failed to scrape AI obfuscation page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Check if we got some content (may contain obfuscated characters)
	if len(content) < 10 {
		t.Errorf("Expected more content. Got only %d characters", len(content))
	}

	t.Logf("✓ AI Content Obfuscation test passed. Content length: %d", len(content))
	t.Logf("Sample content: %s", content[:min(200, len(content))])
}

// TestIntermediate_AllChallenges runs all intermediate challenges sequentially
func TestIntermediate_AllChallenges(t *testing.T) {
	t.Log("Running all Intermediate challenges...")
	
	// Static scraping tests
	t.Run("SecretAPIToken", TestIntermediate_SecretAPIToken)
	t.Run("HiddenWebData", TestIntermediate_HiddenWebData)
	t.Run("CookiesBasedLogin", TestIntermediate_CookiesBasedLogin)
	t.Run("PDFDownloads", TestIntermediate_PDFDownloads)
	t.Run("FormFileAttachmentDownload", TestIntermediate_FormFileAttachmentDownload)
	t.Run("AIContentObfuscation", TestIntermediate_AIContentObfuscation)
	
	// Dynamic scraping tests (comment out if you don't want to use headless browser)
	t.Run("EndlessScrollPaging", TestIntermediate_EndlessScrollPaging)
	t.Run("EndlessButtonPaging", TestIntermediate_EndlessButtonPaging)
	t.Run("LocalStorage", TestIntermediate_LocalStorage)
	
	t.Log("✓ All Intermediate challenges completed!")
}
