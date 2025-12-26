# Issue 005: Inconsistent Error Handling and Resource Cleanup in Sitemap Parsing

## Priority: Low
## Confidence: Medium

## Summary

The `DiscoverSitemapURL` function performs network requests in a loop without per-request timeouts or consistent response body handling, potentially leading to connection leaks or hanging processes.

## Evidence

**File:** [sitemap_parser.go](file:///Users/onur/Documents/workspace/botla-co/internal/scraper/sitemap_parser.go#L277-L297)

```go
func DiscoverSitemapURL(ctx context.Context, baseURL string) (string, error) {
    // ...
    client := &http.Client{Timeout: 10 * time.Second}

    for _, path := range commonPaths {
        testURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, path)

        req, err := http.NewRequestWithContext(ctx, http.MethodHead, testURL, nil)
        if err != nil {
            continue
        }
        req.Header.Set("User-Agent", "Botla-Sitemap-Parser/1.0")

        resp, err := client.Do(req)
        if err != nil {
            continue
        }
        _ = resp.Body.Close()

        if resp.StatusCode == http.StatusOK {
            return testURL, nil
        }
    }
    // ...
}
```

## Issues Identified

### 1. No Per-Request Timeout Granularity

The 10-second timeout applies to the entire client, but not individually to each request in the loop. If the context passed in has a longer timeout, individual requests could block for extended periods.

**Impact**: In the worst case, discovery of 4 paths × 10 seconds = 40 seconds blocking.

### 2. Inconsistent Body Closure Timing

```go
resp, err := client.Do(req)
if err != nil {
    continue  // Body is nil, no cleanup needed - OK
}
_ = resp.Body.Close()  // Closed here - OK

if resp.StatusCode == http.StatusOK {
    return testURL, nil  // Body already closed - OK
}
```

Actually, the current code **does** close the body correctly before checking status. However, if future modifications add processing between response receipt and close, leaks could occur.

### 3. Missing Request Cancellation Propagation

If the parent context is cancelled mid-loop, in-flight requests will still complete (up to their timeout). There's no early exit on context cancellation between iterations.

### 4. HEAD Request Assumption

Some servers don't properly support HEAD requests or return different status codes for HEAD vs GET. A 405 Method Not Allowed would cause the loop to skip a valid sitemap.

## Current Mitigations

The existing code has several positive aspects:
- Uses `http.NewRequestWithContext` for context propagation
- Properly closes response body after each request
- Uses limited reader in `fetchURL` (10MB cap)
- Client has a timeout configured

## Recommended Improvements

### Add Per-Request Context Timeout

```go
func DiscoverSitemapURL(ctx context.Context, baseURL string) (string, error) {
    parsed, err := url.Parse(baseURL)
    if err != nil {
        return "", fmt.Errorf("invalid base URL: %w", err)
    }

    commonPaths := []string{
        "/sitemap.xml",
        "/sitemap_index.xml",
        "/sitemap-index.xml",
        "/sitemaps/sitemap.xml",
    }

    client := &http.Client{
        Timeout: 10 * time.Second,
        // Prevent redirect following for HEAD requests
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 3 {
                return fmt.Errorf("too many redirects")
            }
            return nil
        },
    }

    for _, path := range commonPaths {
        // Check context cancellation between iterations
        if ctx.Err() != nil {
            return "", ctx.Err()
        }

        // Per-request timeout
        reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
        
        testURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, path)
        found, err := p.probeSitemap(reqCtx, client, testURL)
        cancel() // Always cancel to free resources
        
        if err != nil {
            continue
        }
        if found {
            return testURL, nil
        }
    }

    return "", fmt.Errorf("no sitemap found at common locations")
}

func (p *SitemapParser) probeSitemap(ctx context.Context, client *http.Client, url string) (bool, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
    if err != nil {
        return false, err
    }
    req.Header.Set("User-Agent", "Botla-Sitemap-Parser/1.0")

    resp, err := client.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    // Drain any body content to reuse connection
    _, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

    return resp.StatusCode == http.StatusOK, nil
}
```

### Consider GET Fallback for HEAD Failures

```go
func (p *SitemapParser) probeSitemap(ctx context.Context, client *http.Client, url string) (bool, error) {
    // Try HEAD first
    found, err := p.probeWithMethod(ctx, client, url, http.MethodHead)
    if err == nil {
        return found, nil
    }
    
    // Fallback to GET with range header for servers that don't support HEAD
    return p.probeWithMethod(ctx, client, url, http.MethodGet)
}
```

## Verification

1. Unit test: Verify context cancellation stops discovery loop
2. Unit test: Mock slow server, verify per-request timeout works
3. Integration test: Discover sitemap from real sites with different configurations
4. Test coverage for 4xx/5xx responses ensuring proper cleanup

## Related Files

- `internal/scraper/sitemap_parser.go` - Main sitemap parsing logic
- `internal/scraper/url_processor.go` - Uses sitemap URLs for crawling
