package services

import (
	"context"
	"net"
	"net/http"

	"github.com/onurceri/botla-app/internal/repository"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/logger"
)

type AdminService struct {
	adminRepo repository.AdminRepository
	Log       *logger.Logger
}

func NewAdminService(adminRepo repository.AdminRepository, log *logger.Logger) *AdminService {
	return &AdminService{
		adminRepo: adminRepo,
		Log:       log,
	}
}

// LogAction logs an admin action to the audit log.
func (s *AdminService) LogAction(ctx context.Context, adminID, action, targetType string, targetID *string, details map[string]any, r *http.Request) error {
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	entry := repository.AuditLogEntry{
		AdminUserID: adminID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		Details:     details,
		IPAddress:   ip,
		UserAgent:   r.Header.Get("User-Agent"),
	}

	if err := s.adminRepo.InsertAuditLog(ctx, entry); err != nil {
		s.Log.Error("failed to log admin action", map[string]any{
			"admin_id": adminID,
			"action":   action,
			"error":    err.Error(),
		})
		return pkgerrors.Wrapf(err, "insert audit log")
	}

	return nil
}
