package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-co/internal/repository"
)

// mockAdminChatbotRepo is a mock implementation of AdminChatbotRepository for handler tests.
type mockAdminChatbotRepo struct {
	ListChatbotsFunc  func(ctx context.Context, filter repository.AdminChatbotFilter, limit, offset int) ([]repository.AdminChatbot, int, error)
	GetByIDFunc       func(ctx context.Context, id string) (*repository.AdminChatbot, error)
	ResetSourcesFunc  func(ctx context.Context, chatbotID string) (int64, error)
	GetSourceIDsFunc  func(ctx context.Context, chatbotID string) ([]string, error)
	DeleteVectorsFunc func(ctx context.Context, chatbotID string) error

	calls struct {
		ListChatbots  int
		GetByID       []string
		ResetSources  []string
		GetSourceIDs  []string
		DeleteVectors []string
	}
}

func (m *mockAdminChatbotRepo) ListChatbots(ctx context.Context, filter repository.AdminChatbotFilter, limit, offset int) ([]repository.AdminChatbot, int, error) {
	m.calls.ListChatbots++
	if m.ListChatbotsFunc != nil {
		return m.ListChatbotsFunc(ctx, filter, limit, offset)
	}
	return nil, 0, nil
}

func (m *mockAdminChatbotRepo) GetByID(ctx context.Context, id string) (*repository.AdminChatbot, error) {
	m.calls.GetByID = append(m.calls.GetByID, id)
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockAdminChatbotRepo) ResetSources(ctx context.Context, chatbotID string) (int64, error) {
	m.calls.ResetSources = append(m.calls.ResetSources, chatbotID)
	if m.ResetSourcesFunc != nil {
		return m.ResetSourcesFunc(ctx, chatbotID)
	}
	return 0, nil
}

func (m *mockAdminChatbotRepo) GetSourceIDs(ctx context.Context, chatbotID string) ([]string, error) {
	m.calls.GetSourceIDs = append(m.calls.GetSourceIDs, chatbotID)
	if m.GetSourceIDsFunc != nil {
		return m.GetSourceIDsFunc(ctx, chatbotID)
	}
	return nil, nil
}

func (m *mockAdminChatbotRepo) DeleteVectors(ctx context.Context, chatbotID string) error {
	m.calls.DeleteVectors = append(m.calls.DeleteVectors, chatbotID)
	if m.DeleteVectorsFunc != nil {
		return m.DeleteVectorsFunc(ctx, chatbotID)
	}
	return nil
}

