# Task 007: SSRF Protection for URL Sources

**Priority:** 🔴 Critical (Security)  
**Phase:** 4 - Security Hardening  
**Estimated Time:** 3-4 hours  
**Dependencies:** None  

---

## Problem Statement

The URL source feature allows users to provide arbitrary URLs for scraping. This creates Server-Side Request Forgery (SSRF) risk where attackers could:
- Access internal services (localhost, 127.0.0.1, internal IPs)
- Scan internal networks
- Access cloud metadata endpoints (169.254.169.254)
- Bypass firewalls

**Risk Level:** HIGH

---

## Objective

Implement comprehensive SSRF protection:
1. Block private/internal IP ranges
2. Block localhost and loopback addresses
3. Block cloud metadata endpoints
4. Block dangerous URL schemes
5. Validate resolved IP before making requests

---

## Implementation Details

### Step 1: Create SSRF Validator

**File:** `pkg/urlutil/ssrf.go` (NEW)

```go
package urlutil

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// SSRFValidator validates URLs against SSRF attacks
type SSRFValidator struct {
	allowPrivate bool // For testing only
}

// NewSSRFValidator creates a new SSRF validator
func NewSSRFValidator() *SSRFValidator {
	return &SSRFValidator{allowPrivate: false}
}

// BlockedSchemes are URL schemes that should never be allowed
var BlockedSchemes = []string{
	"file",
	"ftp",
	"gopher",
	"data",
	"javascript",
}

// BlockedHosts are hostnames that should never be allowed
var BlockedHosts = []string{
	"localhost",
	"127.0.0.1",
	"0.0.0.0",
	"[::1]",
	"metadata.google.internal",
}

// BlockedIPRanges are CIDR ranges for private/internal networks
var BlockedIPRanges = []string{
	"10.0.0.0/8",      // Private Class A
	"172.16.0.0/12",   // Private Class B
	"192.168.0.0/16",  // Private Class C
	"127.0.0.0/8",     // Loopback
	"169.254.0.0/16",  // Link-local (includes cloud metadata)
	"::1/128",         // IPv6 loopback
	"fc00::/7",        // IPv6 private
	"fe80::/10",       // IPv6 link-local
	"100.64.0.0/10",   // Carrier-grade NAT
	"0.0.0.0/8",       // Current network
}

var parsedBlockedRanges []*net.IPNet

func init() {
	for _, cidr := range BlockedIPRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			parsedBlockedRanges = append(parsedBlockedRanges, network)
		}
	}
}

// ValidateURL checks if a URL is safe from SSRF attacks
func (v *SSRFValidator) ValidateURL(rawURL string) error {
	// Parse URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check scheme
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("blocked URL scheme: %s", scheme)
	}

	for _, blocked := range BlockedSchemes {
		if scheme == blocked {
			return fmt.Errorf("blocked URL scheme: %s", scheme)
		}
	}

	// Check hostname
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("missing hostname")
	}

	for _, blocked := range BlockedHosts {
		if host == blocked {
			return fmt.Errorf("blocked hostname: %s", host)
		}
	}

	// Check for IP address directly in URL
	if ip := net.ParseIP(host); ip != nil {
		if err := v.validateIP(ip); err != nil {
			return err
		}
	}

	// Resolve hostname and validate IPs
	if !v.allowPrivate {
		ips, err := net.LookupIP(host)
		if err != nil {
			return fmt.Errorf("failed to resolve hostname: %w", err)
		}

		for _, ip := range ips {
			if err := v.validateIP(ip); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateIP checks if an IP is in a blocked range
func (v *SSRFValidator) validateIP(ip net.IP) error {
	if v.allowPrivate {
		return nil
	}

	// Check if loopback
	if ip.IsLoopback() {
		return fmt.Errorf("loopback address not allowed: %s", ip)
	}

	// Check if private
	if ip.IsPrivate() {
		return fmt.Errorf("private address not allowed: %s", ip)
	}

	// Check if link-local
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("link-local address not allowed: %s", ip)
	}

	// Check against blocked ranges
	for _, network := range parsedBlockedRanges {
		if network.Contains(ip) {
			return fmt.Errorf("address in blocked range: %s", ip)
		}
	}

	return nil
}

// ValidateURLStrict performs validation and also returns resolved IPs
// Use this for logging what IP was actually contacted
func (v *SSRFValidator) ValidateURLStrict(rawURL string) ([]net.IP, error) {
	if err := v.ValidateURL(rawURL); err != nil {
		return nil, err
	}

	parsed, _ := url.Parse(rawURL)
	host := parsed.Hostname()

	// Return resolved IPs for logging
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	return ips, nil
}
```

