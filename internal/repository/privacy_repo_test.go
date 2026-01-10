package repository

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestMockPrivacyRepo_InterfaceCompliance ensures MockPrivacyRepo implements PrivacyRepository
func TestMockPrivacyRepo_InterfaceCompliance(t *testing.T) {
	var _ PrivacyRepository = (*MockPrivacyRepo)(nil)
}

// TestPostgresPrivacyRepo_InterfaceCompliance ensures PostgresPrivacyRepo implements PrivacyRepository
func TestPostgresPrivacyRepo_InterfaceCompliance(t *testing.T) {
	var _ PrivacyRepository = (*PostgresPrivacyRepo)(nil)
}

// TestNewMockPrivacyRepo verifies that NewMockPrivacyRepo creates a valid mock
func TestNewMockPrivacyRepo(t *testing.T) {
	mock := NewMockPrivacyRepo()
	if mock == nil {
		t.Fatal("NewMockPrivacyRepo returned nil")
	}
}

// TestNewPostgresPrivacyRepo verifies that NewPostgresPrivacyRepo creates a valid repo
func TestNewPostgresPrivacyRepo(t *testing.T) {
	repo := NewPostgresPrivacyRepo(nil)
	if repo == nil {
		t.Fatal("NewPostgresPrivacyRepo returned nil")
	}
}

func TestMockPrivacyRepo_GetUserDataForExport(t *testing.T) {
	t.Run("default returns empty export", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		result, err := mock.GetUserDataForExport(context.Background(), "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetUserDataForExport) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetUserDataForExport))
		}
		if mock.Calls.GetUserDataForExport[0].UserID != "user-123" {
			t.Errorf("expected call with UserID 'user-123', got: %s", mock.Calls.GetUserDataForExport[0].UserID)
		}
	})

	t.Run("custom function returns full export", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		expectedExport := &UserDataExport{
			User: ExportUser{
				ID:    "user-123",
				Email: "test@example.com",
			},
			Organizations: []ExportOrg{
				{ID: "org-1", Name: "Test Org"},
			},
			Chatbots: []ExportChatbot{
				{ID: "bot-1", Name: "Test Bot"},
			},
			Conversations: []ExportConv{
				{ID: "conv-1", ChatbotID: "bot-1"},
			},
			Messages: []ExportMessage{
				{ID: "msg-1", ConversationID: "conv-1", Role: "user", Content: "Hello"},
			},
			ExportedAt: time.Now(),
		}
		mock.GetUserDataForExportFunc = func(ctx context.Context, userID string) (*UserDataExport, error) {
			if userID == "user-123" {
				return expectedExport, nil
			}
			return nil, nil
		}

		result, err := mock.GetUserDataForExport(context.Background(), "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected export data, got nil")
		}
		if result.User.ID != "user-123" {
			t.Errorf("expected user ID 'user-123', got: %s", result.User.ID)
		}
		if len(result.Organizations) != 1 {
			t.Errorf("expected 1 organization, got: %d", len(result.Organizations))
		}
		if len(result.Chatbots) != 1 {
			t.Errorf("expected 1 chatbot, got: %d", len(result.Chatbots))
		}
		if len(result.Conversations) != 1 {
			t.Errorf("expected 1 conversation, got: %d", len(result.Conversations))
		}
		if len(result.Messages) != 1 {
			t.Errorf("expected 1 message, got: %d", len(result.Messages))
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		expectedErr := errors.New("database error")
		mock.GetUserDataForExportFunc = func(ctx context.Context, userID string) (*UserDataExport, error) {
			return nil, expectedErr
		}

		result, err := mock.GetUserDataForExport(context.Background(), "user-123")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
	})
}

func TestMockPrivacyRepo_CompletePrivacyExportRequest(t *testing.T) {
	t.Run("default returns nil error", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		err := mock.CompletePrivacyExportRequest(context.Background(), "req-123", "admin-456", "https://example.com/export.json", expiresAt)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.CompletePrivacyExportRequest) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.CompletePrivacyExportRequest))
		}
		call := mock.Calls.CompletePrivacyExportRequest[0]
		if call.RequestID != "req-123" {
			t.Errorf("expected RequestID 'req-123', got: %s", call.RequestID)
		}
		if call.AdminID != "admin-456" {
			t.Errorf("expected AdminID 'admin-456', got: %s", call.AdminID)
		}
		if call.ExportURL != "https://example.com/export.json" {
			t.Errorf("expected ExportURL 'https://example.com/export.json', got: %s", call.ExportURL)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		expectedErr := errors.New("update failed")
		mock.CompletePrivacyExportRequestFunc = func(ctx context.Context, requestID, adminID, exportURL string, expiresAt time.Time) error {
			return expectedErr
		}

		err := mock.CompletePrivacyExportRequest(context.Background(), "req-123", "admin-456", "url", time.Now())
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})

	t.Run("custom function succeeds", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		mock.CompletePrivacyExportRequestFunc = func(ctx context.Context, requestID, adminID, exportURL string, expiresAt time.Time) error {
			if requestID != "req-valid" {
				return errors.New("invalid request")
			}
			return nil
		}

		err := mock.CompletePrivacyExportRequest(context.Background(), "req-valid", "admin-1", "url", time.Now())
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})
}

