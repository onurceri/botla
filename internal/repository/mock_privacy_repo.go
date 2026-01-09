package repository

import (
	"context"
	"time"
)

// MockPrivacyRepo is a mock implementation of PrivacyRepository for testing.
type MockPrivacyRepo struct {
	CreateDataExportFunc               func(ctx context.Context, exp DataExport) (*DataExport, error)
	UpdateDataExportFunc               func(ctx context.Context, exp DataExport) error
	GetDataExportFunc                  func(ctx context.Context, id string) (*DataExport, error)
	CreatePrivacyRequestFunc           func(ctx context.Context, req PrivacyRequest) (*PrivacyRequest, error)
	GetPrivacyRequestFunc              func(ctx context.Context, requestID string) (*PrivacyRequest, error)
	ListPrivacyRequestsFunc            func(ctx context.Context, status string, limit, offset int) ([]PrivacyRequest, int, error)
	ListPrivacyRequestsByUserIDFunc    func(ctx context.Context, userID string, limit, offset int) ([]PrivacyRequest, int, error)
	HasActivePrivacyRequestFunc        func(ctx context.Context, userID, requestType string) (bool, error)
	UpdatePrivacyRequestStatusFunc     func(ctx context.Context, requestID, status, adminID string, denialReason *string) error
	CompletePrivacyExportRequestFunc   func(ctx context.Context, requestID, adminID, exportURL string, expiresAt time.Time) error
	GetUserDataForExportFunc           func(ctx context.Context, userID string) (*UserDataExport, error)
	GetUserFilesForDeletionFunc        func(ctx context.Context, userID string) ([]string, error)
	AnonymizeUserDataFunc              func(ctx context.Context, userID string) error
	GetUserConsentsFunc                func(ctx context.Context, userID string) ([]UserConsent, error)
	UpsertConsentFunc                  func(ctx context.Context, userID string, consentType string, granted bool, ipAddress, userAgent string) error

	Calls struct {
		CreateDataExport             []CreateDataExportCall
		UpdateDataExport             []UpdateDataExportCall
		GetDataExport                []GetDataExportCall
		CreatePrivacyRequest         []CreatePrivacyRequestCall
		GetPrivacyRequest            []GetPrivacyRequestCall
		ListPrivacyRequests          []ListPrivacyRequestsCall
		ListPrivacyRequestsByUserID  []ListPrivacyRequestsByUserIDCall
		HasActivePrivacyRequest      []HasActivePrivacyRequestCall
		UpdatePrivacyRequestStatus   []UpdatePrivacyRequestStatusCall
		CompletePrivacyExportRequest []CompletePrivacyExportRequestCall
		GetUserDataForExport         []GetUserDataForExportCall
		GetUserFilesForDeletion      []GetUserFilesForDeletionCall
		AnonymizeUserData            []AnonymizeUserDataCall
		GetUserConsents              []GetUserConsentsCall
		UpsertConsent                []UpsertConsentCall
	}
}

type CreateDataExportCall struct {
	Exp DataExport
}

type UpdateDataExportCall struct {
	Exp DataExport
}

type GetDataExportCall struct {
	ID string
}

type CreatePrivacyRequestCall struct {
	Req PrivacyRequest
}

type GetPrivacyRequestCall struct {
	RequestID string
}

type ListPrivacyRequestsCall struct {
	Status string
	Limit  int
	Offset int
}

type UpdatePrivacyRequestStatusCall struct {
	RequestID    string
	Status       string
	AdminID      string
	DenialReason *string
}

type CompletePrivacyExportRequestCall struct {
	RequestID string
	AdminID   string
	ExportURL string
	ExpiresAt time.Time
}

type GetUserDataForExportCall struct {
	UserID string
}

type GetUserFilesForDeletionCall struct {
	UserID string
}

type AnonymizeUserDataCall struct {
	UserID string
}

type GetUserConsentsCall struct {
	UserID string
}

