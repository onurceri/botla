package services

import (
	"context"

	"github.com/onurceri/botla-app/internal/repository"
)

// QuotaEnforcer handles token quota reservation and adjustment for chat operations.
type QuotaEnforcer struct {
	usageRepo repository.UsageRepository
}

// NewQuotaEnforcer creates a new QuotaEnforcer.
func NewQuotaEnforcer(usageRepo repository.UsageRepository) *QuotaEnforcer {
	return &QuotaEnforcer{usageRepo: usageRepo}
}

// ReservationResult holds the result of a token reservation.
type ReservationResult struct {
	Reserved         bool
	EstimatedTokens  int
	MaxMonthlyTokens int
}

// ReserveTokens reserves estimated tokens for a chat request.
// Returns nil if no quota enforcement is needed (maxMonthlyTokens <= 0).
func (q *QuotaEnforcer) ReserveTokens(ctx context.Context, userID string, estimatedTokens, maxMonthlyTokens int) error {
	if maxMonthlyTokens <= 0 {
		return nil
	}
	if q.usageRepo == nil {
		return nil
	}
	err := q.usageRepo.ReserveChatTokens(ctx, userID, estimatedTokens, maxMonthlyTokens)
	if err != nil {
		return err
	}
	return nil
}

// AdjustTokens adjusts token usage after chat completion.
func (q *QuotaEnforcer) AdjustTokens(ctx context.Context, userID string, estimatedTokens, actualTokens int) {
	if q.usageRepo == nil {
		return
	}
	delta := actualTokens - estimatedTokens
	if delta != 0 {
		_ = q.usageRepo.AdjustChatTokens(ctx, userID, delta)
	}
}

// RefundTokens refunds reserved tokens on error.
func (q *QuotaEnforcer) RefundTokens(ctx context.Context, userID string, tokens int) {
	if q.usageRepo == nil {
		return
	}
	_ = q.usageRepo.AdjustChatTokens(ctx, userID, -tokens)
}

// GetDefaultTokenEstimate returns a default token estimate when not specified.
func GetDefaultTokenEstimate() int {
	return 512
}
