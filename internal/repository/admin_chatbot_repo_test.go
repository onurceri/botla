package repository

import (
	"context"
	"errors"
	"testing"
)

// TestMockAdminChatbotRepo_ListChatbots tests the ListChatbots mock method.
func TestMockAdminChatbotRepo_ListChatbots(t *testing.T) {
	t.Parallel()

	t.Run("returns empty list by default", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()

		filter := AdminChatbotFilter{}
		result, total, err := mock.ListChatbots(context.Background(), filter, 10, 0)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
		if total != 0 {
			t.Fatalf("expected 0 total, got %d", total)
		}

		// Verify call was recorded
		if len(mock.Calls.ListChatbots) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.Calls.ListChatbots))
		}
		if mock.Calls.ListChatbots[0].Limit != 10 {
			t.Errorf("expected limit 10, got %d", mock.Calls.ListChatbots[0].Limit)
		}
		if mock.Calls.ListChatbots[0].Offset != 0 {
			t.Errorf("expected offset 0, got %d", mock.Calls.ListChatbots[0].Offset)
		}
	})

	t.Run("uses custom function when provided", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()
		expectedChatbots := []AdminChatbot{
			{ID: "chatbot-1", Name: "Bot 1"},
			{ID: "chatbot-2", Name: "Bot 2"},
		}
		mock.ListChatbotsFunc = func(ctx context.Context, filter AdminChatbotFilter, limit, offset int) ([]AdminChatbot, int, error) {
			return expectedChatbots, 2, nil
		}

		result, total, err := mock.ListChatbots(context.Background(), AdminChatbotFilter{}, 10, 0)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("expected 2 chatbots, got %d", len(result))
		}
		if total != 2 {
			t.Fatalf("expected total 2, got %d", total)
		}
	})

	t.Run("records filter parameters", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()
		name := "TestBot"
		orgID := "org-123"
		filter := AdminChatbotFilter{
			Name:           &name,
			OrganizationID: &orgID,
		}

		_, _, _ = mock.ListChatbots(context.Background(), filter, 20, 5)

		if len(mock.Calls.ListChatbots) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.Calls.ListChatbots))
		}
		call := mock.Calls.ListChatbots[0]
		if call.Filter.Name == nil || *call.Filter.Name != "TestBot" {
			t.Errorf("expected filter.Name to be 'TestBot'")
		}
		if call.Filter.OrganizationID == nil || *call.Filter.OrganizationID != "org-123" {
			t.Errorf("expected filter.OrganizationID to be 'org-123'")
		}
		if call.Limit != 20 {
			t.Errorf("expected limit 20, got %d", call.Limit)
		}
		if call.Offset != 5 {
			t.Errorf("expected offset 5, got %d", call.Offset)
		}
	})
}

// TestMockAdminChatbotRepo_GetByID tests the GetByID mock method.
func TestMockAdminChatbotRepo_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("returns nil by default", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()

		result, err := mock.GetByID(context.Background(), "chatbot-123")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}

		// Verify call was recorded
		if len(mock.Calls.GetByID) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.Calls.GetByID))
		}
		if mock.Calls.GetByID[0].ID != "chatbot-123" {
			t.Errorf("expected ID 'chatbot-123', got %s", mock.Calls.GetByID[0].ID)
		}
	})

	t.Run("uses custom function when provided", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()
		expectedChatbot := &AdminChatbot{
			ID:         "chatbot-456",
			Name:       "Test Chatbot",
			OwnerEmail: "test@example.com",
		}
		mock.GetByIDFunc = func(ctx context.Context, id string) (*AdminChatbot, error) {
			if id == "chatbot-456" {
				return expectedChatbot, nil
			}
			return nil, errors.New("not found")
		}

		result, err := mock.GetByID(context.Background(), "chatbot-456")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.ID != "chatbot-456" {
			t.Errorf("expected ID 'chatbot-456', got %s", result.ID)
		}
		if result.Name != "Test Chatbot" {
			t.Errorf("expected name 'Test Chatbot', got %s", result.Name)
		}
	})

	t.Run("returns error when function returns error", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()
		expectedErr := errors.New("database error")
		mock.GetByIDFunc = func(ctx context.Context, id string) (*AdminChatbot, error) {
			return nil, expectedErr
		}

		result, err := mock.GetByID(context.Background(), "chatbot-789")

		if err != expectedErr {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}
	})
}

