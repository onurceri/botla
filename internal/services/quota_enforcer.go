package services

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/db"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
)

// QuotaEnforcer handles token quota reservation and adjustment for chat operations.
type QuotaEnforcer struct {
	DB *sql.DB
}

// NewQuotaEnforcer creates a new QuotaEnforcer.
func NewQuotaEnforcer(db *sql.DB) *QuotaEnforcer {
	return &QuotaEnforcer{DB: db}
}

// ReservationResult holds the result of a token reservation.
type ReservationResult struct {
	Reserved      bool
	EstimatedTokens int
	MaxMonthlyTokens int
}

// ReserveTokens reserves estimated tokens for a chat request.
// Returns nil if no quota enforcement is needed (maxMonthlyTokens <= 0).
func (q *QuotaEnforcer) ReserveTokens(ctx context.Context, userID string, estimatedTokens, maxMonthlyTokens int) error {
	if maxMonthlyTokens <= 0 {
		return nil // No quota enforcement
	}
	err := db.ReserveChatTokens(ctx, q.DB, userID, estimatedTokens, maxMonthlyTokens)
	if err != nil {
		if err == db.ErrTokenQuotaExceeded {
			return ErrTokenQuotaExceeded
		}
		return pkgerrors.Wrapf(err, "reserve tokens")
	}
	return nil
}

// AdjustTokens adjusts token usage after chat completion.
func (q *QuotaEnforcer) AdjustTokens(ctx context.Context, userID string, estimatedTokens, actualTokens int) {
	if q.DB == nil {
		return
	}
	delta := actualTokens - estimatedTokens
	if delta != 0 {
		_ = db.AdjustChatTokens(ctx, q.DB, userID, delta)
	}
}

// RefundTokens refunds reserved tokens on error.
func (q *QuotaEnforcer) RefundTokens(ctx context.Context, userID string, tokens int) {
	if q.DB == nil {
		return
	}
	_ = db.AdjustChatTokens(ctx, q.DB, userID, -tokens)
}

// GetDefaultTokenEstimate returns a default token estimate when not specified.
func GetDefaultTokenEstimate() int {
	return 512
}
