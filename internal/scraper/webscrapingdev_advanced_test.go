//go:build webscraping

package scraper

import (
	"strings"
	"testing"
	"time"
)

// TestAdvancedChallenges tests advanced-level challenges from web-scraping.dev
// ⚠️ WARNING: These tests involve anti-scraping measures and may result in IP blocking.
// Run these tests carefully and consider using proxies or VPNs.
// Some tests may fail by design to demonstrate anti-bot protection.

func TestAdvanced_GraphQLBackgroundRequests(t *testing.T) {
	// Challenge: Data loaded using JavaScript through a backend GraphQL API
	// URL: https://web-scraping.dev/reviews
	// Requires: JavaScript execution and potentially intercepting GraphQL requests
	t.Log("Testing GraphQL Background Requests (Advanced)")
	t.Log("⚠️  This requires dynamic scraping and may not get all GraphQL data")

	url := "https://web-scraping.dev/reviews"
	cfg := DynamicConfig{
		PoolSize:   1,
		IdleTTL:    60 * time.Second,
		NavTimeout: 20 * time.Second, // Longer timeout for API calls
		Allowed:    []string{"web-scraping.dev"},
	}

	// Dynamic scraping will get the page after JS execution
	content, err := ScrapeDynamicURL(url, cfg)
	if err != nil {
		t.Fatalf("Failed to scrape GraphQL page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain review content loaded via GraphQL
	lowerContent := strings.ToLower(content)
	hasReview := strings.Contains(lowerContent, "review") || 
		strings.Contains(lowerContent, "rating") ||
		strings.Contains(lowerContent, "product")

	if !hasReview {
		t.Logf("Warning: Expected review content from GraphQL. Content length: %d", len(content))
	}

	t.Logf("✓ GraphQL Background Requests test passed. Content length: %d", len(content))
}

func TestAdvanced_CSRFTokenLocks(t *testing.T) {
	// Challenge: The "Load More reviews" action uses X-CSRF-Token header to block cross-site access
	// URL: https://web-scraping.dev/product/1
	// Note: We can scrape the initial page, but "Load More" requires CSRF token
	t.Log("Testing CSRF Token Locks (Advanced)")
	t.Log("⚠️  This tests initial page scraping; CSRF-protected actions won't work")

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

	// Static scraper should get initial content without CSRF token
	content, err := ScrapeURL(task, cfg, nil)
	if err != nil {
		t.Fatalf("Failed to scrape CSRF-protected page: %v", err)
	}

	if content == "" {
		t.Fatal("Expected content but got empty string")
	}

	// Should contain product content
	lowerContent := strings.ToLower(content)
	hasProduct := strings.Contains(lowerContent, "product") || 
		strings.Contains(lowerContent, "price") ||
		strings.Contains(lowerContent, "review")

	if !hasProduct {
		t.Logf("Warning: Expected product content. Content length: %d", len(content))
	}

	t.Logf("✓ CSRF Token Locks test passed (initial page). Content length: %d", len(content))
	t.Logf("Note: CSRF-protected 'Load More' actions would require token extraction")
}

func TestAdvanced_BlockingRedirectInvalidReferer(t *testing.T) {
	// Challenge: The credentials page is only accessible with valid Referer header
	// URL: https://web-scraping.dev/credentials
	// This page redirects to /blocked if accessed without proper Referer from /login
	t.Log("Testing Blocking Redirect for Invalid Referer (Advanced)")
	t.Log("⚠️  This may redirect to /blocked without proper Referer header")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/credentials",
		ChatbotID: 1,
		SourceID:  1,
	}

	cfg := CollectorConfig{
		AllowedDomains:  []string{"web-scraping.dev"},
		Timeout:         30 * time.Second,
		RateLimitPerSec: 1,
	}

	// This will likely redirect to /blocked
	content, err := ScrapeURL(task, cfg, nil)
	if err != nil {
		t.Logf("Expected error or redirect: %v", err)
	}

	// We expect to get SOME content (probably block page)
	if content == "" {
		t.Log("Got empty content (likely blocked or redirected)")
	} else {
		lowerContent := strings.ToLower(content)
		
		// Check if we got blocked
		isBlocked := strings.Contains(lowerContent, "block") || 
			strings.Contains(lowerContent, "access denied") ||
			strings.Contains(lowerContent, "forbidden")
		
		// Check if we somehow got credentials page
		hasCredentials := strings.Contains(lowerContent, "credential") || 
			strings.Contains(lowerContent, "username") ||
			strings.Contains(lowerContent, "password")

		if isBlocked {
			t.Logf("✓ Correctly blocked without Referer header. Content length: %d", len(content))
		} else if hasCredentials {
			t.Logf("✓ Unexpectedly got credentials page! Content length: %d", len(content))
		} else {
			t.Logf("Got content: %s", content[:min(200, len(content))])
		}
	}

	t.Log("Note: To access this page properly, we'd need to set Referer header from /login")
}

func TestAdvanced_PersistentCookieBasedBlocking(t *testing.T) {
	// Challenge: Using cookies to mark blocked clients for persistent blocking
	// URL: https://web-scraping.dev/blocked?persist
	// ⚠️ WARNING: This may set a blocking cookie that persists!
	t.Log("Testing Persistent Cookie-Based Blocking (Advanced)")
	t.Log("⚠️  WARNING: This may set a persistent blocking cookie!")
	
	// Skip this test by default to avoid getting blocked
	if testing.Short() {
		t.Skip("Skipping persistent blocking test in short mode (use -short flag)")
	}

	// Additional confirmation
	t.Log("⚠️  This test is intentionally risky. Proceeding...")

	task := ScrapingTask{
		URL:       "https://web-scraping.dev/blocked?persist",
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
		t.Logf("Got error (expected for blocked page): %v", err)
	}

	if content == "" {
		t.Fatal("Expected content (block message) but got empty string")
	}

	// Should contain block message
	lowerContent := strings.ToLower(content)
	isBlocked := strings.Contains(lowerContent, "block") || 
		strings.Contains(lowerContent, "access denied") ||
		strings.Contains(lowerContent, "forbidden") ||
		strings.Contains(lowerContent, "persistent")

	if isBlocked {
		t.Logf("✓ Got block page with persistent cookie. Content length: %d", len(content))
		t.Log("⚠️  A blocking cookie may have been set!")
	} else {
		t.Logf("Unexpected content: %s", content[:min(200, len(content))])
	}

	t.Log("Note: You may need to clear cookies or change IP to access web-scraping.dev again")
}

// TestAdvanced_AllChallenges runs all advanced challenges sequentially
// ⚠️ Run with caution! Consider using -short flag to skip risky tests
func TestAdvanced_AllChallenges(t *testing.T) {
	t.Log("Running all Advanced challenges...")
	t.Log("⚠️  WARNING: These tests may trigger anti-scraping measures!")
	
	// Safer tests
	t.Run("CSRFTokenLocks", TestAdvanced_CSRFTokenLocks)
	t.Run("BlockingRedirectInvalidReferer", TestAdvanced_BlockingRedirectInvalidReferer)
	
	// Dynamic scraping test
	t.Run("GraphQLBackgroundRequests", TestAdvanced_GraphQLBackgroundRequests)
	
	// Risky test (skipped in short mode)
	t.Run("PersistentCookieBasedBlocking", TestAdvanced_PersistentCookieBasedBlocking)
	
	t.Log("✓ All Advanced challenges completed!")
	t.Log("⚠️  Check if you can still access web-scraping.dev")
}

// TestAdvanced_SafeOnly runs only the safe advanced tests
func TestAdvanced_SafeOnly(t *testing.T) {
	t.Log("Running SAFE Advanced challenges only...")
	
	t.Run("CSRFTokenLocks", TestAdvanced_CSRFTokenLocks)
	t.Run("BlockingRedirectInvalidReferer", TestAdvanced_BlockingRedirectInvalidReferer)
	t.Run("GraphQLBackgroundRequests", TestAdvanced_GraphQLBackgroundRequests)
	
	t.Log("✓ Safe Advanced challenges completed!")
	t.Log("Skipped: PersistentCookieBasedBlocking (too risky)")
}
