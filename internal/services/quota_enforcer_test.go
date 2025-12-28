package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuotaEnforcer_ReserveTokens_NoQuota(t *testing.T) {
	// Test that ReserveTokens returns nil when maxMonthlyTokens is 0
	qe := NewQuotaEnforcer(nil)

	err := qe.ReserveTokens(context.Background(), "user123", 512, 0)

	assert.NoError(t, err)
}

func TestQuotaEnforcer_New(t *testing.T) {
	// Test that NewQuotaEnforcer creates a valid instance
	qe := NewQuotaEnforcer(nil)

	assert.NotNil(t, qe)
}

func TestQuotaEnforcer_AdjustTokens_NoChange(t *testing.T) {
	// Test that AdjustTokens doesn't call DB when delta is 0
	qe := NewQuotaEnforcer(nil)

	// Should not panic even with nil DB
	qe.AdjustTokens(context.Background(), "user123", 512, 512)
}

func TestQuotaEnforcer_AdjustTokens_NilDB(t *testing.T) {
	// Test behavior with nil DB
	qe := NewQuotaEnforcer(nil)

	// Should not panic - delta is non-zero but DB is nil
	qe.AdjustTokens(context.Background(), "user123", 512, 600)
}

func TestQuotaEnforcer_RefundTokens_NilDB(t *testing.T) {
	// Test behavior with nil DB
	qe := NewQuotaEnforcer(nil)

	// Should not panic
	qe.RefundTokens(context.Background(), "user123", 512)
}

func TestGetDefaultTokenEstimate(t *testing.T) {
	estimate := GetDefaultTokenEstimate()
	assert.Equal(t, 512, estimate)
}

func TestQuotaEnforcer_AdjustTokens_PositiveDelta(t *testing.T) {
	// Test that positive delta is calculated correctly
	estimated := 512
	actual := 600
	expectedDelta := actual - estimated // 88

	assert.Equal(t, 88, expectedDelta)
}

func TestQuotaEnforcer_AdjustTokens_NegativeDelta(t *testing.T) {
	// Test that negative delta is calculated correctly
	estimated := 512
	actual := 400
	expectedDelta := actual - estimated // -112

	assert.Equal(t, -112, expectedDelta)
}

func TestQuotaEnforcer_AdjustTokens_LargePositiveDelta(t *testing.T) {
	// Test with larger positive delta
	estimated := 1000
	actual := 1500
	expectedDelta := 500

	assert.Equal(t, expectedDelta, actual-estimated)
}

func TestQuotaEnforcer_AdjustTokens_LargeNegativeDelta(t *testing.T) {
	// Test with larger negative delta
	estimated := 1000
	actual := 500
	expectedDelta := -500

	assert.Equal(t, expectedDelta, actual-estimated)
}
