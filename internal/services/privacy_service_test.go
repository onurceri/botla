package services

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/storage"
)

func TestPrivacyService_RequestExport_RateLimit(t *testing.T) {
	t.Run("allows export when no previous completed export exists", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		// Mock repository to return no completed exports
		mockRepo.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return nil, nil
		}

		mockRepo.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return false, nil
		}

		mockRepo.CreatePrivacyRequestFunc = func(ctx context.Context, req repository.PrivacyRequest) (*repository.PrivacyRequest, error) {
			return &repository.PrivacyRequest{
				ID:          "req-123",
				RequestType: "export",
				Status:      "pending",
			}, nil
		}

		result, err := service.RequestExport(context.Background(), "user-123", "test@example.com", "")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if result.ID != "req-123" {
			t.Errorf("expected request ID 'req-123', got: %s", result.ID)
		}
	})

	t.Run("blocks export when last completed export was within 24 hours", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		// Mock repository to return a completed export from 12 hours ago
		completedTime := time.Now().Add(-12 * time.Hour)
		mockRepo.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return &completedTime, nil
		}

		mockRepo.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return false, nil
		}

		_, err := service.RequestExport(context.Background(), "user-123", "test@example.com", "")
		if err != ErrRateLimitExceeded {
			t.Errorf("expected ErrRateLimitExceeded, got: %v", err)
		}
	})

	t.Run("allows export when last completed export was more than 24 hours ago", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		// Mock repository to return a completed export from 25 hours ago
		completedTime := time.Now().Add(-25 * time.Hour)
		mockRepo.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return &completedTime, nil
		}

		mockRepo.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return false, nil
		}

		mockRepo.CreatePrivacyRequestFunc = func(ctx context.Context, req repository.PrivacyRequest) (*repository.PrivacyRequest, error) {
			return &repository.PrivacyRequest{
				ID:          "req-456",
				RequestType: "export",
				Status:      "pending",
			}, nil
		}

		result, err := service.RequestExport(context.Background(), "user-123", "test@example.com", "")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if result.ID != "req-456" {
			t.Errorf("expected request ID 'req-456', got: %s", result.ID)
		}
	})

	t.Run("returns error when repository check fails", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		mockRepo.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return nil, errors.New("database error")
		}

		mockRepo.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return false, nil
		}

		_, err := service.RequestExport(context.Background(), "user-123", "test@example.com", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestPrivacyService_DeleteMyPrivacyRequest(t *testing.T) {
	t.Run("successfully deletes own request", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		mockRepo.GetPrivacyRequestFunc = func(ctx context.Context, requestID string) (*repository.PrivacyRequest, error) {
			userID := "user-123"
			return &repository.PrivacyRequest{
				ID:          requestID,
				UserID:      &userID,
				RequestType: "export",
				Status:      "completed",
			}, nil
		}

		mockRepo.DeletePrivacyRequestFunc = func(ctx context.Context, requestID string) error {
			return nil
		}

		err := service.DeleteMyPrivacyRequest(context.Background(), "req-123", "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("fails when request belongs to different user", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		otherUserID := "user-456"
		mockRepo.GetPrivacyRequestFunc = func(ctx context.Context, requestID string) (*repository.PrivacyRequest, error) {
			return &repository.PrivacyRequest{
				ID:          requestID,
				UserID:      &otherUserID,
				RequestType: "export",
				Status:      "completed",
			}, nil
		}

		err := service.DeleteMyPrivacyRequest(context.Background(), "req-123", "user-123")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("deletes export file from storage", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		exportURL := "exports/user-123/req-123.json"
		mockRepo.GetPrivacyRequestFunc = func(ctx context.Context, requestID string) (*repository.PrivacyRequest, error) {
			userID := "user-123"
			return &repository.PrivacyRequest{
				ID:          requestID,
				UserID:      &userID,
				RequestType: "export",
				Status:      "completed",
				ExportURL:   &exportURL,
			}, nil
		}

		mockRepo.DeletePrivacyRequestFunc = func(ctx context.Context, requestID string) error {
			return nil
		}

		fileDeleted := false
		mockStorage.deleteFileFunc = func(ctx context.Context, key string) error {
			fileDeleted = true
			return nil
		}

		err := service.DeleteMyPrivacyRequest(context.Background(), "req-123", "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if !fileDeleted {
			t.Error("expected file to be deleted from storage")
		}
	})

	t.Run("continues with DB deletion even if file deletion fails", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		exportURL := "exports/user-123/req-123.json"
		mockRepo.GetPrivacyRequestFunc = func(ctx context.Context, requestID string) (*repository.PrivacyRequest, error) {
			userID := "user-123"
			return &repository.PrivacyRequest{
				ID:          requestID,
				UserID:      &userID,
				RequestType: "export",
				Status:      "completed",
				ExportURL:   &exportURL,
			}, nil
		}

		dbDeleted := false
		mockRepo.DeletePrivacyRequestFunc = func(ctx context.Context, requestID string) error {
			dbDeleted = true
			return nil
		}

		mockStorage.deleteFileFunc = func(ctx context.Context, key string) error {
			return errors.New("storage error")
		}

		err := service.DeleteMyPrivacyRequest(context.Background(), "req-123", "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if !dbDeleted {
			t.Error("expected DB record to be deleted even if file deletion failed")
		}
	})
}

// mockStorageService is a minimal mock for testing
type mockStorageService struct {
	deleteFileFunc func(ctx context.Context, key string) error
}

func (m *mockStorageService) UploadFile(ctx context.Context, key string, reader io.Reader) (string, error) {
	return "", nil
}

func (m *mockStorageService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func TestPrivacyService_RequestCorrection_RateLimit(t *testing.T) {
	t.Run("allows correction when no previous completed correction exists", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		mockRepo.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return nil, nil
		}

		mockRepo.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return false, nil
		}

		mockRepo.CreatePrivacyRequestFunc = func(ctx context.Context, req repository.PrivacyRequest) (*repository.PrivacyRequest, error) {
			return &repository.PrivacyRequest{
				ID:          "req-123",
				RequestType: "correction",
				Status:      "pending",
			}, nil
		}

		result, err := service.RequestCorrection(context.Background(), "user-123", "test@example.com", "Please correct my data")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if result.ID != "req-123" {
			t.Errorf("expected request ID 'req-123', got: %s", result.ID)
		}
	})

	t.Run("blocks correction when last completed correction was within 24 hours", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		completedTime := time.Now().Add(-12 * time.Hour)
		mockRepo.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return &completedTime, nil
		}

		mockRepo.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return false, nil
		}

		_, err := service.RequestCorrection(context.Background(), "user-123", "test@example.com", "Please correct my data")
		if err != ErrRateLimitExceeded {
			t.Errorf("expected ErrRateLimitExceeded, got: %v", err)
		}
	})

	t.Run("allows correction when last completed correction was more than 24 hours ago", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		completedTime := time.Now().Add(-25 * time.Hour)
		mockRepo.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return &completedTime, nil
		}

		mockRepo.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return false, nil
		}

		mockRepo.CreatePrivacyRequestFunc = func(ctx context.Context, req repository.PrivacyRequest) (*repository.PrivacyRequest, error) {
			return &repository.PrivacyRequest{
				ID:          "req-456",
				RequestType: "correction",
				Status:      "pending",
			}, nil
		}

		result, err := service.RequestCorrection(context.Background(), "user-123", "test@example.com", "Please correct my data")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if result.ID != "req-456" {
			t.Errorf("expected request ID 'req-456', got: %s", result.ID)
		}
	})

	t.Run("blocks correction when active request exists", func(t *testing.T) {
		mockRepo := repository.NewMockPrivacyRepo()
		mockStorage := &mockStorageService{}
		log := logger.New("")
		service := NewPrivacyService(mockRepo, log, mockStorage)

		mockRepo.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return nil, nil
		}

		mockRepo.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return true, nil
		}

		_, err := service.RequestCorrection(context.Background(), "user-123", "test@example.com", "Please correct my data")
		if err != ErrActiveRequestExists {
			t.Errorf("expected ErrActiveRequestExists, got: %v", err)
		}
	})
}

func (m *mockStorageService) DeleteFile(ctx context.Context, key string) error {
	if m.deleteFileFunc != nil {
		return m.deleteFileFunc(ctx, key)
	}
	return nil
}

func (m *mockStorageService) GetFileURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	return "", nil
}

func (m *mockStorageService) GetSignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return "", nil
}

// Ensure mockStorageService implements storage.StorageService
var _ storage.StorageService = (*mockStorageService)(nil)
