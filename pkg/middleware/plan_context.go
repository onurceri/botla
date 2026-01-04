package middleware

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
)

type planContextKey struct{}

// PlanToContext stores the plan in the request context
func PlanToContext(ctx context.Context, plan *models.Plan) context.Context {
	return context.WithValue(ctx, planContextKey{}, plan)
}

// PlanFromContext retrieves the plan from the request context
func PlanFromContext(ctx context.Context) (*models.Plan, bool) {
	plan, ok := ctx.Value(planContextKey{}).(*models.Plan)
	return plan, ok
}
