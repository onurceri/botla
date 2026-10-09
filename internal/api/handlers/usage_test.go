package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/middleware"
)

// stubUserRepo implements only GetByID; other methods panic via the nil embed.
type stubUserRepo struct {
	repository.UserRepository
	user *models.User
	err  error
}

func (s stubUserRepo) GetByID(context.Context, string) (*models.User, error) { return s.user, s.err }

// stubUsageRepo implements the read methods GetUsage uses.
type stubUsageRepo struct {
	repository.UsageRepository
	failOn        string
	byUser, byWS  int
	workspaceSeen string
}

func (s *stubUsageRepo) fail(name string) error {
	if s.failOn == name {
		return errors.New(name + " failed")
	}
	return nil
}

func (s *stubUsageRepo) CountChatbotsByUserID(context.Context, string) (int, error) {
	return s.byUser, s.fail("chatbots")
}

func (s *stubUsageRepo) CountChatbotsByWorkspace(_ context.Context, ws string) (int, error) {
	s.workspaceSeen = ws
	return s.byWS, s.fail("chatbots")
}

func (s *stubUsageRepo) GetFileCountByUserID(context.Context, string) (int, error) {
	return 4, s.fail("files")
}

func (s *stubUsageRepo) GetURLCountByUserID(context.Context, string) (int, error) {
	return 5, s.fail("urls")
}

func (s *stubUsageRepo) GetMonthlyTokenUsage(context.Context, string) (int, error) {
	return 1234, s.fail("tokens")
}

func (s *stubUsageRepo) GetStorageUsedMBByUserID(context.Context, string) (int, error) {
	return 7, s.fail("storage")
}

func (s *stubUsageRepo) GetMonthlyIngestionUsage(context.Context, string, time.Time) (int, int, error) {
	return 8, 900, s.fail("ingestion")
}

func (s *stubUsageRepo) GetMaxFileCountInAnyBot(context.Context, string) (int, error) {
	return 3, s.fail("maxfiles")
}

func (s *stubUsageRepo) GetMaxURLCountInAnyBot(context.Context, string) (int, error) {
	return 2, s.fail("maxurls")
}

func (s *stubUsageRepo) GetMonthlyRefreshCount(context.Context, string, time.Time) (int, error) {
	return 6, s.fail("refresh")
}

func usageRequest(userID, workspaceID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/usage", nil)
	ctx := req.Context()
	if userID != "" {
		ctx = context.WithValue(ctx, middleware.ContextKeyUserID, userID)
	}
	if workspaceID != "" {
		ctx = context.WithValue(ctx, middleware.ContextKeyWorkspaceID, workspaceID)
	}
	return req.WithContext(ctx)
}

func TestGetUsage_ReturnsAggregatedUsage(t *testing.T) {
	usage := &stubUsageRepo{byUser: 9, byWS: 1}
	h := &UsageHandlers{UserRepo: stubUserRepo{user: &models.User{ID: "u1"}}, UsageRepo: usage}

	rr := httptest.NewRecorder()
	h.GetUsage(rr, usageRequest("u1", ""))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got models.Usage
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := models.Usage{
		ChatbotsCount: 9, FilesCount: 4, MaxFilesCountInOneBot: 3, StorageUsedMB: 7,
		URLsCount: 5, MaxURLsCountInOneBot: 2, TokensUsed: 1234, IngestionsUsed: 8,
		IngestionEmbeddingTokens: 900, RefreshCount: 6,
	}
	if got != want {
		t.Fatalf("usage = %+v, want %+v", got, want)
	}
	if usage.workspaceSeen != "" {
		t.Fatal("workspace count must not be used without a workspace context")
	}
}

func TestGetUsage_WorkspaceScopesChatbotCount(t *testing.T) {
	usage := &stubUsageRepo{byUser: 9, byWS: 1}
	h := &UsageHandlers{UserRepo: stubUserRepo{user: &models.User{ID: "u1"}}, UsageRepo: usage}

	rr := httptest.NewRecorder()
	h.GetUsage(rr, usageRequest("u1", "ws-1"))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got models.Usage
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ChatbotsCount != 1 || usage.workspaceSeen != "ws-1" {
		t.Fatalf("expected workspace-scoped chatbot count, got %d (ws=%q)", got.ChatbotsCount, usage.workspaceSeen)
	}
}

func TestGetUsage_Errors(t *testing.T) {
	tests := []struct {
		name   string
		userID string
		users  stubUserRepo
		failOn string
		want   int
	}{
		{name: "unauthenticated", userID: "", want: http.StatusUnauthorized},
		{name: "user lookup fails", userID: "u1", users: stubUserRepo{err: errors.New("db down")}, want: http.StatusInternalServerError},
		{name: "user not found", userID: "u1", users: stubUserRepo{}, want: http.StatusNotFound},
	}
	for _, step := range []string{"chatbots", "files", "urls", "tokens", "storage", "ingestion", "maxfiles", "maxurls", "refresh"} {
		tests = append(tests, struct {
			name   string
			userID string
			users  stubUserRepo
			failOn string
			want   int
		}{name: step + " query fails", userID: "u1", users: stubUserRepo{user: &models.User{ID: "u1"}}, failOn: step, want: http.StatusInternalServerError})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &UsageHandlers{UserRepo: tt.users, UsageRepo: &stubUsageRepo{failOn: tt.failOn}}
			rr := httptest.NewRecorder()
			h.GetUsage(rr, usageRequest(tt.userID, ""))
			if rr.Code != tt.want {
				t.Fatalf("expected %d, got %d", tt.want, rr.Code)
			}
		})
	}
}