// TestAdminChatbotHandlers_ListChatbots tests the ListChatbots handler.
func TestAdminChatbotHandlers_ListChatbots(t *testing.T) {
	t.Parallel()

	t.Run("success with default pagination", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			ListChatbotsFunc: func(ctx context.Context, filter repository.AdminChatbotFilter, limit, offset int) ([]repository.AdminChatbot, int, error) {
				if limit != 20 {
					t.Errorf("expected default limit 20, got %d", limit)
				}
				if offset != 0 {
					t.Errorf("expected default offset 0, got %d", offset)
				}
				return []repository.AdminChatbot{
					{ID: "bot-1", Name: "Bot 1", OwnerEmail: "owner1@example.com"},
					{ID: "bot-2", Name: "Bot 2", OwnerEmail: "owner2@example.com"},
				}, 2, nil
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots", nil)
		w := httptest.NewRecorder()

		h.ListChatbots(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var response map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		chatbots, ok := response["chatbots"].([]any)
		if !ok {
			t.Fatalf("expected chatbots array in response")
		}
		if len(chatbots) != 2 {
			t.Errorf("expected 2 chatbots, got %d", len(chatbots))
		}

		total, ok := response["total"].(float64)
		if !ok {
			t.Fatalf("expected total in response")
		}
		if int(total) != 2 {
			t.Errorf("expected total 2, got %d", int(total))
		}
	})

	t.Run("success with custom pagination", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			ListChatbotsFunc: func(ctx context.Context, filter repository.AdminChatbotFilter, limit, offset int) ([]repository.AdminChatbot, int, error) {
				if limit != 10 {
					t.Errorf("expected limit 10, got %d", limit)
				}
				if offset != 5 {
					t.Errorf("expected offset 5, got %d", offset)
				}
				return []repository.AdminChatbot{}, 0, nil
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots?limit=10&offset=5", nil)
		w := httptest.NewRecorder()

		h.ListChatbots(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("success with filters", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			ListChatbotsFunc: func(ctx context.Context, filter repository.AdminChatbotFilter, limit, offset int) ([]repository.AdminChatbot, int, error) {
				if filter.Name == nil || *filter.Name != "TestBot" {
					t.Error("expected name filter 'TestBot'")
				}
				if filter.OrganizationID == nil || *filter.OrganizationID != "org-123" {
					t.Error("expected organization_id filter 'org-123'")
				}
				if filter.OwnerID == nil || *filter.OwnerID != "user-456" {
					t.Error("expected owner_id filter 'user-456'")
				}
				return []repository.AdminChatbot{}, 0, nil
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots?name=TestBot&organization_id=org-123&owner_id=user-456", nil)
		w := httptest.NewRecorder()

		h.ListChatbots(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("handles negative pagination values", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			ListChatbotsFunc: func(ctx context.Context, filter repository.AdminChatbotFilter, limit, offset int) ([]repository.AdminChatbot, int, error) {
				if limit != 20 {
					t.Errorf("expected default limit 20, got %d", limit)
				}
				if offset != 0 {
					t.Errorf("expected default offset 0, got %d", offset)
				}
				return []repository.AdminChatbot{}, 0, nil
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots?limit=-5&offset=-10", nil)
		w := httptest.NewRecorder()

		h.ListChatbots(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("returns error on repository failure", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			ListChatbotsFunc: func(ctx context.Context, filter repository.AdminChatbotFilter, limit, offset int) ([]repository.AdminChatbot, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots", nil)
		w := httptest.NewRecorder()

		h.ListChatbots(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}

// TestAdminChatbotHandlers_GetChatbot tests the GetChatbot handler.
func TestAdminChatbotHandlers_GetChatbot(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return &repository.AdminChatbot{
					ID:           id,
					Name:         "Test Chatbot",
					OwnerEmail:   "owner@example.com",
					SourceCount:  5,
					MessageCount: 100,
				}, nil
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots/bot-123", nil)
		req.SetPathValue("id", "bot-123")
		w := httptest.NewRecorder()

		h.GetChatbot(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var response repository.AdminChatbot
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.ID != "bot-123" {
			t.Errorf("expected ID 'bot-123', got %s", response.ID)
		}
		if response.Name != "Test Chatbot" {
			t.Errorf("expected name 'Test Chatbot', got %s", response.Name)
		}
	})

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{}
		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots/", nil)
		// Don't set path value
		w := httptest.NewRecorder()

		h.GetChatbot(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return nil, sql.ErrNoRows
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots/nonexistent", nil)
		req.SetPathValue("id", "nonexistent")
		w := httptest.NewRecorder()

		h.GetChatbot(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return nil, errors.New("database error")
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/chatbots/bot-123", nil)
		req.SetPathValue("id", "bot-123")
		w := httptest.NewRecorder()

		h.GetChatbot(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}

// TestAdminChatbotHandlers_ForceRefreshChatbot tests the ForceRefreshChatbot handler.
func TestAdminChatbotHandlers_ForceRefreshChatbot(t *testing.T) {
	t.Parallel()

	t.Run("success without queue", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return &repository.AdminChatbot{
					ID:   id,
					Name: "Test Chatbot",
				}, nil
			},
			DeleteVectorsFunc: func(ctx context.Context, chatbotID string) error {
				return nil
			},
			ResetSourcesFunc: func(ctx context.Context, chatbotID string) (int64, error) {
				return 3, nil
			},
			GetSourceIDsFunc: func(ctx context.Context, chatbotID string) ([]string, error) {
				return []string{"source-1", "source-2", "source-3"}, nil
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/chatbots/bot-123/force-refresh", nil)
		req.SetPathValue("id", "bot-123")
		w := httptest.NewRecorder()

		h.ForceRefreshChatbot(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var response map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response["status"] != "refreshing" {
			t.Errorf("expected status 'refreshing', got %v", response["status"])
		}
		if int(response["sources_reset"].(float64)) != 3 {
			t.Errorf("expected sources_reset 3, got %v", response["sources_reset"])
		}
		if int(response["sources_queued"].(float64)) != 0 {
			t.Errorf("expected sources_queued 0 (no queue), got %v", response["sources_queued"])
		}
	})

	t.Run("missing id", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{}
		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/chatbots//force-refresh", nil)
		w := httptest.NewRecorder()

		h.ForceRefreshChatbot(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("chatbot not found", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return nil, sql.ErrNoRows
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/chatbots/nonexistent/force-refresh", nil)
		req.SetPathValue("id", "nonexistent")
		w := httptest.NewRecorder()

		h.ForceRefreshChatbot(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("delete vectors error", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return &repository.AdminChatbot{
					ID:   id,
					Name: "Test Chatbot",
				}, nil
			},
			DeleteVectorsFunc: func(ctx context.Context, chatbotID string) error {
				return errors.New("vector delete error")
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/chatbots/bot-123/force-refresh", nil)
		req.SetPathValue("id", "bot-123")
		w := httptest.NewRecorder()

		h.ForceRefreshChatbot(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("reset sources error", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return &repository.AdminChatbot{
					ID:   id,
					Name: "Test Chatbot",
				}, nil
			},
			DeleteVectorsFunc: func(ctx context.Context, chatbotID string) error {
				return nil
			},
			ResetSourcesFunc: func(ctx context.Context, chatbotID string) (int64, error) {
				return 0, errors.New("reset error")
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/chatbots/bot-123/force-refresh", nil)
		req.SetPathValue("id", "bot-123")
		w := httptest.NewRecorder()

		h.ForceRefreshChatbot(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("get source IDs error", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return &repository.AdminChatbot{
					ID:   id,
					Name: "Test Chatbot",
				}, nil
			},
			DeleteVectorsFunc: func(ctx context.Context, chatbotID string) error {
				return nil
			},
			ResetSourcesFunc: func(ctx context.Context, chatbotID string) (int64, error) {
				return 3, nil
			},
			GetSourceIDsFunc: func(ctx context.Context, chatbotID string) ([]string, error) {
				return nil, errors.New("get source IDs error")
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/chatbots/bot-123/force-refresh", nil)
		req.SetPathValue("id", "bot-123")
		w := httptest.NewRecorder()

		h.ForceRefreshChatbot(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("verifies repository calls", func(t *testing.T) {
		t.Parallel()

		mockRepo := &mockAdminChatbotRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*repository.AdminChatbot, error) {
				return &repository.AdminChatbot{
					ID:   id,
					Name: "Test Chatbot",
				}, nil
			},
			DeleteVectorsFunc: func(ctx context.Context, chatbotID string) error {
				return nil
			},
			ResetSourcesFunc: func(ctx context.Context, chatbotID string) (int64, error) {
				return 3, nil
			},
			GetSourceIDsFunc: func(ctx context.Context, chatbotID string) ([]string, error) {
				return []string{"source-1", "source-2", "source-3"}, nil
			},
		}

		h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/chatbots/bot-123/force-refresh", nil)
		req.SetPathValue("id", "bot-123")
		w := httptest.NewRecorder()

		h.ForceRefreshChatbot(w, req)

		// Verify repository calls
		if len(mockRepo.calls.GetByID) != 1 {
			t.Errorf("expected 1 GetByID call, got %d", len(mockRepo.calls.GetByID))
		}
		if mockRepo.calls.GetByID[0] != "bot-123" {
			t.Errorf("expected GetByID called with 'bot-123', got %s", mockRepo.calls.GetByID[0])
		}
		if len(mockRepo.calls.DeleteVectors) != 1 {
			t.Errorf("expected 1 DeleteVectors call, got %d", len(mockRepo.calls.DeleteVectors))
		}
		if len(mockRepo.calls.ResetSources) != 1 {
			t.Errorf("expected 1 ResetSources call, got %d", len(mockRepo.calls.ResetSources))
		}
		if len(mockRepo.calls.GetSourceIDs) != 1 {
			t.Errorf("expected 1 GetSourceIDs call, got %d", len(mockRepo.calls.GetSourceIDs))
		}
	})
}

// TestNewAdminChatbotHandlers tests the constructor.
func TestNewAdminChatbotHandlers(t *testing.T) {
	t.Parallel()

	mockRepo := &mockAdminChatbotRepo{}

	h := NewAdminChatbotHandlers(mockRepo, nil, nil, nil)

	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.AdminChatbotRepo != mockRepo {
		t.Error("expected AdminChatbotRepo to be set")
	}
	if h.AdminService != nil {
		t.Error("expected AdminService to be nil")
	}
	if h.RagService != nil {
		t.Error("expected RagService to be nil")
	}
	if h.Queue != nil {
		t.Error("expected Queue to be nil")
	}
}
