package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/onurceri/botla-app/internal/repository"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/storage"
)

// ErrActiveRequestExists is returned when user already has a pending/processing request of the same type.
var ErrActiveRequestExists = errors.New("active request already exists")

type PrivacyService struct {
	PrivacyRepo repository.PrivacyRepository
	Log         *logger.Logger
	Storage     storage.StorageService
}

func NewPrivacyService(privacyRepo repository.PrivacyRepository, log *logger.Logger, storage storage.StorageService) *PrivacyService {
	return &PrivacyService{
		PrivacyRepo: privacyRepo,
		Log:         log,
		Storage:     storage,
	}
}

// ExportUserData generates a JSON export of all user data
func (s *PrivacyService) ExportUserData(ctx context.Context, userID string, requestedBy string) (*repository.DataExport, error) {
	if s.Storage == nil {
		return nil, fmt.Errorf("storage not configured")
	}

	// 1. Create data_exports record (pending)
	export := repository.DataExport{
		UserID:      &userID,
		RequestedBy: &requestedBy,
		Format:      "json",
		Status:      "processing",
	}
	createdExport, err := s.PrivacyRepo.CreateDataExport(ctx, export)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create data export")
	}

	// Run export in background to avoid timeout
	go func(expID string) {
		// Create a new context for background job
		bgCtx := context.Background()

		_, _, _, err := s.processExport(bgCtx, expID, userID)
		if err != nil {
			s.Log.Error("export_failed", map[string]any{
				"export_id": expID,
				"user_id":   userID,
				"error":     err.Error(),
			})

			// Update status to failed
			errMsg := err.Error()
			failedExp := repository.DataExport{
				ID:           expID,
				Status:       "failed",
				ErrorMessage: &errMsg,
			}
			_ = s.PrivacyRepo.UpdateDataExport(bgCtx, failedExp)
		}
	}(createdExport.ID)

	return createdExport, nil
}

func (s *PrivacyService) processExport(ctx context.Context, exportID, userID string) (string, time.Time, int64, error) {
	if s.Storage == nil {
		return "", time.Time{}, 0, fmt.Errorf("storage not configured")
	}

	// 1. Collect all user data using the repository helper
	exportData, err := s.PrivacyRepo.GetUserDataForExport(ctx, userID)
	if err != nil {
		return "", time.Time{}, 0, pkgerrors.Wrapf(err, "get user data for export")
	}

	// 2. Generate JSON file
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", time.Time{}, 0, pkgerrors.Wrapf(err, "marshal json")
	}

	// 3. Upload to storage
	filename := fmt.Sprintf("exports/%s/%s-%d.json", userID, exportID, time.Now().Unix())
	fileURL, err := s.Storage.UploadFile(ctx, filename, bytes.NewReader(jsonData))
	if err != nil {
		return "", time.Time{}, 0, pkgerrors.Wrapf(err, "upload file")
	}

	// 4. Update data_exports record
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days expiration
	size := int64(len(jsonData))

	update := repository.DataExport{
		ID:            exportID,
		Status:        "completed",
		DownloadURL:   &fileURL,
		FileSizeBytes: &size,
		ExpiresAt:     &expiresAt,
		CompletedAt:   &now,
	}

	if err := s.PrivacyRepo.UpdateDataExport(ctx, update); err != nil {
		return "", time.Time{}, 0, pkgerrors.Wrapf(err, "update data export")
	}

	return fileURL, expiresAt, size, nil
}

// RequestDeletion initiates user data deletion process
func (s *PrivacyService) RequestDeletion(ctx context.Context, userID, email, reason string) (*repository.PrivacyRequest, error) {
	req := repository.PrivacyRequest{
		UserID:      &userID,
		UserEmail:   email,
		RequestType: "deletion",
		Status:      "pending",
		Reason:      reason,
	}
	res, err := s.PrivacyRepo.CreatePrivacyRequest(ctx, req)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create deletion request")
	}
	return res, nil
}

func (s *PrivacyService) RequestExport(ctx context.Context, userID, email, reason string) (*repository.PrivacyRequest, error) {
	hasActive, err := s.PrivacyRepo.HasActivePrivacyRequest(ctx, userID, "export")
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "check active export request")
	}
	if hasActive {
		return nil, ErrActiveRequestExists
	}

	req := repository.PrivacyRequest{
		UserID:      &userID,
		UserEmail:   email,
		RequestType: "export",
		Status:      "pending",
		Reason:      reason,
	}
	res, err := s.PrivacyRepo.CreatePrivacyRequest(ctx, req)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create export request")
	}
	return res, nil
}

