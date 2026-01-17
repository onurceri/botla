package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/pkg/logger"
)

// MockUserRepo for testing DeletedAccountMiddleware
type MockUserRepo struct {
	GetByIDFunc func(ctx context.Context, id string) (*models.User, error)
}

func (m *MockUserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &models.User{ID: id, Email: "test@example.com"}, nil
}

func TestDeletedAccountMiddleware(t *testing.T) {
	log := logger.New("INFO")

	t.Run("allows access for existing user", func(t *testing.T) {
		mockRepo := &MockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				return &models.User{ID: id, Email: "test@example.com"}, nil
			},
		}

		middleware := DeletedAccountMiddleware(mockRepo, log)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.WithValue(req.Context(), contextKey("userID"), "user-123"))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for existing user, got %d", w.Code)
		}
	})

	t.Run("returns 403 for deleted user", func(t *testing.T) {
		mockRepo := &MockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				return nil, nil // User not found = deleted
			},
		}

		middleware := DeletedAccountMiddleware(mockRepo, log)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.WithValue(req.Context(), contextKey("userID"), "deleted-user-123"))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 for deleted user, got %d", w.Code)
		}

		body := strings.TrimSpace(w.Body.String())
		if body != `{"code":"ERR_ACCOUNT_DELETED","status":403}` {
			t.Errorf("Expected ERR_ACCOUNT_DELETED response, got %s", body)
		}
	})

	t.Run("skips check for unauthenticated request", func(t *testing.T) {
		mockRepo := &MockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				t.Error("GetByID should not be called for unauthenticated request")
				return nil, nil
			},
		}

		middleware := DeletedAccountMiddleware(mockRepo, log)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for unauthenticated request, got %d", w.Code)
		}
	})

	t.Run("continues on database error", func(t *testing.T) {
		mockRepo := &MockUserRepo{
			GetByIDFunc: func(ctx context.Context, id string) (*models.User, error) {
				return nil, sql.ErrConnDone
			},
		}

		middleware := DeletedAccountMiddleware(mockRepo, log)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req = req.WithContext(context.WithValue(req.Context(), contextKey("userID"), "user-123"))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 on DB error (continues), got %d", w.Code)
		}
	})
}
