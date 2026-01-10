package middleware

import (
	"context"
	"net/http"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/pkg/logger"
)

// DeletedAccountMiddleware checks if the user's account has been deleted
// This middleware should run AFTER AuthMiddleware and BEFORE other protected handlers
func DeletedAccountMiddleware(userRepo interface {
	GetByID(ctx context.Context, id string) (*models.User, error)
}, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Only check for authenticated requests
			userID, ok := UserIDFromContext(ctx)
			if !ok || userID == "" {
				// No user ID, skip check
				next.ServeHTTP(w, r)
				return
			}

			log.Info("deleted_account_check", map[string]any{
				"user_id": userID,
				"path":    r.URL.Path,
			})

			// Check if user exists (deleted users won't be found)
			user, err := userRepo.GetByID(ctx, userID)
			if err != nil {
				log.Error("failed_to_check_user", map[string]any{
					"error":   err.Error(),
					"user_id": userID,
				})
				// Continue on error - let other handlers deal with it
				next.ServeHTTP(w, r)
				return
			}

			if user == nil {
				// User not found - account was deleted
				log.Info("account_deleted_detected", map[string]any{
					"user_id": userID,
				})
				api.WriteErrorCode(w, http.StatusForbidden, api.ErrAccountDeleted)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