### Step 2: Add Error Code

**File:** `internal/api/error_codes.go` (MODIFY)

```go
const (
	// ... existing codes
	ErrBlockedURL = "ERR_BLOCKED_URL"
)
```

### Step 3: Integrate into URL Handler

**File:** `internal/api/handlers/source_create.go` (MODIFY)

```go
import "github.com/onurceri/botla-co/pkg/urlutil"

// At package level or in handler struct
var ssrfValidator = urlutil.NewSSRFValidator()

func (h *SourcesHandlers) handleURLSource(w http.ResponseWriter, r *http.Request, chatbotID string, plan *models.Plan, cfg langconfig.LanguageConfig) {
	// ... existing count check ...

	rawURL := strings.TrimSpace(r.FormValue("source_url"))
	if rawURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// NEW: SSRF validation
	if err := ssrfValidator.ValidateURL(rawURL); err != nil {
		h.logWarn("ssrf_blocked", map[string]any{
			"url":    rawURL,
			"reason": err.Error(),
		})
		api.WriteLocalizedError(w, http.StatusForbidden, api.ErrBlockedURL, cfg)
		return
	}

	// Normalize URL
	url, err := urlutil.NormalizeURL(rawURL)
	// ... rest of code ...
}
```

### Step 4: Integrate into Scraper

**File:** `internal/scraper/colly.go` (MODIFY)

Add validation before making request:

```go
func (s *CollyScraper) Scrape(ctx context.Context, targetURL string) (*ScrapeResult, error) {
	// Validate URL before scraping
	validator := urlutil.NewSSRFValidator()
	if err := validator.ValidateURL(targetURL); err != nil {
		return nil, fmt.Errorf("SSRF blocked: %w", err)
	}

	// ... existing scrape code ...
}
```

### Step 5: Add Localized Messages

**File:** `internal/api/errors_localized.go` (MODIFY)

```go
"ERR_BLOCKED_URL": {
	"en": "This URL cannot be accessed for security reasons",
	"tr": "Bu URL güvenlik nedeniyle erişilemez",
},
```

**File:** `frontend/src/i18n/errors.ts` (MODIFY)

```typescript
ERR_BLOCKED_URL: 'This URL is not allowed for security reasons',
// Turkish
ERR_BLOCKED_URL: 'Bu URL güvenlik nedeniyle engellenmiştir',
```

---

## Tests to Write

### Unit Tests

**File:** `pkg/urlutil/ssrf_test.go` (NEW)

```go
package urlutil

import (
	"testing"
)

func TestSSRFValidator_BlockedSchemes(t *testing.T) {
	v := NewSSRFValidator()

	blocked := []string{
		"file:///etc/passwd",
		"ftp://example.com/file",
		"gopher://example.com",
		"data:text/html,<script>alert(1)</script>",
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_BlockedHosts(t *testing.T) {
	v := NewSSRFValidator()

	blocked := []string{
		"http://localhost/path",
		"http://127.0.0.1/path",
		"http://0.0.0.0/path",
		"http://[::1]/path",
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_BlockedIPs(t *testing.T) {
	v := NewSSRFValidator()

	blocked := []string{
		"http://10.0.0.1/internal",
		"http://172.16.0.1/internal",
		"http://192.168.1.1/internal",
		"http://169.254.169.254/latest/meta-data/", // AWS metadata
	}

	for _, url := range blocked {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to be blocked", url)
		}
	}
}

func TestSSRFValidator_AllowedURLs(t *testing.T) {
	v := NewSSRFValidator()

	allowed := []string{
		"https://example.com",
		"https://www.google.com/search",
		"http://github.com",
	}

	for _, url := range allowed {
		err := v.ValidateURL(url)
		if err != nil {
			t.Errorf("expected %s to be allowed, got error: %v", url, err)
		}
	}
}

func TestSSRFValidator_InvalidURLs(t *testing.T) {
	v := NewSSRFValidator()

	invalid := []string{
		"not-a-url",
		"://missing-scheme",
		"http://", // empty host
	}

	for _, url := range invalid {
		err := v.ValidateURL(url)
		if err == nil {
			t.Errorf("expected %s to fail validation", url)
		}
	}
}
```

