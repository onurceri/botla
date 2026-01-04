package repository

import (
	"context"
)

// MockAdminRepo is a mock implementation of AdminRepository for testing.
// Each method can be customized by setting the corresponding function field.
// If a function field is nil, the method returns sensible defaults (nil/zero values).
type MockAdminRepo struct {
	// InsertAuditLogFunc is called when InsertAuditLog is invoked.
	InsertAuditLogFunc func(ctx context.Context, entry AuditLogEntry) error

	// ListAuditLogsFunc is called when ListAuditLogs is invoked.
	ListAuditLogsFunc func(ctx context.Context, filter AuditFilter, limit, offset int) ([]AuditLogEntry, int, error)

	// AdminListSourcesFunc is called when AdminListSources is invoked.
	AdminListSourcesFunc func(ctx context.Context, filter AdminSourceFilter, limit, offset int) ([]AdminSource, int, error)

	// AdminGetSourceByIDFunc is called when AdminGetSourceByID is invoked.
	AdminGetSourceByIDFunc func(ctx context.Context, id string) (*AdminSource, error)

	// AdminGetSourceStatsFunc is called when AdminGetSourceStats is invoked.
	AdminGetSourceStatsFunc func(ctx context.Context) (*SourceStats, error)

	// AdminReprocessSourceFunc is called when AdminReprocessSource is invoked.
	AdminReprocessSourceFunc func(ctx context.Context, id string) error

	// ListErrorLogsFunc is called when ListErrorLogs is invoked.
	ListErrorLogsFunc func(ctx context.Context, severity string, limit, offset int) ([]ErrorLogEntry, int, error)

	// GetErrorLogByIDFunc is called when GetErrorLogByID is invoked.
	GetErrorLogByIDFunc func(ctx context.Context, id string) (*ErrorLogEntry, error)

	// GetErrorStatsFunc is called when GetErrorStats is invoked.
	GetErrorStatsFunc func(ctx context.Context) (*ErrorStats, error)

	// Invocation tracking for test assertions
	Calls struct {
		InsertAuditLog       []MockInsertAuditLogCall
		ListAuditLogs        []MockListAuditLogsCall
		AdminListSources     []MockAdminListSourcesCall
		AdminGetSourceByID   []MockAdminGetSourceByIDCall
		AdminGetSourceStats  []MockAdminGetSourceStatsCall
		AdminReprocessSource []MockAdminReprocessSourceCall
		ListErrorLogs        []MockListErrorLogsCall
		GetErrorLogByID      []MockGetErrorLogByIDCall
		GetErrorStats        []MockGetErrorStatsCall
	}
}

// Call recording types for test verification.
type MockInsertAuditLogCall struct {
	Entry AuditLogEntry
}

type MockListAuditLogsCall struct {
	Filter AuditFilter
	Limit  int
	Offset int
}

type MockAdminListSourcesCall struct {
	Filter AdminSourceFilter
	Limit  int
	Offset int
}

type MockAdminGetSourceByIDCall struct {
	ID string
}

type MockAdminGetSourceStatsCall struct{}

type MockAdminReprocessSourceCall struct {
	ID string
}

type MockListErrorLogsCall struct {
	Severity string
	Limit    int
	Offset   int
}

type MockGetErrorLogByIDCall struct {
	ID string
}

type MockGetErrorStatsCall struct{}

// Compile-time check that MockAdminRepo implements AdminRepository.
var _ AdminRepository = (*MockAdminRepo)(nil)

// NewMockAdminRepo creates a new MockAdminRepo with default no-op behavior.
func NewMockAdminRepo() *MockAdminRepo {
	return &MockAdminRepo{}
}

// InsertAuditLog persists a new audit log entry.
func (m *MockAdminRepo) InsertAuditLog(ctx context.Context, entry AuditLogEntry) error {
	m.Calls.InsertAuditLog = append(m.Calls.InsertAuditLog, MockInsertAuditLogCall{Entry: entry})
	if m.InsertAuditLogFunc != nil {
		return m.InsertAuditLogFunc(ctx, entry)
	}
	return nil
}

// ListAuditLogs returns a paginated list of audit logs with optional filtering.
func (m *MockAdminRepo) ListAuditLogs(ctx context.Context, filter AuditFilter, limit, offset int) ([]AuditLogEntry, int, error) {
	m.Calls.ListAuditLogs = append(m.Calls.ListAuditLogs, MockListAuditLogsCall{Filter: filter, Limit: limit, Offset: offset})
	if m.ListAuditLogsFunc != nil {
		return m.ListAuditLogsFunc(ctx, filter, limit, offset)
	}
	return nil, 0, nil
}

