package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertAuditLog_Success(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{
		IsPlatformAdmin: true,
	})

	targetID := uuid.NewString()
	details := map[string]any{
		"previous_value": "old_value",
		"new_value":      "new_value",
		"changed_fields": []string{"name", "email"},
	}

	entry := db.AuditLogEntry{
		AdminUserID: admin.ID,
		Action:      "user_update",
		TargetType:  "user",
		TargetID:    &targetID,
		Details:     details,
		IPAddress:   "192.168.1.100",
		UserAgent:   "Mozilla/5.0 Test Browser",
	}

	err := db.InsertAuditLog(ctx, dbConn, entry)
	require.NoError(t, err)

	var dbID, dbDetails string
	var dbCreatedAt time.Time
	err = dbConn.QueryRowContext(ctx,
		"SELECT id, details, created_at FROM admin_audit_logs WHERE admin_user_id = $1 AND action = $2 LIMIT 1",
		admin.ID, "user_update",
	).Scan(&dbID, &dbDetails, &dbCreatedAt)

	require.NoError(t, err)
	assert.NotEmpty(t, dbID)
	assert.Contains(t, dbDetails, "new_value")
	assert.False(t, dbCreatedAt.IsZero())
	assert.WithinDuration(t, time.Now(), dbCreatedAt, time.Second*5)
}

func TestInsertAuditLog_WithNilTargetID(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{
		IsPlatformAdmin: true,
	})

	entry := db.AuditLogEntry{
		AdminUserID: admin.ID,
		Action:      "system_check",
		TargetType:  "system",
		TargetID:    nil,
		Details:     map[string]any{"check_type": "health"},
		IPAddress:   "192.168.1.100",
		UserAgent:   "Mozilla/5.0 Test Browser",
	}

	err := db.InsertAuditLog(ctx, dbConn, entry)
	require.NoError(t, err)

	var targetID *string
	err = dbConn.QueryRowContext(ctx,
		"SELECT target_id FROM admin_audit_logs WHERE admin_user_id = $1 AND action = $2 LIMIT 1",
		admin.ID, "system_check",
	).Scan(&targetID)

	require.NoError(t, err)
	assert.Nil(t, targetID)
}

func TestInsertAuditLog_WithComplexDetails(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{
		IsPlatformAdmin: true,
	})

	details := map[string]any{
		"user_changes": map[string]any{
			"old": map[string]string{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"new": map[string]string{
				"name":  "Jane Smith",
				"email": "jane@example.com",
			},
		},
		"metadata": map[string]string{
			"source":    "admin_panel",
			"reason":    "user_request",
			"ticket_id": "12345",
		},
		"changed_fields_count": 2,
	}

	entry := db.AuditLogEntry{
		AdminUserID: admin.ID,
		Action:      "user_profile_update",
		TargetType:  "user",
		Details:     details,
		IPAddress:   "192.168.1.100",
		UserAgent:   "Mozilla/5.0 Test Browser",
	}

	err := db.InsertAuditLog(ctx, dbConn, entry)
	require.NoError(t, err)

	var dbID string
	err = dbConn.QueryRowContext(ctx,
		"SELECT id FROM admin_audit_logs WHERE admin_user_id = $1 AND action = $2 LIMIT 1",
		admin.ID, "user_profile_update",
	).Scan(&dbID)

	require.NoError(t, err)
	assert.NotEmpty(t, dbID)
}

func TestListAuditLogs_NoFilters(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin1 := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	admin2 := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})

	now := time.Now()

	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "create", "chatbot", uuid.NewString(), now.Add(-2*time.Hour))
	_ = createAuditLog(t, ctx, dbConn, admin2.ID, "update", "user", uuid.NewString(), now.Add(-1*time.Hour))
	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "delete", "source", uuid.NewString(), now.Add(-30*time.Minute))

	filter := db.AuditFilter{}
	logs, total, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, logs, 3)

	assert.Equal(t, "delete", logs[0].Action)
	assert.Equal(t, "update", logs[1].Action)
	assert.Equal(t, "create", logs[2].Action)
}

func TestListAuditLogs_FilterByAdminUserID(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin1 := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	admin2 := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	admin3 := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})

	now := time.Now()

	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "create", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "update", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "delete", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "create", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "update", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin2.ID, "create", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin2.ID, "update", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin2.ID, "delete", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin3.ID, "create", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin3.ID, "update", "chatbot", uuid.NewString(), now)

	filter := db.AuditFilter{AdminUserID: &admin1.ID}
	logs, total, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, logs, 5)

	for _, log := range logs {
		assert.Equal(t, admin1.ID, log.AdminUserID)
	}
}