### Integration Test

**File:** `internal/integration/ssrf_test.go` (NEW)

```go
package integration

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestSSRFProtection_Integration(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	token := authToken(t, te.Server.URL, "ssrf@example.com")
	chatbotID := createChatbot(t, te.Server.URL, token, "SSRF Test Bot")

	blockedURLs := []string{
		"http://localhost/admin",
		"http://127.0.0.1:8080/internal",
		"http://192.168.1.1/router",
		"http://169.254.169.254/latest/meta-data/",
		"file:///etc/passwd",
	}

	for _, url := range blockedURLs {
		resp := createURLSource(t, te.Server.URL, token, chatbotID, url)
		
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("URL %s should be blocked with 403, got %d", url, resp.StatusCode)
		}

		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()

		if body["error"] != "ERR_BLOCKED_URL" {
			t.Errorf("expected ERR_BLOCKED_URL, got %s", body["error"])
		}
	}
}

func TestSSRFProtection_AllowsPublicURLs(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	token := authToken(t, te.Server.URL, "ssrfpublic@example.com")
	chatbotID := createChatbot(t, te.Server.URL, token, "Public URL Test")

	// This should succeed (though it might fail to scrape, it should pass SSRF check)
	resp := createURLSource(t, te.Server.URL, token, chatbotID, "https://example.com")
	
	// Should be 201 Created (might process successfully) or not 403 (SSRF blocked)
	if resp.StatusCode == http.StatusForbidden {
		var body map[string]string
		json.NewDecoder(resp.Body).Decode(&body)
		if body["error"] == "ERR_BLOCKED_URL" {
			t.Error("public URL should not be SSRF blocked")
		}
	}
}
```

---

## Verification Steps

1. **Run unit tests:**
   ```bash
   go test ./pkg/urlutil/... -v -run TestSSRF
   ```

2. **Run integration tests:**
   ```bash
   go test ./internal/integration/... -v -run TestSSRF
   ```

3. **Manual verification:**
   ```bash
   # Test blocked URL
   curl -X POST http://localhost:8080/api/v1/chatbots/{id}/sources \
     -H "Authorization: Bearer $TOKEN" \
     -F "source_type=url" \
     -F "source_url=http://localhost/admin"
   # Should return 403 with ERR_BLOCKED_URL

   # Test allowed URL
   curl -X POST http://localhost:8080/api/v1/chatbots/{id}/sources \
     -H "Authorization: Bearer $TOKEN" \
     -F "source_type=url" \
     -F "source_url=https://example.com"
   # Should return 201
   ```

---

## Acceptance Criteria

- [ ] Localhost and 127.0.0.1 are blocked
- [ ] Private IP ranges (10.x, 172.16.x, 192.168.x) are blocked
- [ ] Cloud metadata endpoints (169.254.169.254) are blocked
- [ ] file://, ftp://, gopher:// schemes are blocked
- [ ] Public URLs are allowed
- [ ] Appropriate error messages returned
- [ ] Blocked attempts are logged
- [ ] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `pkg/urlutil/ssrf.go` | CREATE |
| `pkg/urlutil/ssrf_test.go` | CREATE |
| `internal/api/error_codes.go` | MODIFY |
| `internal/api/handlers/source_create.go` | MODIFY |
| `internal/scraper/colly.go` | MODIFY |
| `internal/api/errors_localized.go` | MODIFY |
| `frontend/src/i18n/errors.ts` | MODIFY |
| `internal/integration/ssrf_test.go` | CREATE |

---

## Security Considerations

1. **DNS Rebinding:** The validator resolves DNS before making requests. For extra security, consider using a custom HTTP transport that validates IPs at connection time.

2. **IPv6:** Ensure all IPv6 private ranges are covered.

3. **Logging:** Log all blocked attempts for security monitoring.

4. **Time-of-check vs Time-of-use:** DNS could change between validation and actual request. Consider validating at request time in the HTTP transport.
