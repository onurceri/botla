package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/services"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

func TestRetentionJob_Conversations(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// 1. Setup Data
	// Create user & chatbot via API to ensure valid foreign keys
	token := authToken(t, te.Server.URL, "retention@example.com")

	// create chatbot inline helper
	createChatbot := func() string {
		create := map[string]any{"name": "Retention Bot", "language": "en-US"}
		cb, _ := json.Marshal(create)
		req, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cb))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		res, _ := http.DefaultClient.Do(req)
		if res.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", res.StatusCode)
		}
		var created struct {
			ID string `json:"id"`
		}
		json.NewDecoder(res.Body).Decode(&created)
		res.Body.Close()
		return created.ID
	}
	chatbotID := createChatbot()

	// 2. Insert old conversation (731 days old)
	oldDate := time.Now().AddDate(0, 0, -731)
	var oldConvID string
	err = te.DB.QueryRow(`
		INSERT INTO conversations (chatbot_id, session_id, created_at, updated_at)
		VALUES ($1, 'old-session', $2, $2)
		RETURNING id
	`, chatbotID, oldDate).Scan(&oldConvID)
	if err != nil {
		t.Fatalf("failed to insert old conversation: %v", err)
	}

	// 3. Insert new conversation (1 day old)
	newDate := time.Now().AddDate(0, 0, -1)
	var newConvID string
	err = te.DB.QueryRow(`
		INSERT INTO conversations (chatbot_id, session_id, created_at, updated_at)
		VALUES ($1, 'new-session', $2, $2)
		RETURNING id
	`, chatbotID, newDate).Scan(&newConvID)
	if err != nil {
		t.Fatalf("failed to insert new conversation: %v", err)
	}

	// 4. Run Retention Job
	log := logger.New("DEBUG")
	mockStorage := &storage.MemoryStorage{}
	job := services.NewRetentionJob(te.DB, log, mockStorage)
	// Default retention is 730 days
	err = job.Run(context.Background())
	if err != nil {
		t.Fatalf("job run failed: %v", err)
	}

	// 5. Verify
	var exists bool
	// Old should be gone
	err = te.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM conversations WHERE id=$1)", oldConvID).Scan(&exists)
	if err != nil {
		t.Fatalf("check old failed: %v", err)
	}
	if exists {
		t.Error("old conversation should be deleted")
	}

	// New should remain
	err = te.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM conversations WHERE id=$1)", newConvID).Scan(&exists)
	if err != nil {
		t.Fatalf("check new failed: %v", err)
	}
	if !exists {
		t.Error("new conversation should remain")
	}
}

func TestRetentionJob_ExpiredExports(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// We need a user for exports
	_ = authToken(t, te.Server.URL, "export@example.com")
	// We need to decode token to get user ID or just query DB for user
	var userID string
	err = te.DB.QueryRow("SELECT id FROM users WHERE email='export@example.com'").Scan(&userID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	// 1. Insert old export (expired 1 day ago)
	oldDate := time.Now().AddDate(0, 0, -1)
	_, err = te.DB.Exec(`
		INSERT INTO data_exports (user_id, status, format, expires_at)
		VALUES ($1, 'completed', 'json', $2)
	`, userID, oldDate)
	if err != nil {
		t.Fatalf("failed to insert old export: %v", err)
	}

	// 2. Insert new export (expires in 6 days)
	newDate := time.Now().AddDate(0, 0, 6)
	_, err = te.DB.Exec(`
		INSERT INTO data_exports (user_id, status, format, expires_at)
		VALUES ($1, 'completed', 'json', $2)
	`, userID, newDate)
	if err != nil {
		t.Fatalf("failed to insert new export: %v", err)
	}

	// 3. Run Retention Job
	log := logger.New("DEBUG")
	mockStorage := &storage.MemoryStorage{}
	job := services.NewRetentionJob(te.DB, log, mockStorage)
	err = job.Run(context.Background())
	if err != nil {
		t.Fatalf("job run failed: %v", err)
	}

	// 4. Verify
	var count int
	err = te.DB.QueryRow("SELECT COUNT(*) FROM data_exports WHERE expires_at < $1", time.Now()).Scan(&count)
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count > 0 {
		t.Errorf("expected 0 expired exports, got %d", count)
	}

	err = te.DB.QueryRow("SELECT COUNT(*) FROM data_exports WHERE expires_at > $1", time.Now()).Scan(&count)
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 valid export, got %d", count)
	}
}

func TestRetentionJob_AuditLogs(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	// We need a user for audit logs
	_ = authToken(t, te.Server.URL, "admin@example.com")
	var adminID string
	err = te.DB.QueryRow("SELECT id FROM users WHERE email='admin@example.com'").Scan(&adminID)
	if err != nil {
		t.Fatalf("failed to get admin user: %v", err)
	}

	// 1. Setup old audit logs (731 days old)
	oldDate := time.Now().AddDate(0, 0, -731)
	_, err = te.DB.Exec(`
		INSERT INTO admin_audit_logs (admin_user_id, action, target_type, created_at)
		VALUES ($1, 'test_action', 'test_target', $2)
	`, adminID, oldDate)
	if err != nil {
		t.Fatalf("failed to insert old audit log: %v", err)
	}

	// 2. Setup new audit logs (1 day old)
	newDate := time.Now().AddDate(0, 0, -1)
	_, err = te.DB.Exec(`
		INSERT INTO admin_audit_logs (admin_user_id, action, target_type, created_at)
		VALUES ($1, 'test_action', 'test_target', $2)
	`, adminID, newDate)
	if err != nil {
		t.Fatalf("failed to insert new audit log: %v", err)
	}

	// 3. Run Retention Job
	log := logger.New("DEBUG")
	mockStorage := &storage.MemoryStorage{}
	job := services.NewRetentionJob(te.DB, log, mockStorage)
	err = job.Run(context.Background())
	if err != nil {
		t.Fatalf("job run failed: %v", err)
	}

	// 4. Verify
	var count int
	// Old should be gone (cutoff is 730 days by default)
	err = te.DB.QueryRow("SELECT COUNT(*) FROM admin_audit_logs WHERE created_at < $1", time.Now().AddDate(0, 0, -730)).Scan(&count)
	if err != nil {
		t.Fatalf("count old failed: %v", err)
	}
	if count > 0 {
		t.Errorf("expected 0 old audit logs, got %d", count)
	}

	// New should remain
	err = te.DB.QueryRow("SELECT COUNT(*) FROM admin_audit_logs WHERE created_at > $1", time.Now().AddDate(0, 0, -730)).Scan(&count)
	if err != nil {
		t.Fatalf("count new failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 new audit log, got %d", count)
	}
}