// AdminListSources returns a paginated list of all data sources with metadata.
func (m *MockAdminRepo) AdminListSources(ctx context.Context, filter AdminSourceFilter, limit, offset int) ([]AdminSource, int, error) {
	m.Calls.AdminListSources = append(m.Calls.AdminListSources, MockAdminListSourcesCall{Filter: filter, Limit: limit, Offset: offset})
	if m.AdminListSourcesFunc != nil {
		return m.AdminListSourcesFunc(ctx, filter, limit, offset)
	}
	return nil, 0, nil
}

// AdminGetSourceByID retrieves a single source by ID with all admin-visible details.
func (m *MockAdminRepo) AdminGetSourceByID(ctx context.Context, id string) (*AdminSource, error) {
	m.Calls.AdminGetSourceByID = append(m.Calls.AdminGetSourceByID, MockAdminGetSourceByIDCall{ID: id})
	if m.AdminGetSourceByIDFunc != nil {
		return m.AdminGetSourceByIDFunc(ctx, id)
	}
	return nil, nil
}

// AdminGetSourceStats returns aggregated statistics for data sources.
func (m *MockAdminRepo) AdminGetSourceStats(ctx context.Context) (*SourceStats, error) {
	m.Calls.AdminGetSourceStats = append(m.Calls.AdminGetSourceStats, MockAdminGetSourceStatsCall{})
	if m.AdminGetSourceStatsFunc != nil {
		return m.AdminGetSourceStatsFunc(ctx)
	}
	return &SourceStats{StatusCounts: make(map[string]int)}, nil
}

// AdminReprocessSource resets a source to pending status for reprocessing.
func (m *MockAdminRepo) AdminReprocessSource(ctx context.Context, id string) error {
	m.Calls.AdminReprocessSource = append(m.Calls.AdminReprocessSource, MockAdminReprocessSourceCall{ID: id})
	if m.AdminReprocessSourceFunc != nil {
		return m.AdminReprocessSourceFunc(ctx, id)
	}
	return nil
}

// ListErrorLogs returns a paginated list of error logs with optional severity filtering.
func (m *MockAdminRepo) ListErrorLogs(ctx context.Context, severity string, limit, offset int) ([]ErrorLogEntry, int, error) {
	m.Calls.ListErrorLogs = append(m.Calls.ListErrorLogs, MockListErrorLogsCall{Severity: severity, Limit: limit, Offset: offset})
	if m.ListErrorLogsFunc != nil {
		return m.ListErrorLogsFunc(ctx, severity, limit, offset)
	}
	return nil, 0, nil
}

// GetErrorLogByID retrieves a single error log entry by ID.
func (m *MockAdminRepo) GetErrorLogByID(ctx context.Context, id string) (*ErrorLogEntry, error) {
	m.Calls.GetErrorLogByID = append(m.Calls.GetErrorLogByID, MockGetErrorLogByIDCall{ID: id})
	if m.GetErrorLogByIDFunc != nil {
		return m.GetErrorLogByIDFunc(ctx, id)
	}
	return nil, nil
}

// GetErrorStats returns aggregated error statistics for the last 24 hours.
func (m *MockAdminRepo) GetErrorStats(ctx context.Context) (*ErrorStats, error) {
	m.Calls.GetErrorStats = append(m.Calls.GetErrorStats, MockGetErrorStatsCall{})
	if m.GetErrorStatsFunc != nil {
		return m.GetErrorStatsFunc(ctx)
	}
	return &ErrorStats{SeverityCounts: make(map[string]int)}, nil
}

// Reset clears all recorded calls. Useful for resetting state between tests.
func (m *MockAdminRepo) Reset() {
	m.Calls.InsertAuditLog = nil
	m.Calls.ListAuditLogs = nil
	m.Calls.AdminListSources = nil
	m.Calls.AdminGetSourceByID = nil
	m.Calls.AdminGetSourceStats = nil
	m.Calls.AdminReprocessSource = nil
	m.Calls.ListErrorLogs = nil
	m.Calls.GetErrorLogByID = nil
	m.Calls.GetErrorStats = nil
}
