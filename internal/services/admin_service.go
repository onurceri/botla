package services

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/logger"
)

type AdminService struct {
	DB  *sql.DB
	Log *logger.Logger
}

func NewAdminService(db *sql.DB, log *logger.Logger) *AdminService {
	return &AdminService{
		DB:  db,
		Log: log,
	}
}

// LogAction logs an admin action to the audit log.
func (s *AdminService) LogAction(ctx context.Context, adminID, action, targetType string, targetID *string, details map[string]any, r *http.Request) error {
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	entry := db.AuditLogEntry{
		AdminUserID: adminID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Details:     details,
		IPAddress:   ip,
		UserAgent:   r.Header.Get("User-Agent"),
	}

	if err := db.InsertAuditLog(ctx, s.DB, entry); err != nil {
		s.Log.Error("failed to log admin action", map[string]any{
			"admin_id": adminID,
			"action":   action,
			"error":    err.Error(),
		})
		return fmt.Errorf("insert audit log: %w", err)
	}

	return nil
}
