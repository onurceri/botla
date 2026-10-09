package repository

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
)

type MockHandoffRepo struct {
	HasActiveHandoffRequestFunc     func(ctx context.Context, conversationID string) (bool, error)
	CreateHandoffRequestFunc        func(ctx context.Context, req *models.HandoffRequest) (string, error)
	ListHandoffRequestsByBotIDFunc  func(ctx context.Context, chatbotID string, filter models.HandoffListFilter) ([]*models.HandoffRequest, int, error)
	GetHandoffRequestByIDFunc       func(ctx context.Context, id string) (*models.HandoffRequest, error)
	UpdateHandoffRequestStatusFunc  func(ctx context.Context, id, chatbotID, status string, assignedTo *string) error
	CountPendingHandoffRequestsFunc func(ctx context.Context, chatbotID string) (int, error)
	ListHandoffMessagesFunc         func(ctx context.Context, conversationID string, limit int) ([]models.Message, error)
	SetHandoffContactEmailFunc      func(ctx context.Context, id, chatbotID, email string) error

	Calls struct {
		HasActiveHandoffRequest     []HasActiveHandoffRequestCall
		CreateHandoffRequest        []CreateHandoffRequestCall
		ListHandoffRequestsByBotID  []ListHandoffRequestsByBotIDCall
		GetHandoffRequestByID       []GetHandoffRequestByIDCall
		UpdateHandoffRequestStatus  []UpdateHandoffRequestStatusCall
		CountPendingHandoffRequests []CountPendingHandoffRequestsCall
		ListHandoffMessages         []ListHandoffMessagesCall
		SetHandoffContactEmail      []SetHandoffContactEmailCall
	}
}

type HasActiveHandoffRequestCall struct {
	ConversationID string
}

type CreateHandoffRequestCall struct {
	Req *models.HandoffRequest
}

type ListHandoffRequestsByBotIDCall struct {
	ChatbotID string
	Filter    models.HandoffListFilter
}

type GetHandoffRequestByIDCall struct {
	ID string
}

type UpdateHandoffRequestStatusCall struct {
	ID         string
	ChatbotID  string
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

type SetHandoffContactEmailCall struct {
	ID        string
	ChatbotID string
	Email     string
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

func (m *MockHandoffRepo) ListHandoffRequestsByBotID(ctx context.Context, chatbotID string, filter models.HandoffListFilter) ([]*models.HandoffRequest, int, error) {
	m.Calls.ListHandoffRequestsByBotID = append(m.Calls.ListHandoffRequestsByBotID, ListHandoffRequestsByBotIDCall{ChatbotID: chatbotID, Filter: filter})
	if m.ListHandoffRequestsByBotIDFunc != nil {
		return m.ListHandoffRequestsByBotIDFunc(ctx, chatbotID, filter)
	}
	return nil, 0, nil
}

func (m *MockHandoffRepo) GetHandoffRequestByID(ctx context.Context, id string) (*models.HandoffRequest, error) {
	m.Calls.GetHandoffRequestByID = append(m.Calls.GetHandoffRequestByID, GetHandoffRequestByIDCall{ID: id})
	if m.GetHandoffRequestByIDFunc != nil {
		return m.GetHandoffRequestByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockHandoffRepo) UpdateHandoffRequestStatus(ctx context.Context, id, chatbotID, status string, assignedTo *string) error {
	m.Calls.UpdateHandoffRequestStatus = append(m.Calls.UpdateHandoffRequestStatus, UpdateHandoffRequestStatusCall{ID: id, ChatbotID: chatbotID, Status: status, AssignedTo: assignedTo})
	if m.UpdateHandoffRequestStatusFunc != nil {
		return m.UpdateHandoffRequestStatusFunc(ctx, id, chatbotID, status, assignedTo)
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

func (m *MockHandoffRepo) SetHandoffContactEmail(ctx context.Context, id, chatbotID, email string) error {
	m.Calls.SetHandoffContactEmail = append(m.Calls.SetHandoffContactEmail, SetHandoffContactEmailCall{ID: id, ChatbotID: chatbotID, Email: email})
	if m.SetHandoffContactEmailFunc != nil {
		return m.SetHandoffContactEmailFunc(ctx, id, chatbotID, email)
	}
	return nil
}

func (m *MockHandoffRepo) Reset() {
	m.Calls.HasActiveHandoffRequest = nil
	m.Calls.CreateHandoffRequest = nil
	m.Calls.ListHandoffRequestsByBotID = nil
	m.Calls.GetHandoffRequestByID = nil
	m.Calls.UpdateHandoffRequestStatus = nil
	m.Calls.CountPendingHandoffRequests = nil
	m.Calls.ListHandoffMessages = nil
	m.Calls.SetHandoffContactEmail = nil
}