func TestMockPrivacyRepo_CreatePrivacyRequest(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		req := PrivacyRequest{
			UserEmail:   "test@example.com",
			RequestType: "export",
			Status:      "pending",
		}
		result, err := mock.CreatePrivacyRequest(context.Background(), req)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.CreatePrivacyRequest) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.CreatePrivacyRequest))
		}
	})

	t.Run("custom function returns request", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		mock.CreatePrivacyRequestFunc = func(ctx context.Context, req PrivacyRequest) (*PrivacyRequest, error) {
			req.ID = "generated-id"
			req.CreatedAt = time.Now()
			return &req, nil
		}

		req := PrivacyRequest{
			UserEmail:   "test@example.com",
			RequestType: "export",
			Status:      "pending",
		}
		result, err := mock.CreatePrivacyRequest(context.Background(), req)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected request, got nil")
		}
		if result.ID != "generated-id" {
			t.Errorf("expected ID 'generated-id', got: %s", result.ID)
		}
	})
}

func TestMockPrivacyRepo_UpdatePrivacyRequestStatus(t *testing.T) {
	t.Run("default returns nil error", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		err := mock.UpdatePrivacyRequestStatus(context.Background(), "req-123", "completed", "admin-456", nil)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.UpdatePrivacyRequestStatus) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.UpdatePrivacyRequestStatus))
		}
	})

	t.Run("with denial reason", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		denialReason := "Invalid request"
		err := mock.UpdatePrivacyRequestStatus(context.Background(), "req-123", "denied", "admin-456", &denialReason)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		call := mock.Calls.UpdatePrivacyRequestStatus[0]
		if call.DenialReason == nil || *call.DenialReason != "Invalid request" {
			t.Errorf("expected denial reason 'Invalid request', got: %v", call.DenialReason)
		}
	})
}

func TestMockPrivacyRepo_HasActivePrivacyRequest(t *testing.T) {
	t.Run("default returns false", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		result, err := mock.HasActivePrivacyRequest(context.Background(), "user-123", "export")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != false {
			t.Errorf("expected false, got: %v", result)
		}
	})

	t.Run("custom function returns true", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		mock.HasActivePrivacyRequestFunc = func(ctx context.Context, userID, requestType string) (bool, error) {
			return userID == "user-with-request", nil
		}

		result, err := mock.HasActivePrivacyRequest(context.Background(), "user-with-request", "export")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != true {
			t.Errorf("expected true, got: %v", result)
		}
	})
}

func TestMockPrivacyRepo_Reset(t *testing.T) {
	mock := NewMockPrivacyRepo()

	// Make some calls
	_, _ = mock.GetUserDataForExport(context.Background(), "user-1")
	_ = mock.CompletePrivacyExportRequest(context.Background(), "req-1", "admin-1", "url", time.Now())
	_, _ = mock.CreatePrivacyRequest(context.Background(), PrivacyRequest{})
	_ = mock.UpdatePrivacyRequestStatus(context.Background(), "req-1", "completed", "admin-1", nil)

	// Verify calls were recorded
	if len(mock.Calls.GetUserDataForExport) != 1 {
		t.Errorf("expected 1 GetUserDataForExport call, got: %d", len(mock.Calls.GetUserDataForExport))
	}
	if len(mock.Calls.CompletePrivacyExportRequest) != 1 {
		t.Errorf("expected 1 CompletePrivacyExportRequest call, got: %d", len(mock.Calls.CompletePrivacyExportRequest))
	}

	// Reset
	mock.Reset()

	// Verify calls were cleared
	if len(mock.Calls.GetUserDataForExport) != 0 {
		t.Errorf("expected 0 GetUserDataForExport calls after reset, got: %d", len(mock.Calls.GetUserDataForExport))
	}
	if len(mock.Calls.CompletePrivacyExportRequest) != 0 {
		t.Errorf("expected 0 CompletePrivacyExportRequest calls after reset, got: %d", len(mock.Calls.CompletePrivacyExportRequest))
	}
	if len(mock.Calls.CreatePrivacyRequest) != 0 {
		t.Errorf("expected 0 CreatePrivacyRequest calls after reset, got: %d", len(mock.Calls.CreatePrivacyRequest))
	}
	if len(mock.Calls.UpdatePrivacyRequestStatus) != 0 {
		t.Errorf("expected 0 UpdatePrivacyRequestStatus calls after reset, got: %d", len(mock.Calls.UpdatePrivacyRequestStatus))
	}
}

