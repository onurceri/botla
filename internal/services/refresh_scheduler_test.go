package services

import (
	"testing"
	"time"
)

func TestCalculateNextRefresh_Daily(t *testing.T) {
	// Test from a Wednesday at 10:30 AM
	from := time.Date(2024, 12, 5, 10, 30, 0, 0, time.UTC)
	next := CalculateNextRefresh(RefreshFrequencyDaily, from)
	
	// Should be next day at midnight
	expected := time.Date(2024, 12, 6, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Daily: expected %v, got %v", expected, next)
	}
}

func TestCalculateNextRefresh_Daily_BeforeMidnight(t *testing.T) {
	// Test from 11:59 PM
	from := time.Date(2024, 12, 5, 23, 59, 0, 0, time.UTC)
	next := CalculateNextRefresh(RefreshFrequencyDaily, from)
	
	// Should be December 6th at midnight
	expected := time.Date(2024, 12, 6, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Daily before midnight: expected %v, got %v", expected, next)
	}
}

func TestCalculateNextRefresh_Weekly_FromWednesday(t *testing.T) {
	// Wednesday, December 4, 2024
	from := time.Date(2024, 12, 4, 10, 30, 0, 0, time.UTC)
	next := CalculateNextRefresh(RefreshFrequencyWeekly, from)
	
	// Should be next Sunday (December 8, 2024) at midnight
	expected := time.Date(2024, 12, 8, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Weekly from Wednesday: expected %v (Sunday), got %v (weekday %v)", 
			expected, next, next.Weekday())
	}
}

func TestCalculateNextRefresh_Weekly_FromSunday(t *testing.T) {
	// Sunday, December 8, 2024
	from := time.Date(2024, 12, 8, 10, 30, 0, 0, time.UTC)
	next := CalculateNextRefresh(RefreshFrequencyWeekly, from)
	
	// Should be NEXT Sunday (December 15, 2024), not the same Sunday
	expected := time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Weekly from Sunday: expected %v (next Sunday), got %v", expected, next)
	}
}

func TestCalculateNextRefresh_Weekly_FromSaturday(t *testing.T) {
	// Saturday, December 7, 2024
	from := time.Date(2024, 12, 7, 22, 0, 0, 0, time.UTC)
	next := CalculateNextRefresh(RefreshFrequencyWeekly, from)
	
	// Should be next Sunday (December 8, 2024)
	expected := time.Date(2024, 12, 8, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Weekly from Saturday: expected %v (Sunday), got %v", expected, next)
	}
}

func TestCalculateNextRefresh_Monthly_MidMonth(t *testing.T) {
	// December 15, 2024
	from := time.Date(2024, 12, 15, 10, 30, 0, 0, time.UTC)
	next := CalculateNextRefresh(RefreshFrequencyMonthly, from)
	
	// Should be January 1, 2025
	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Monthly mid-month: expected %v, got %v", expected, next)
	}
}

func TestCalculateNextRefresh_Monthly_EndOfMonth(t *testing.T) {
	// December 31, 2024
	from := time.Date(2024, 12, 31, 23, 59, 0, 0, time.UTC)
	next := CalculateNextRefresh(RefreshFrequencyMonthly, from)
	
	// Should be January 1, 2025
	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Monthly end of month: expected %v, got %v", expected, next)
	}
}

func TestCalculateNextRefresh_Monthly_FirstOfMonth(t *testing.T) {
	// December 1, 2024
	from := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	next := CalculateNextRefresh(RefreshFrequencyMonthly, from)
	
	// Should be January 1, 2025
	expected := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Monthly first of month: expected %v, got %v", expected, next)
	}
}

func TestCalculateNextRefresh_Default(t *testing.T) {
	// Empty frequency should default to weekly
	from := time.Date(2024, 12, 4, 10, 30, 0, 0, time.UTC) // Wednesday
	next := CalculateNextRefresh("", from)
	
	// Should behave like weekly - next Sunday
	expected := time.Date(2024, 12, 8, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Default (empty) frequency: expected %v (weekly behavior), got %v", expected, next)
	}
}

func TestCalculateNextRefresh_InvalidFrequency(t *testing.T) {
	// Invalid frequency should default to weekly
	from := time.Date(2024, 12, 4, 10, 30, 0, 0, time.UTC) // Wednesday
	next := CalculateNextRefresh("invalid", from)
	
	// Should behave like weekly - next Sunday
	expected := time.Date(2024, 12, 8, 0, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Invalid frequency: expected %v (weekly behavior), got %v", expected, next)
	}
}

func TestCalculateInitialNextRefresh(t *testing.T) {
	// Just verify it doesn't panic and returns a future time
	for _, freq := range []string{RefreshFrequencyDaily, RefreshFrequencyWeekly, RefreshFrequencyMonthly} {
		next := CalculateInitialNextRefresh(freq)
		if next.Before(time.Now()) {
			t.Errorf("CalculateInitialNextRefresh(%s) returned past time: %v", freq, next)
		}
	}
}

func TestRefreshPolicyConstants(t *testing.T) {
	if RefreshPolicyManual != "manual" {
		t.Errorf("RefreshPolicyManual should be 'manual', got '%s'", RefreshPolicyManual)
	}
	if RefreshPolicyAuto != "auto" {
		t.Errorf("RefreshPolicyAuto should be 'auto', got '%s'", RefreshPolicyAuto)
	}
}

func TestRefreshFrequencyConstants(t *testing.T) {
	if RefreshFrequencyDaily != "daily" {
		t.Errorf("RefreshFrequencyDaily should be 'daily', got '%s'", RefreshFrequencyDaily)
	}
	if RefreshFrequencyWeekly != "weekly" {
		t.Errorf("RefreshFrequencyWeekly should be 'weekly', got '%s'", RefreshFrequencyWeekly)
	}
	if RefreshFrequencyMonthly != "monthly" {
		t.Errorf("RefreshFrequencyMonthly should be 'monthly', got '%s'", RefreshFrequencyMonthly)
	}
}

func TestNewRefreshScheduler(t *testing.T) {
	scheduler := NewRefreshScheduler(nil, nil, nil)
	
	if scheduler.interval != 5*time.Minute {
		t.Errorf("Default interval should be 5 minutes, got %v", scheduler.interval)
	}
	if scheduler.stopChan == nil {
		t.Error("stopChan should be initialized")
	}
}

func TestNewRefreshSchedulerWithInterval(t *testing.T) {
	customInterval := 10 * time.Minute
	scheduler := NewRefreshSchedulerWithInterval(nil, nil, nil, customInterval)
	
	if scheduler.interval != customInterval {
		t.Errorf("Custom interval should be %v, got %v", customInterval, scheduler.interval)
	}
}
