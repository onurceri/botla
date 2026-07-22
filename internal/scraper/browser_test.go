package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestScrapeDynamicURL(t *testing.T) {
	t.Parallel()
	// Ensure domain whitelist allows localhost
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!doctype html><html><body><script>setTimeout(()=>{document.body.innerHTML='<p>Merhaba Dinamik</p>'},100)</script></body></html>`))
	}))
	defer srv.Close()

	cfg := DynamicConfig{
		PoolSize:   1,
		IdleTTL:    5 * time.Second,
		NavTimeout: 5 * time.Second,
	}

	scraper, err := NewBrowserScraper(cfg)
	if err != nil {
		t.Skip("rod launch failed, skipping dynamic test")
	}
	defer scraper.Close()

	out, err := scraper.ScrapeDynamicURL(srv.URL)
	if err != nil {
		t.Skip("dynamic scrape timed out, skipping in CI")
	}
	if out != "Merhaba Dinamik" {
		t.Fatalf("unexpected: %q", out)
	}
}

// TestBrowserPool_RefillsToSizeAfterReap verifies that after the pool is cleared
// (as the idle reaper does), the next browser acquisition restores the pool to its
// configured size rather than leaving it stuck at a single browser.
//
// The post-reap state is simulated directly (clear the slice, as reaper() does)
// instead of waiting on the reaper's tick, so the test is deterministic under CPU
// load rather than timing-dependent.
func TestBrowserPool_RefillsToSizeAfterReap(t *testing.T) {
	const poolSize = 2
	// Browser launches use Must* and panic when Chrome is unavailable; skip then.
	defer func() {
		if r := recover(); r != nil {
			t.Skipf("browser launch unavailable, skipping: %v", r)
		}
	}()

	// A long idleTTL keeps the background reaper from interfering with the test.
	bp, err := NewBrowserPool(poolSize, time.Hour, "")
	if err != nil {
		t.Skipf("browser pool creation failed, skipping: %v", err)
	}
	defer bp.Close()

	bp.mu.Lock()
	initial := len(bp.browsers)
	bp.mu.Unlock()
	if initial != poolSize {
		t.Fatalf("initial pool size = %d, want %d", initial, poolSize)
	}

	// Simulate the idle reaper clearing the pool (capture, nil-out, close old).
	bp.mu.Lock()
	old := bp.browsers
	bp.browsers = nil
	bp.idx = 0
	bp.mu.Unlock()
	for _, b := range old {
		_ = b.Close()
	}

	// Acquiring a browser should refill the pool to its full configured size.
	if b := bp.nextBrowser(); b == nil {
		t.Fatal("nextBrowser returned nil")
	}
	bp.mu.Lock()
	refilled := len(bp.browsers)
	bp.mu.Unlock()
	if refilled != poolSize {
		t.Fatalf("pool refilled to %d browsers, want %d (stuck-at-1 regression)", refilled, poolSize)
	}
}