func TestMockPrivacyRepo_GetUserConsents(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		result, err := mock.GetUserConsents(context.Background(), "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
	})

	t.Run("custom function returns consents", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		mock.GetUserConsentsFunc = func(ctx context.Context, userID string) ([]UserConsent, error) {
			return []UserConsent{
				{ID: "consent-1", UserID: userID, ConsentType: "marketing", Granted: true},
				{ID: "consent-2", UserID: userID, ConsentType: "analytics", Granted: false},
			}, nil
		}

		result, err := mock.GetUserConsents(context.Background(), "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 consents, got: %d", len(result))
		}
	})
}

func TestMockPrivacyRepo_AnonymizeUserData(t *testing.T) {
	t.Run("default returns nil error", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		err := mock.AnonymizeUserData(context.Background(), "user-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.AnonymizeUserData) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.AnonymizeUserData))
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		expectedErr := errors.New("anonymization failed")
		mock.AnonymizeUserDataFunc = func(ctx context.Context, userID string) error {
			return expectedErr
		}

		err := mock.AnonymizeUserData(context.Background(), "user-123")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

func TestMockPrivacyRepo_DeletePrivacyRequest(t *testing.T) {
	t.Run("default returns nil error", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		err := mock.DeletePrivacyRequest(context.Background(), "request-123")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if len(mock.Calls.DeletePrivacyRequest) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.DeletePrivacyRequest))
		}
		if mock.Calls.DeletePrivacyRequest[0].RequestID != "request-123" {
			t.Errorf("expected call with RequestID 'request-123', got: %s", mock.Calls.DeletePrivacyRequest[0].RequestID)
		}
	})

	t.Run("custom function returns error", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		expectedErr := errors.New("delete failed")
		mock.DeletePrivacyRequestFunc = func(ctx context.Context, requestID string) error {
			return expectedErr
		}

		err := mock.DeletePrivacyRequest(context.Background(), "request-123")
		if err != expectedErr {
			t.Errorf("expected error %v, got: %v", expectedErr, err)
		}
	})
}

func TestMockPrivacyRepo_GetLastCompletedRequestDate(t *testing.T) {
	t.Run("default returns nil", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		result, err := mock.GetLastCompletedRequestDate(context.Background(), "user-123", "export")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
		if len(mock.Calls.GetLastCompletedRequestDate) != 1 {
			t.Errorf("expected 1 call recorded, got: %d", len(mock.Calls.GetLastCompletedRequestDate))
		}
		if mock.Calls.GetLastCompletedRequestDate[0].UserID != "user-123" {
			t.Errorf("expected call with UserID 'user-123', got: %s", mock.Calls.GetLastCompletedRequestDate[0].UserID)
		}
		if mock.Calls.GetLastCompletedRequestDate[0].RequestType != "export" {
			t.Errorf("expected call with RequestType 'export', got: %s", mock.Calls.GetLastCompletedRequestDate[0].RequestType)
		}
	})

	t.Run("custom function returns completed date", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		expectedTime := time.Now().Add(-12 * time.Hour)
		mock.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			if userID == "user-123" && requestType == "export" {
				return &expectedTime, nil
			}
			return nil, nil
		}

		result, err := mock.GetLastCompletedRequestDate(context.Background(), "user-123", "export")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result == nil {
			t.Fatal("expected time result, got nil")
		}
		if result.Sub(expectedTime) > time.Second {
			t.Errorf("expected time close to %v, got: %v", expectedTime, *result)
		}
	})

	t.Run("custom function returns nil for no completed requests", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		mock.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			return nil, nil
		}

		result, err := mock.GetLastCompletedRequestDate(context.Background(), "user-456", "correction")
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got: %v", result)
		}
	})

	t.Run("filters by request type", func(t *testing.T) {
		mock := NewMockPrivacyRepo()
		exportTime := time.Now().Add(-12 * time.Hour)
		correctionTime := time.Now().Add(-25 * time.Hour)

		mock.GetLastCompletedRequestDateFunc = func(ctx context.Context, userID, requestType string) (*time.Time, error) {
			if userID == "user-123" {
				if requestType == "export" {
					return &exportTime, nil
				}
				if requestType == "correction" {
					return &correctionTime, nil
				}
			}
			return nil, nil
		}

		exportResult, err := mock.GetLastCompletedRequestDate(context.Background(), "user-123", "export")
		if err != nil {
			t.Errorf("expected no error for export, got: %v", err)
		}
		if exportResult == nil || exportResult.Sub(exportTime) > time.Second {
			t.Errorf("expected export time close to %v, got: %v", exportTime, exportResult)
		}

		correctionResult, err := mock.GetLastCompletedRequestDate(context.Background(), "user-123", "correction")
		if err != nil {
			t.Errorf("expected no error for correction, got: %v", err)
		}
		if correctionResult == nil || correctionResult.Sub(correctionTime) > time.Second {
			t.Errorf("expected correction time close to %v, got: %v", correctionTime, correctionResult)
		}
	})
}