// TestMockAdminChatbotRepo_ResetSources tests the ResetSources mock method.
func TestMockAdminChatbotRepo_ResetSources(t *testing.T) {
	t.Parallel()

	t.Run("returns 0 by default", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()

		count, err := mock.ResetSources(context.Background(), "chatbot-123")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if count != 0 {
			t.Fatalf("expected count 0, got %d", count)
		}

		// Verify call was recorded
		if len(mock.Calls.ResetSources) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.Calls.ResetSources))
		}
		if mock.Calls.ResetSources[0].ChatbotID != "chatbot-123" {
			t.Errorf("expected ChatbotID 'chatbot-123', got %s", mock.Calls.ResetSources[0].ChatbotID)
		}
	})

	t.Run("uses custom function when provided", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()
		mock.ResetSourcesFunc = func(ctx context.Context, chatbotID string) (int64, error) {
			return 5, nil
		}

		count, err := mock.ResetSources(context.Background(), "chatbot-456")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if count != 5 {
			t.Fatalf("expected count 5, got %d", count)
		}
	})
}

// TestMockAdminChatbotRepo_GetSourceIDs tests the GetSourceIDs mock method.
func TestMockAdminChatbotRepo_GetSourceIDs(t *testing.T) {
	t.Parallel()

	t.Run("returns nil by default", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()

		result, err := mock.GetSourceIDs(context.Background(), "chatbot-123")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != nil {
			t.Fatalf("expected nil result, got %v", result)
		}

		// Verify call was recorded
		if len(mock.Calls.GetSourceIDs) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.Calls.GetSourceIDs))
		}
	})

	t.Run("uses custom function when provided", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()
		expectedIDs := []string{"source-1", "source-2", "source-3"}
		mock.GetSourceIDsFunc = func(ctx context.Context, chatbotID string) ([]string, error) {
			return expectedIDs, nil
		}

		result, err := mock.GetSourceIDs(context.Background(), "chatbot-456")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 3 {
			t.Fatalf("expected 3 source IDs, got %d", len(result))
		}
	})
}

// TestMockAdminChatbotRepo_DeleteVectors tests the DeleteVectors mock method.
func TestMockAdminChatbotRepo_DeleteVectors(t *testing.T) {
	t.Parallel()

	t.Run("returns nil by default", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()

		err := mock.DeleteVectors(context.Background(), "chatbot-123")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify call was recorded
		if len(mock.Calls.DeleteVectors) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.Calls.DeleteVectors))
		}
		if mock.Calls.DeleteVectors[0].ChatbotID != "chatbot-123" {
			t.Errorf("expected ChatbotID 'chatbot-123', got %s", mock.Calls.DeleteVectors[0].ChatbotID)
		}
	})

	t.Run("uses custom function when provided", func(t *testing.T) {
		t.Parallel()
		mock := NewMockAdminChatbotRepo()
		expectedErr := errors.New("delete error")
		mock.DeleteVectorsFunc = func(ctx context.Context, chatbotID string) error {
			return expectedErr
		}

		err := mock.DeleteVectors(context.Background(), "chatbot-456")

		if err != expectedErr {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})
}

