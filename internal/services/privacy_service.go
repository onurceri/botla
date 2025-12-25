package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

type PrivacyService struct {
	DB      *sql.DB
	Log     *logger.Logger
	Storage storage.StorageService
}

func NewPrivacyService(db *sql.DB, log *logger.Logger, storage storage.StorageService) *PrivacyService {
	return &PrivacyService{
		DB:      db,
		Log:     log,
		Storage: storage,
	}
}

// ExportUserData generates a JSON export of all user data
func (s *PrivacyService) ExportUserData(ctx context.Context, userID string, requestedBy string) (*db.DataExport, error) {
	if s.Storage == nil {
		return nil, fmt.Errorf("storage not configured")
	}

	// 1. Create data_exports record (pending)
	export := db.DataExport{
		UserID:      &userID,
		RequestedBy: &requestedBy,
		Format:      "json",
		Status:      "processing",
	}
	createdExport, err := db.CreateDataExport(ctx, s.DB, export)
	if err != nil {
		return nil, err
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
			failedExp := db.DataExport{
				ID:           expID,
				Status:       "failed",
				ErrorMessage: &errMsg,
			}
			_ = db.UpdateDataExport(bgCtx, s.DB, failedExp)
		}
	}(createdExport.ID)

	return createdExport, nil
}

func (s *PrivacyService) processExport(ctx context.Context, exportID, userID string) (string, time.Time, int64, error) {
	if s.Storage == nil {
		return "", time.Time{}, 0, fmt.Errorf("storage not configured")
	}

	// 1. Collect all user data using the DB helper
	exportData, err := db.GetUserDataForExport(ctx, s.DB, userID)
	if err != nil {
		return "", time.Time{}, 0, fmt.Errorf("get user data for export: %w", err)
	}

	// 2. Generate JSON file
	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", time.Time{}, 0, fmt.Errorf("marshal json: %w", err)
	}

	// 3. Upload to storage
	filename := fmt.Sprintf("exports/%s/%s-%d.json", userID, exportID, time.Now().Unix())
	fileURL, err := s.Storage.UploadFile(ctx, filename, bytes.NewReader(jsonData))
	if err != nil {
		return "", time.Time{}, 0, fmt.Errorf("upload file: %w", err)
	}

	// 4. Update data_exports record
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour) // 7 days expiration
	size := int64(len(jsonData))

	update := db.DataExport{
		ID:            exportID,
		Status:        "completed",
		DownloadURL:   &fileURL,
		FileSizeBytes: &size,
		ExpiresAt:     &expiresAt,
		CompletedAt:   &now,
	}

	if err := db.UpdateDataExport(ctx, s.DB, update); err != nil {
		return "", time.Time{}, 0, err
	}

	return fileURL, expiresAt, size, nil
}

// RequestDeletion initiates user data deletion process
func (s *PrivacyService) RequestDeletion(ctx context.Context, userID, email, reason string) (*db.PrivacyRequest, error) {
	req := db.PrivacyRequest{
		UserID:      &userID,
		UserEmail:   email,
		RequestType: "deletion",
		Status:      "pending",
		Reason:      reason,
	}
	return db.CreatePrivacyRequest(ctx, s.DB, req)
}

func (s *PrivacyService) RequestExport(ctx context.Context, userID, email, reason string) (*db.PrivacyRequest, error) {
	req := db.PrivacyRequest{
		UserID:      &userID,
		UserEmail:   email,
		RequestType: "export",
		Status:      "pending",
		Reason:      reason,
	}
	return db.CreatePrivacyRequest(ctx, s.DB, req)
}

func (s *PrivacyService) RequestCorrection(ctx context.Context, userID, email, reason string) (*db.PrivacyRequest, error) {
	req := db.PrivacyRequest{
		UserID:      &userID,
		UserEmail:   email,
		RequestType: "correction",
		Status:      "pending",
		Reason:      reason,
	}
	return db.CreatePrivacyRequest(ctx, s.DB, req)
}

func (s *PrivacyService) ProcessExportRequest(ctx context.Context, requestID, adminID string) error {
	req, err := db.GetPrivacyRequest(ctx, s.DB, requestID)
	if err != nil {
		return err
	}
	if req == nil {
		return fmt.Errorf("request not found")
	}
	if req.RequestType != "export" {
		return fmt.Errorf("request is not an export request")
	}

	if req.UserID == nil {
		return db.UpdatePrivacyRequestStatus(ctx, s.DB, requestID, "completed", adminID, nil)
	}

	if s.Storage == nil {
		return fmt.Errorf("storage not configured")
	}

	if statusErr := db.UpdatePrivacyRequestStatus(ctx, s.DB, requestID, "processing", adminID, nil); statusErr != nil {
		return statusErr
	}

	export := db.DataExport{
		UserID:      req.UserID,
		RequestedBy: &adminID,
		Format:      "json",
		Status:      "processing",
	}
	createdExport, err := db.CreateDataExport(ctx, s.DB, export)
	if err != nil {
		errMsg := err.Error()
		_ = db.UpdatePrivacyRequestStatus(ctx, s.DB, requestID, "denied", adminID, &errMsg)
		return err
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
			failedExp := db.DataExport{
				ID:           expID,
				Status:       "failed",
				ErrorMessage: &errMsg,
			}
			_ = db.UpdateDataExport(bgCtx, s.DB, failedExp)
			_ = db.UpdatePrivacyRequestStatus(bgCtx, s.DB, prID, "denied", adminID, &errMsg)
			return
		}

		_ = db.CompletePrivacyExportRequest(bgCtx, s.DB, prID, adminID, downloadURL, expiresAt)
	}(createdExport.ID, *req.UserID, requestID)

	return nil
}

// ProcessDeletion performs the actual deletion (admin-initiated)
func (s *PrivacyService) ProcessDeletion(ctx context.Context, requestID, adminID string) error {
	// 1. Get request
	req, err := db.GetPrivacyRequest(ctx, s.DB, requestID)
	if err != nil {
		return err
	}
	if req == nil {
		return fmt.Errorf("request not found")
	}
	if req.UserID == nil {
		return fmt.Errorf("user id is missing in request")
	}

	// 2. Cleanup user files from storage
	files, err := db.GetUserFilesForDeletion(ctx, s.DB, *req.UserID)
	if err != nil {
		s.Log.Error("failed_to_get_user_files_for_deletion", map[string]any{
			"user_id": *req.UserID,
			"error":   err.Error(),
		})
		// We continue even if we fail to get files, to ensure DB anonymization happens
	} else {
		if s.Storage == nil {
			s.Log.Info("storage_not_configured_skipping_user_file_deletions", map[string]any{
				"user_id": *req.UserID,
			})
			goto anonymizeUser
		}

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
anonymizeUser:
	err = db.AnonymizeUserData(ctx, s.DB, *req.UserID)
	if err != nil {
		return fmt.Errorf("anonymize user: %w", err)
	}

	// 4. Update request status
	return db.UpdatePrivacyRequestStatus(ctx, s.DB, requestID, "completed", adminID, nil)
}
