package repository

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
)

type MockHandoffRepo struct {
	HasActiveHandoffRequestFunc     func(ctx context.Context, conversationID string) (bool, error)
	CreateHandoffRequestFunc        func(ctx context.Context, req *models.HandoffRequest) (string, error)
	GetHandoffRequestsByBotIDFunc   func(ctx context.Context, chatbotID string) ([]*models.HandoffRequest, error)
	GetHandoffRequestByIDFunc       func(ctx context.Context, id string) (*models.HandoffRequest, error)
	UpdateHandoffRequestStatusFunc  func(ctx context.Context, id, status string, assignedTo *string) error
	CountPendingHandoffRequestsFunc func(ctx context.Context, chatbotID string) (int, error)
	ListHandoffMessagesFunc         func(ctx context.Context, conversationID string, limit int) ([]models.Message, error)

	Calls struct {
		HasActiveHandoffRequest     []HasActiveHandoffRequestCall
		CreateHandoffRequest        []CreateHandoffRequestCall
		GetHandoffRequestsByBotID   []GetHandoffRequestsByBotIDCall
		GetHandoffRequestByID       []GetHandoffRequestByIDCall
		UpdateHandoffRequestStatus  []UpdateHandoffRequestStatusCall
		CountPendingHandoffRequests []CountPendingHandoffRequestsCall
		ListHandoffMessages         []ListHandoffMessagesCall
	}
}

type HasActiveHandoffRequestCall struct {
	ConversationID string
}

type CreateHandoffRequestCall struct {
	Req *models.HandoffRequest
}

type GetHandoffRequestsByBotIDCall struct {
	ChatbotID string
}

type GetHandoffRequestByIDCall struct {
	ID string
}

type UpdateHandoffRequestStatusCall struct {
	ID         string
	Status     string
	AssignedTo *string
}

type CountPendingHandoffRequestsCall struct {
	ChatbotID string
}

type ListHandoffMessagesCall struct {
	ConversationID string
	Limit          int
}

var _ HandoffRepository = (*MockHandoffRepo)(nil)

func NewMockHandoffRepo() *MockHandoffRepo {
	return &MockHandoffRepo{}
}

func (m *MockHandoffRepo) HasActiveHandoffRequest(ctx context.Context, conversationID string) (bool, error) {
	m.Calls.HasActiveHandoffRequest = append(m.Calls.HasActiveHandoffRequest, HasActiveHandoffRequestCall{ConversationID: conversationID})
	if m.HasActiveHandoffRequestFunc != nil {
		return m.HasActiveHandoffRequestFunc(ctx, conversationID)
	}
	return false, nil
}

func (m *MockHandoffRepo) CreateHandoffRequest(ctx context.Context, req *models.HandoffRequest) (string, error) {
	m.Calls.CreateHandoffRequest = append(m.Calls.CreateHandoffRequest, CreateHandoffRequestCall{Req: req})
	if m.CreateHandoffRequestFunc != nil {
		return m.CreateHandoffRequestFunc(ctx, req)
	}
	return "", nil
}

func (m *MockHandoffRepo) GetHandoffRequestsByBotID(ctx context.Context, chatbotID string) ([]*models.HandoffRequest, error) {
	m.Calls.GetHandoffRequestsByBotID = append(m.Calls.GetHandoffRequestsByBotID, GetHandoffRequestsByBotIDCall{ChatbotID: chatbotID})
	if m.GetHandoffRequestsByBotIDFunc != nil {
		return m.GetHandoffRequestsByBotIDFunc(ctx, chatbotID)
	}
	return nil, nil
}

func (m *MockHandoffRepo) GetHandoffRequestByID(ctx context.Context, id string) (*models.HandoffRequest, error) {
	m.Calls.GetHandoffRequestByID = append(m.Calls.GetHandoffRequestByID, GetHandoffRequestByIDCall{ID: id})
	if m.GetHandoffRequestByIDFunc != nil {
		return m.GetHandoffRequestByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockHandoffRepo) UpdateHandoffRequestStatus(ctx context.Context, id, status string, assignedTo *string) error {
	m.Calls.UpdateHandoffRequestStatus = append(m.Calls.UpdateHandoffRequestStatus, UpdateHandoffRequestStatusCall{ID: id, Status: status, AssignedTo: assignedTo})
	if m.UpdateHandoffRequestStatusFunc != nil {
		return m.UpdateHandoffRequestStatusFunc(ctx, id, status, assignedTo)
	}
	return nil
}

func (m *MockHandoffRepo) CountPendingHandoffRequests(ctx context.Context, chatbotID string) (int, error) {
	m.Calls.CountPendingHandoffRequests = append(m.Calls.CountPendingHandoffRequests, CountPendingHandoffRequestsCall{ChatbotID: chatbotID})
	if m.CountPendingHandoffRequestsFunc != nil {
		return m.CountPendingHandoffRequestsFunc(ctx, chatbotID)
	}
	return 0, nil
}

func (m *MockHandoffRepo) ListHandoffMessages(ctx context.Context, conversationID string, limit int) ([]models.Message, error) {
	m.Calls.ListHandoffMessages = append(m.Calls.ListHandoffMessages, ListHandoffMessagesCall{ConversationID: conversationID, Limit: limit})
	if m.ListHandoffMessagesFunc != nil {
		return m.ListHandoffMessagesFunc(ctx, conversationID, limit)
	}
	return nil, nil
}

func (m *MockHandoffRepo) Reset() {
	m.Calls.HasActiveHandoffRequest = nil
	m.Calls.CreateHandoffRequest = nil
	m.Calls.GetHandoffRequestsByBotID = nil
	m.Calls.GetHandoffRequestByID = nil
	m.Calls.UpdateHandoffRequestStatus = nil
	m.Calls.CountPendingHandoffRequests = nil
	m.Calls.ListHandoffMessages = nil
}