// TestMockAdminChatbotRepo_Reset tests the Reset method.
func TestMockAdminChatbotRepo_Reset(t *testing.T) {
	t.Parallel()

	mock := NewMockAdminChatbotRepo()

	// Make some calls
	_, _, _ = mock.ListChatbots(context.Background(), AdminChatbotFilter{}, 10, 0)
	_, _ = mock.GetByID(context.Background(), "test-id")
	_, _ = mock.ResetSources(context.Background(), "test-id")
	_, _ = mock.GetSourceIDs(context.Background(), "test-id")
	_ = mock.DeleteVectors(context.Background(), "test-id")

	// Verify calls were recorded
	if len(mock.Calls.ListChatbots) == 0 {
		t.Error("expected ListChatbots calls to be recorded")
	}
	if len(mock.Calls.GetByID) == 0 {
		t.Error("expected GetByID calls to be recorded")
	}
	if len(mock.Calls.ResetSources) == 0 {
		t.Error("expected ResetSources calls to be recorded")
	}
	if len(mock.Calls.GetSourceIDs) == 0 {
		t.Error("expected GetSourceIDs calls to be recorded")
	}
	if len(mock.Calls.DeleteVectors) == 0 {
		t.Error("expected DeleteVectors calls to be recorded")
	}

	// Reset
	mock.Reset()

	// Verify all calls are cleared
	if len(mock.Calls.ListChatbots) != 0 {
		t.Errorf("expected no ListChatbots calls, got %d", len(mock.Calls.ListChatbots))
	}
	if len(mock.Calls.GetByID) != 0 {
		t.Errorf("expected no GetByID calls, got %d", len(mock.Calls.GetByID))
	}
	if len(mock.Calls.ResetSources) != 0 {
		t.Errorf("expected no ResetSources calls, got %d", len(mock.Calls.ResetSources))
	}
	if len(mock.Calls.GetSourceIDs) != 0 {
		t.Errorf("expected no GetSourceIDs calls, got %d", len(mock.Calls.GetSourceIDs))
	}
	if len(mock.Calls.DeleteVectors) != 0 {
		t.Errorf("expected no DeleteVectors calls, got %d", len(mock.Calls.DeleteVectors))
	}
}

// TestAdminChatbotRepository_InterfaceCompliance verifies compile-time interface compliance.
func TestAdminChatbotRepository_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	// These are compile-time checks but we include them in tests for documentation
	var _ AdminChatbotRepository = (*MockAdminChatbotRepo)(nil)
	var _ AdminChatbotRepository = (*PostgresAdminChatbotRepo)(nil)
}

// TestAdminChatbotFilter tests the AdminChatbotFilter struct.
func TestAdminChatbotFilter(t *testing.T) {
	t.Parallel()

	t.Run("empty filter", func(t *testing.T) {
		t.Parallel()
		filter := AdminChatbotFilter{}

		if filter.Name != nil {
			t.Error("expected Name to be nil")
		}
		if filter.OrganizationID != nil {
			t.Error("expected OrganizationID to be nil")
		}
		if filter.OwnerID != nil {
			t.Error("expected OwnerID to be nil")
		}
	})

	t.Run("filter with values", func(t *testing.T) {
		t.Parallel()
		name := "TestBot"
		orgID := "org-123"
		ownerID := "user-456"
		filter := AdminChatbotFilter{
			Name:           &name,
			OrganizationID: &orgID,
			OwnerID:        &ownerID,
		}

		if filter.Name == nil || *filter.Name != "TestBot" {
			t.Error("expected Name to be 'TestBot'")
		}
		if filter.OrganizationID == nil || *filter.OrganizationID != "org-123" {
			t.Error("expected OrganizationID to be 'org-123'")
		}
		if filter.OwnerID == nil || *filter.OwnerID != "user-456" {
			t.Error("expected OwnerID to be 'user-456'")
		}
	})
}

// TestAdminChatbot tests the AdminChatbot struct.
func TestAdminChatbot(t *testing.T) {
	t.Parallel()

	chatbot := AdminChatbot{
		ID:               "chatbot-123",
		Name:             "Test Chatbot",
		OwnerID:          "user-456",
		WorkspaceID:      "workspace-789",
		OrganizationID:   "org-abc",
		OrganizationName: "Test Org",
		OwnerEmail:       "test@example.com",
		SourceCount:      5,
		MessageCount:     100,
		CustomBranding:   []byte(`{"color": "blue"}`),
		CreatedAt:        "2024-01-01T00:00:00Z",
		UpdatedAt:        "2024-01-02T00:00:00Z",
	}

	if chatbot.ID != "chatbot-123" {
		t.Errorf("expected ID 'chatbot-123', got %s", chatbot.ID)
	}
	if chatbot.Name != "Test Chatbot" {
		t.Errorf("expected Name 'Test Chatbot', got %s", chatbot.Name)
	}
	if chatbot.OwnerID != "user-456" {
		t.Errorf("expected OwnerID 'user-456', got %s", chatbot.OwnerID)
	}
	if chatbot.SourceCount != 5 {
		t.Errorf("expected SourceCount 5, got %d", chatbot.SourceCount)
	}
	if chatbot.MessageCount != 100 {
		t.Errorf("expected MessageCount 100, got %d", chatbot.MessageCount)
	}
}