func TestListAuditLogs_FilterByAction(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	now := time.Now()

	_ = createAuditLog(t, ctx, dbConn, admin.ID, "create", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "create", "user", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "update", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "delete", "source", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "create", "organization", uuid.NewString(), now)

	action := "create"
	filter := db.AuditFilter{Action: &action}
	logs, total, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, logs, 3)

	for _, log := range logs {
		assert.Equal(t, "create", log.Action)
	}
}

func TestListAuditLogs_FilterByDateRange(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	now := time.Now()

	oldDate := now.Add(-10 * 24 * time.Hour)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "old", "chatbot", uuid.NewString(), oldDate)

	startDate := now.Add(-24 * time.Hour)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "within", "chatbot", uuid.NewString(), startDate.Add(2*time.Hour))
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "within", "chatbot", uuid.NewString(), now.Add(-12*time.Hour))
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "within", "chatbot", uuid.NewString(), now.Add(-6*time.Hour))

	endDate := now.Add(-2 * time.Hour)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "future", "chatbot", uuid.NewString(), now.Add(1*time.Hour))

	filter := db.AuditFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	logs, total, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, logs, 3)

	for _, log := range logs {
		assert.True(t, log.CreatedAt.After(startDate) || log.CreatedAt.Equal(startDate))
		assert.True(t, log.CreatedAt.Before(endDate) || log.CreatedAt.Equal(endDate))
	}
}

func TestListAuditLogs_FilterByTargetType(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	now := time.Now()

	_ = createAuditLog(t, ctx, dbConn, admin.ID, "action", "user", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "action", "user", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "action", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "action", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "action", "chatbot", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "action", "organization", uuid.NewString(), now)

	targetType := "chatbot"
	filter := db.AuditFilter{TargetType: &targetType}
	logs, total, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, logs, 3)

	for _, log := range logs {
		assert.Equal(t, "chatbot", log.TargetType)
	}
}

func TestListAuditLogs_FilterByTargetID(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	now := time.Now()

	targetID1 := uuid.NewString()
	targetID2 := uuid.NewString()

	_ = createAuditLog(t, ctx, dbConn, admin.ID, "create", "chatbot", targetID1, now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "update", "chatbot", targetID1, now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "delete", "chatbot", targetID1, now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "update", "chatbot", targetID2, now)

	filter := db.AuditFilter{TargetID: &targetID1}
	logs, total, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, logs, 3)

	for _, log := range logs {
		assert.Equal(t, targetID1, *log.TargetID)
	}
}

func TestListAuditLogs_CombinedFilters(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin1 := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	admin2 := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	now := time.Now()

	startDate := now.Add(-24 * time.Hour)
	endDate := now.Add(-2 * time.Hour)
	targetID := uuid.NewString()

	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "create", "chatbot", targetID, startDate.Add(1*time.Hour))
	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "update", "chatbot", targetID, startDate.Add(2*time.Hour))
	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "delete", "chatbot", uuid.NewString(), now.Add(-1*time.Hour))
	_ = createAuditLog(t, ctx, dbConn, admin2.ID, "create", "chatbot", targetID, startDate.Add(3*time.Hour))
	_ = createAuditLog(t, ctx, dbConn, admin1.ID, "create", "user", uuid.NewString(), startDate.Add(4*time.Hour))

	filter := db.AuditFilter{
		AdminUserID: &admin1.ID,
		Action:      ptr("update"),
		TargetType:  ptr("chatbot"),
		TargetID:    &targetID,
		StartDate:   &startDate,
		EndDate:     &endDate,
	}
	logs, total, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, logs, 1)

	assert.Equal(t, admin1.ID, logs[0].AdminUserID)
	assert.Equal(t, "update", logs[0].Action)
	assert.Equal(t, "chatbot", logs[0].TargetType)
	assert.Equal(t, targetID, *logs[0].TargetID)
}

func TestListAuditLogs_Pagination(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	now := time.Now()

	var expectedIDs []string
	for i := 0; i < 25; i++ {
		id := createAuditLog(t, ctx, dbConn, admin.ID, "action", "chatbot", uuid.NewString(), now.Add(-time.Duration(i)*time.Minute))
		expectedIDs = append(expectedIDs, id)
	}

	page1, total1, err := db.ListAuditLogs(ctx, dbConn, db.AuditFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 25, total1)
	assert.Len(t, page1, 10)
	assert.Equal(t, expectedIDs[0], page1[0].ID)

	page2, total2, err := db.ListAuditLogs(ctx, dbConn, db.AuditFilter{}, 10, 10)
	require.NoError(t, err)
	assert.Equal(t, 25, total2)
	assert.Len(t, page2, 10)
	assert.Equal(t, expectedIDs[10], page2[0].ID)

	page3, total3, err := db.ListAuditLogs(ctx, dbConn, db.AuditFilter{}, 10, 20)
	require.NoError(t, err)
	assert.Equal(t, 25, total3)
	assert.Len(t, page3, 5)
	assert.Equal(t, expectedIDs[20], page3[0].ID)
}