type UpsertConsentCall struct {
	UserID      string
	ConsentType string
	Granted     bool
	IPAddress   string
	UserAgent   string
}

type ListPrivacyRequestsByUserIDCall struct {
	UserID string
	Limit  int
	Offset int
}

type HasActivePrivacyRequestCall struct {
	UserID      string
	RequestType string
}

var _ PrivacyRepository = (*MockPrivacyRepo)(nil)

func NewMockPrivacyRepo() *MockPrivacyRepo {
	return &MockPrivacyRepo{}
}

func (m *MockPrivacyRepo) CreateDataExport(ctx context.Context, exp DataExport) (*DataExport, error) {
	m.Calls.CreateDataExport = append(m.Calls.CreateDataExport, CreateDataExportCall{Exp: exp})
	if m.CreateDataExportFunc != nil {
		return m.CreateDataExportFunc(ctx, exp)
	}
	return nil, nil
}

func (m *MockPrivacyRepo) UpdateDataExport(ctx context.Context, exp DataExport) error {
	m.Calls.UpdateDataExport = append(m.Calls.UpdateDataExport, UpdateDataExportCall{Exp: exp})
	if m.UpdateDataExportFunc != nil {
		return m.UpdateDataExportFunc(ctx, exp)
	}
	return nil
}