func (s *PrivacyService) RequestCorrection(ctx context.Context, userID, email, reason string) (*repository.PrivacyRequest, error) {
	req := repository.PrivacyRequest{
		UserID:      &userID,
		UserEmail:   email,
		RequestType: "correction",
		Status:      "pending",
		Reason:      reason,
	}
	res, err := s.PrivacyRepo.CreatePrivacyRequest(ctx, req)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "create correction request")
	}
	return res, nil
}

func (s *PrivacyService) ProcessExportRequest(ctx context.Context, requestID, adminID string) error {
	req, err := s.PrivacyRepo.GetPrivacyRequest(ctx, requestID)
	if err != nil {
		return pkgerrors.Wrapf(err, "get privacy request")
	}
	if req == nil {
		return fmt.Errorf("request not found")
	}
	if req.RequestType != "export" {
		return fmt.Errorf("request is not an export request")
	}

	if req.UserID == nil {
		if errStatus := s.PrivacyRepo.UpdatePrivacyRequestStatus(ctx, requestID, "completed", adminID, nil); errStatus != nil {
			return pkgerrors.Wrapf(errStatus, "update privacy request status")
		}
		return nil
	}

	if s.Storage == nil {
		return fmt.Errorf("storage not configured")
	}

	if statusErr := s.PrivacyRepo.UpdatePrivacyRequestStatus(ctx, requestID, "processing", adminID, nil); statusErr != nil {
		return pkgerrors.Wrapf(err, "update privacy request status")
	}

	export := repository.DataExport{
		UserID:      req.UserID,
		RequestedBy: &adminID,
		Format:      "json",
		Status:      "processing",
	}
	createdExport, err := s.PrivacyRepo.CreateDataExport(ctx, export)
	if err != nil {
		errMsg := err.Error()
		_ = s.PrivacyRepo.UpdatePrivacyRequestStatus(ctx, requestID, "denied", adminID, &errMsg)
		return pkgerrors.Wrapf(err, "create data export")
	}

	go func(expID, userID, prID string) {
		bgCtx := context.Background()
		downloadURL, expiresAt, _, err := s.processExport(bgCtx, expID, userID)
		if err != nil {
			s.Log.Error("export_failed", map[string]any{
				"export_id":          expID,
				"user_id":            userID,
				"privacy_request_id": prID,
				"error":              err.Error(),
			})

			errMsg := err.Error()
			failedExp := repository.DataExport{
				ID:           expID,
				Status:       "failed",
				ErrorMessage: &errMsg,
			}
			_ = s.PrivacyRepo.UpdateDataExport(bgCtx, failedExp)
			_ = s.PrivacyRepo.UpdatePrivacyRequestStatus(bgCtx, prID, "denied", adminID, &errMsg)
			return
		}

		_ = s.PrivacyRepo.CompletePrivacyExportRequest(bgCtx, prID, adminID, downloadURL, expiresAt)
	}(createdExport.ID, *req.UserID, requestID)

	return nil
}

// ProcessDeletion performs the actual deletion (admin-initiated)
func (s *PrivacyService) ProcessDeletion(ctx context.Context, requestID, adminID string) error {
	// 1. Get request
	req, err := s.PrivacyRepo.GetPrivacyRequest(ctx, requestID)
	if err != nil {
		return pkgerrors.Wrapf(err, "get privacy request")
	}
	if req == nil {
		return fmt.Errorf("request not found")
	}
	if req.UserID == nil {
		return fmt.Errorf("user id is missing in request")
	}

	// 2. Cleanup user files from storage
	files, err := s.PrivacyRepo.GetUserFilesForDeletion(ctx, *req.UserID)
	switch {
	case err != nil:
		s.Log.Error("failed_to_get_user_files_for_deletion", map[string]any{
			"user_id": *req.UserID,
			"error":   err.Error(),
		})
		// We continue even if we fail to get files, to ensure DB anonymization happens
	case s.Storage == nil:
		s.Log.Info("storage_not_configured_skipping_user_file_deletions", map[string]any{
			"user_id": *req.UserID,
		})
		// Skip file deletion when storage is not configured
	default:
		for _, file := range files {
			if file == "" {
				continue
			}
			// Extract key from URL if needed (same logic as retention job)
			key := file
			if delErr := s.Storage.DeleteFile(ctx, key); delErr != nil {
				s.Log.Error("failed_to_delete_user_file", map[string]any{
					"key":   key,
					"error": delErr.Error(),
				})
			}
		}
	}

	// 3. Anonymize/Delete user data
	err = s.PrivacyRepo.AnonymizeUserData(ctx, *req.UserID)
	if err != nil {
		return pkgerrors.Wrapf(err, "anonymize user")
	}

	// 4. Update request status
	if errStatus := s.PrivacyRepo.UpdatePrivacyRequestStatus(ctx, requestID, "completed", adminID, nil); errStatus != nil {
		return pkgerrors.Wrapf(errStatus, "update privacy request status")
	}
	return nil
}