func TestListAuditLogs_EmptyResults(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	nonExistentAdminID := "00000000-0000-0000-0000-000000000000"
	filter := db.AuditFilter{AdminUserID: &nonExistentAdminID}
	logs, total, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Len(t, logs, 0)
}

func TestListAuditLogs_DetailsUnmarshaling(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})

	details := map[string]any{
		"nested": map[string]any{
			"deeply": map[string]any{
				"value": "test",
			},
		},
		"array":  []string{"item1", "item2", "item3"},
		"number": 123,
	}

	entry := db.AuditLogEntry{
		AdminUserID: admin.ID,
		Action:      "complex",
		TargetType:  "test",
		Details:     details,
		IPAddress:   "192.168.1.100",
		UserAgent:   "Test Browser",
	}

	err := db.InsertAuditLog(ctx, dbConn, entry)
	require.NoError(t, err)

	filter := db.AuditFilter{Action: ptr("complex")}
	logs, _, err := db.ListAuditLogs(ctx, dbConn, filter, 10, 0)

	require.NoError(t, err)
	assert.Len(t, logs, 1)

	nestedMap := logs[0].Details["nested"].(map[string]any)
	deeplyMap := nestedMap["deeply"].(map[string]any)
	assert.Equal(t, "test", deeplyMap["value"])

	array := logs[0].Details["array"].([]any)
	assert.Equal(t, 3, len(array))

	assert.Equal(t, float64(123), logs[0].Details["number"])
}

func TestListAuditLogs_SortingByCreatedAt(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	now := time.Now()

	oldLogID := createAuditLog(t, ctx, dbConn, admin.ID, "old", "chatbot", uuid.NewString(), now.Add(-3*time.Hour))
	middleLogID := createAuditLog(t, ctx, dbConn, admin.ID, "middle", "chatbot", uuid.NewString(), now.Add(-2*time.Hour))
	newLogID := createAuditLog(t, ctx, dbConn, admin.ID, "new", "chatbot", uuid.NewString(), now.Add(-1*time.Hour))

	logs, _, err := db.ListAuditLogs(ctx, dbConn, db.AuditFilter{}, 10, 0)

	require.NoError(t, err)
	assert.Len(t, logs, 3)

	assert.Equal(t, newLogID, logs[0].ID)
	assert.Equal(t, middleLogID, logs[1].ID)
	assert.Equal(t, oldLogID, logs[2].ID)
}

func TestListAuditLogs_AllNilFilter(t *testing.T) {
	dbConn := testdb.OpenParallelTestDB(t)
	ctx := context.Background()

	admin := testdb.CreateUser(t, dbConn, testdb.UserFixture{IsPlatformAdmin: true})
	now := time.Now()

	_ = createAuditLog(t, ctx, dbConn, admin.ID, "action1", "type1", uuid.NewString(), now)
	_ = createAuditLog(t, ctx, dbConn, admin.ID, "action2", "type2", uuid.NewString(), now)

	logs, total, err := db.ListAuditLogs(ctx, dbConn, db.AuditFilter{}, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, logs, 2)
}

func createAuditLog(t *testing.T, ctx context.Context, dbConn *sql.DB, adminID, action, targetType, targetID string, createdAt time.Time) string {
	t.Helper()

	details := map[string]any{"test": "data"}
	entry := db.AuditLogEntry{
		AdminUserID: adminID,
		Action:      action,
		TargetType:  targetType,
		Details:     details,
		IPAddress:   "127.0.0.1",
		UserAgent:   "Test Agent",
	}

	if targetID != "" {
		entry.TargetID = &targetID
	}

	err := db.InsertAuditLog(ctx, dbConn, entry)
	require.NoError(t, err)

	var logID string
	err = dbConn.QueryRowContext(ctx,
		"SELECT id FROM admin_audit_logs WHERE admin_user_id = $1 AND action = $2 ORDER BY created_at DESC LIMIT 1",
		adminID, action,
	).Scan(&logID)
	require.NoError(t, err)

	_, err = dbConn.ExecContext(ctx,
		"UPDATE admin_audit_logs SET created_at = $1 WHERE id = $2",
		createdAt, logID,
	)
	require.NoError(t, err)

	return logID
}