func (m *MockPrivacyRepo) GetDataExport(ctx context.Context, id string) (*DataExport, error) {
	m.Calls.GetDataExport = append(m.Calls.GetDataExport, GetDataExportCall{ID: id})
	if m.GetDataExportFunc != nil {
		return m.GetDataExportFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockPrivacyRepo) CreatePrivacyRequest(ctx context.Context, req PrivacyRequest) (*PrivacyRequest, error) {
	m.Calls.CreatePrivacyRequest = append(m.Calls.CreatePrivacyRequest, CreatePrivacyRequestCall{Req: req})
	if m.CreatePrivacyRequestFunc != nil {
		return m.CreatePrivacyRequestFunc(ctx, req)
	}
	return nil, nil
}

func (m *MockPrivacyRepo) GetPrivacyRequest(ctx context.Context, requestID string) (*PrivacyRequest, error) {
	m.Calls.GetPrivacyRequest = append(m.Calls.GetPrivacyRequest, GetPrivacyRequestCall{RequestID: requestID})
	if m.GetPrivacyRequestFunc != nil {
		return m.GetPrivacyRequestFunc(ctx, requestID)
	}
	return nil, nil
}

func (m *MockPrivacyRepo) ListPrivacyRequests(ctx context.Context, status string, limit, offset int) ([]PrivacyRequest, int, error) {
	m.Calls.ListPrivacyRequests = append(m.Calls.ListPrivacyRequests, ListPrivacyRequestsCall{Status: status, Limit: limit, Offset: offset})
	if m.ListPrivacyRequestsFunc != nil {
		return m.ListPrivacyRequestsFunc(ctx, status, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockPrivacyRepo) UpdatePrivacyRequestStatus(ctx context.Context, requestID, status, adminID string, denialReason *string) error {
	m.Calls.UpdatePrivacyRequestStatus = append(m.Calls.UpdatePrivacyRequestStatus, UpdatePrivacyRequestStatusCall{RequestID: requestID, Status: status, AdminID: adminID, DenialReason: denialReason})
	if m.UpdatePrivacyRequestStatusFunc != nil {
		return m.UpdatePrivacyRequestStatusFunc(ctx, requestID, status, adminID, denialReason)
	}
	return nil
}

func (m *MockPrivacyRepo) CompletePrivacyExportRequest(ctx context.Context, requestID, adminID, exportURL string, expiresAt time.Time) error {
	m.Calls.CompletePrivacyExportRequest = append(m.Calls.CompletePrivacyExportRequest, CompletePrivacyExportRequestCall{RequestID: requestID, AdminID: adminID, ExportURL: exportURL, ExpiresAt: expiresAt})
	if m.CompletePrivacyExportRequestFunc != nil {
		return m.CompletePrivacyExportRequestFunc(ctx, requestID, adminID, exportURL, expiresAt)
	}
	return nil
}

func (m *MockPrivacyRepo) GetUserDataForExport(ctx context.Context, userID string) (*UserDataExport, error) {
	m.Calls.GetUserDataForExport = append(m.Calls.GetUserDataForExport, GetUserDataForExportCall{UserID: userID})
	if m.GetUserDataForExportFunc != nil {
		return m.GetUserDataForExportFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPrivacyRepo) GetUserFilesForDeletion(ctx context.Context, userID string) ([]string, error) {
	m.Calls.GetUserFilesForDeletion = append(m.Calls.GetUserFilesForDeletion, GetUserFilesForDeletionCall{UserID: userID})
	if m.GetUserFilesForDeletionFunc != nil {
		return m.GetUserFilesForDeletionFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPrivacyRepo) AnonymizeUserData(ctx context.Context, userID string) error {
	m.Calls.AnonymizeUserData = append(m.Calls.AnonymizeUserData, AnonymizeUserDataCall{UserID: userID})
	if m.AnonymizeUserDataFunc != nil {
		return m.AnonymizeUserDataFunc(ctx, userID)
	}
	return nil
}

func (m *MockPrivacyRepo) GetUserConsents(ctx context.Context, userID string) ([]UserConsent, error) {
	m.Calls.GetUserConsents = append(m.Calls.GetUserConsents, GetUserConsentsCall{UserID: userID})
	if m.GetUserConsentsFunc != nil {
		return m.GetUserConsentsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockPrivacyRepo) UpsertConsent(ctx context.Context, userID string, consentType string, granted bool, ipAddress, userAgent string) error {
	m.Calls.UpsertConsent = append(m.Calls.UpsertConsent, UpsertConsentCall{UserID: userID, ConsentType: consentType, Granted: granted, IPAddress: ipAddress, UserAgent: userAgent})
	if m.UpsertConsentFunc != nil {
		return m.UpsertConsentFunc(ctx, userID, consentType, granted, ipAddress, userAgent)
	}
	return nil
}

func (m *MockPrivacyRepo) ListPrivacyRequestsByUserID(ctx context.Context, userID string, limit, offset int) ([]PrivacyRequest, int, error) {
	m.Calls.ListPrivacyRequestsByUserID = append(m.Calls.ListPrivacyRequestsByUserID, ListPrivacyRequestsByUserIDCall{UserID: userID, Limit: limit, Offset: offset})
	if m.ListPrivacyRequestsByUserIDFunc != nil {
		return m.ListPrivacyRequestsByUserIDFunc(ctx, userID, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockPrivacyRepo) HasActivePrivacyRequest(ctx context.Context, userID, requestType string) (bool, error) {
	m.Calls.HasActivePrivacyRequest = append(m.Calls.HasActivePrivacyRequest, HasActivePrivacyRequestCall{UserID: userID, RequestType: requestType})
	if m.HasActivePrivacyRequestFunc != nil {
		return m.HasActivePrivacyRequestFunc(ctx, userID, requestType)
	}
	return false, nil
}

func (m *MockPrivacyRepo) Reset() {
	m.Calls.CreateDataExport = nil
	m.Calls.UpdateDataExport = nil
	m.Calls.GetDataExport = nil
	m.Calls.CreatePrivacyRequest = nil
	m.Calls.GetPrivacyRequest = nil
	m.Calls.ListPrivacyRequests = nil
	m.Calls.ListPrivacyRequestsByUserID = nil
	m.Calls.HasActivePrivacyRequest = nil
	m.Calls.UpdatePrivacyRequestStatus = nil
	m.Calls.CompletePrivacyExportRequest = nil
	m.Calls.GetUserDataForExport = nil
	m.Calls.GetUserFilesForDeletion = nil
	m.Calls.AnonymizeUserData = nil
	m.Calls.GetUserConsents = nil
	m.Calls.UpsertConsent = nil
}
