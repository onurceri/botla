package middleware

import (
	"database/sql"
	"net/http"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/pkg/logger"
)

// PlanLoaderMiddleware loads the user's plan and stores it in the request context
// This middleware should run AFTER AuthMiddleware so the user ID is available
func PlanLoaderMiddleware(database *sql.DB, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			// Only load plan for authenticated requests
			userID, ok := UserIDFromContext(ctx)
			if !ok || userID == "" {
				// No user ID, skip plan loading
				next.ServeHTTP(w, r)
				return
			}
			
			// Fetch the user's plan from database
			plan, err := db.GetPlanByUserID(ctx, database, userID)
			if err != nil {
				log.Error("failed_to_load_plan", map[string]any{
					"error":   err.Error(),
					"user_id": userID,
				})
				// Continue without plan - will fall back to global rate limits
				next.ServeHTTP(w, r)
				return
			}
			
			if plan == nil {
				log.Warn("user_has_no_plan", map[string]any{
					"user_id": userID,
				})
				// Continue without plan - will fall back to global rate limits
				next.ServeHTTP(w, r)
				return
			}
			
			// Store plan in context for use by rate limiter and other middleware
			ctx = PlanToContext(ctx, plan)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
