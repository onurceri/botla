package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestAuth_Cookies(t *testing.T) {
	db := testdb.OpenTestDB(t)

	// Setup user
	// Setup user
	var proPlanID string
	if err := db.QueryRow(`SELECT id FROM plans WHERE code='pro'`).Scan(&proPlanID); err != nil {
		t.Fatalf("plan: %v", err)
	}
	email := fmt.Sprintf("cookieuser+%d@example.com", time.Now().UnixNano())
	
	// We need actual hash if we login with password, but here we test handler that generates tokens directly
	// Or we can register first.
	
	// Let's use RegisterHandler to get cookies first.
	h := &AuthHandlers{DB: db, Secret: "testsecret"}

	// 1. Register and check cookies
	rr := httptest.NewRecorder()
	reqBody := fmt.Sprintf(`{"email":"%s","password":"password123","full_name":"Cookie User"}`, email)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(reqBody))
	h.RegisterHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("register failed: %d, %s", rr.Code, rr.Body.String())
	}

	cookies := rr.Result().Cookies()
	var accessToken, refreshToken string
	for _, c := range cookies {
		if c.Name == "botla_token" {
			accessToken = c.Value
			if !c.HttpOnly {
				t.Error("botla_token should be HttpOnly")
			}
			if c.Path != "/" {
				t.Error("botla_token path should be /")
			}
		}
		if c.Name == "botla_refresh_token" {
			refreshToken = c.Value
			if !c.HttpOnly {
				t.Error("botla_refresh_token should be HttpOnly")
			}
		}
	}

	if accessToken == "" || refreshToken == "" {
		t.Fatal("missing auth cookies")
	}

	// 2. Test Middleware with Cookie
	// Create a protected handler
	protected := middleware.AuthMiddleware("testsecret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	// Add cookie to request
	req2.AddCookie(&http.Cookie{Name: "botla_token", Value: accessToken})
	protected.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Errorf("protected handler with cookie failed: %d", rr2.Code)
	}

	// 3. Test Refresh with Cookie
	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil) // Empty body
	req3.AddCookie(&http.Cookie{Name: "botla_refresh_token", Value: refreshToken})
	h.RefreshHandler(rr3, req3)

	if rr3.Code != http.StatusOK {
		t.Errorf("refresh with cookie failed: %d", rr3.Code)
	}
	
	// Check if we got new cookies
	newCookies := rr3.Result().Cookies()
	foundNewToken := false
	for _, c := range newCookies {
		if c.Name == "botla_token" && c.Value != "" {
			foundNewToken = true
		}
	}
	if !foundNewToken {
		t.Error("refresh did not issue new cookie")
	}

	// 4. Test Logout clears cookies
	rr4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req4.AddCookie(&http.Cookie{Name: "botla_refresh_token", Value: refreshToken}) 
	// Note: Logout usually needs refresh token to find what to revoke. 
	// If we send empty body and cookie, it should work.
	h.LogoutHandler(rr4, req4)

	if rr4.Code != http.StatusOK {
		t.Errorf("logout failed: %d", rr4.Code)
	}
	
	logoutCookies := rr4.Result().Cookies()
	clearedAccess := false
	clearedRefresh := false
	for _, c := range logoutCookies {
		if c.Name == "botla_token" && c.MaxAge == -1 {
			clearedAccess = true
		}
		if c.Name == "botla_refresh_token" && c.MaxAge == -1 {
			clearedRefresh = true
		}
	}
	
	if !clearedAccess || !clearedRefresh {
		t.Error("logout did not clear cookies")
	}
}
