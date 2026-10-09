package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/testdb"
)

// TestRefreshHandler_ConcurrentRotation presents the same refresh token from
// many goroutines at once and asserts that exactly one request is allowed to
// rotate it. Requires Postgres (skipped with -short).
func TestRefreshHandler_ConcurrentRotation(t *testing.T) {
	if testing.Short() {
		t.Skip("requires Postgres")
	}
	db := testdb.OpenTestDB(t)
	h := &AuthHandlers{DB: db, Secret: "testsecret-concurrency-0123456789abcdef"}

	email := fmt.Sprintf("refresh-race+%d@example.com", time.Now().UnixNano())
	regBody := fmt.Sprintf(`{"email":%q,"password":"Test@123","full_name":"Race User"}`, email)
	rr := httptest.NewRecorder()
	h.RegisterHandler(rr, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(regBody)))
	if rr.Code != http.StatusCreated {
		t.Fatalf("register failed: %d %s", rr.Code, rr.Body.String())
	}
	var refreshToken string
	for _, c := range rr.Result().Cookies() {
		if c.Name == "botla_refresh_token" {
			refreshToken = c.Value
		}
	}
	if refreshToken == "" {
		t.Fatal("missing refresh token cookie")
	}

	const workers = 16
	var (
		wg    sync.WaitGroup
		start = make(chan struct{})
		mu    sync.Mutex
		codes = map[int]int{}
	)
	body := fmt.Sprintf(`{"refresh_token":%q}`, refreshToken)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			w := httptest.NewRecorder()
			h.RefreshHandler(w, httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body)))
			mu.Lock()
			codes[w.Code]++
			mu.Unlock()
		}()
	}
	close(start)
	wg.Wait()

	if codes[http.StatusOK] != 1 {
		t.Fatalf("expected exactly one successful rotation, got status counts %v", codes)
	}
	if codes[http.StatusUnauthorized] != workers-1 {
		t.Fatalf("expected %d rejected rotations, got status counts %v", workers-1, codes)
	}
}
