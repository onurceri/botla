package services

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/processing"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
	"github.com/onurceri/botla-co/pkg/logger"
)

// RefreshPolicy constants
const (
	RefreshPolicyManual = "manual"
	RefreshPolicyAuto   = "auto"
)

// RefreshFrequency constants
const (
	RefreshFrequencyDaily   = "daily"
	RefreshFrequencyWeekly  = "weekly"
	RefreshFrequencyMonthly = "monthly"
)

// RefreshScheduler handles automatic refresh of URL sources
type RefreshScheduler struct {
	DB       *sql.DB
	Queue    *processing.SourceQueue
	Log      *logger.Logger
	interval time.Duration
	stopChan chan struct{}
	wg       sync.WaitGroup
	running  bool
	mu       sync.Mutex
}

// NewRefreshScheduler creates a new RefreshScheduler
func NewRefreshScheduler(db *sql.DB, queue *processing.SourceQueue, log *logger.Logger) *RefreshScheduler {
	return &RefreshScheduler{
		DB:       db,
		Queue:    queue,
		Log:      log,
		interval: 5 * time.Minute, // Check every 5 minutes
		stopChan: make(chan struct{}),
	}
}

// NewRefreshSchedulerWithInterval creates a new RefreshScheduler with a custom check interval
func NewRefreshSchedulerWithInterval(db *sql.DB, queue *processing.SourceQueue, log *logger.Logger, interval time.Duration) *RefreshScheduler {
	return &RefreshScheduler{
		DB:       db,
		Queue:    queue,
		Log:      log,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start begins the scheduler loop
func (s *RefreshScheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopChan = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run(ctx)

	if s.Log != nil {
		s.Log.Info("refresh_scheduler_started", map[string]any{"interval": s.interval.String()})
	}
}

// run is the main scheduler loop
func (s *RefreshScheduler) run(ctx context.Context) {
	defer s.wg.Done()
	// MI-002: Recover from panics to keep scheduler running
	defer func() {
		if r := recover(); r != nil && s.Log != nil {
			s.Log.Error("scheduler_panic", map[string]any{"panic": r})
		}
	}()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.safeProcessDueChatbots(ctx)
		}
	}
}

// safeProcessDueChatbots wraps processDueChatbots with panic recovery
func (s *RefreshScheduler) safeProcessDueChatbots(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil && s.Log != nil {
			s.Log.Error("process_due_chatbots_panic", map[string]any{"panic": r})
		}
	}()
	s.processDueChatbots(ctx)
}

// Stop halts the scheduler
func (s *RefreshScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopChan)
	s.mu.Unlock()

	s.wg.Wait()

	if s.Log != nil {
		s.Log.Info("refresh_scheduler_stopped", nil)
	}
}

// processDueChatbots finds and processes all chatbots due for refresh
func (s *RefreshScheduler) processDueChatbots(ctx context.Context) {
	now := time.Now()
	bots, err := s.FindDueForRefresh(ctx, now)
	if err != nil {
		s.logWarn("find_due_for_refresh_error", map[string]any{"error": err.Error()})
		return
	}

	if len(bots) > 0 {
		s.logInfo("refresh_due_found", map[string]any{"count": len(bots)})
	}

	for _, bot := range bots {
		if err := s.QueueRefreshForChatbot(ctx, &bot); err != nil {
			s.logWarn("queue_refresh_error", map[string]any{"bot_id": bot.ID, "error": err.Error()})
			continue
		}

		// Calculate next refresh time
		frequency := ""
		if bot.RefreshFrequency != nil {
			frequency = *bot.RefreshFrequency
		}
		nextRefresh := CalculateNextRefresh(frequency, now)

		// Update refresh times
		if err := db.UpdateChatbotRefreshTimes(ctx, s.DB, bot.ID, nextRefresh, now); err != nil {
			s.logWarn("update_refresh_times_error", map[string]any{"bot_id": bot.ID, "error": err.Error()})
		}

		s.logInfo("chatbot_refresh_scheduled", map[string]any{
			"bot_id":       bot.ID,
			"next_refresh": nextRefresh.Format(time.RFC3339),
		})
	}
}

// FindDueForRefresh finds chatbots with auto refresh enabled that are due
func (s *RefreshScheduler) FindDueForRefresh(ctx context.Context, now time.Time) ([]models.Chatbot, error) {
	bots, err := db.GetChatbotsDueForRefresh(ctx, s.DB, now)
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "getting due chatbots")
	}
	return bots, nil
}

// QueueRefreshForChatbot queues all URL sources for a chatbot for refresh
func (s *RefreshScheduler) QueueRefreshForChatbot(ctx context.Context, bot *models.Chatbot) error {
	// Get user's plan to check limits
	plan, err := db.GetPlanByUserID(ctx, s.DB, bot.UserID)
	if err != nil {
		return pkgerrors.Wrapf(err, "getting user plan")
	}
	if plan == nil {
		s.logWarn("user_plan_not_found", map[string]any{"user_id": bot.UserID})
		return nil
	}

	// Check monthly auto-refresh limit
	now := time.Now()
	usage, err := db.GetAutoRefreshCountForMonth(ctx, s.DB, bot.UserID, now)
	if err != nil {
		return pkgerrors.Wrapf(err, "getting auto refresh count")
	}

	// Get limit from plan config - use refresh.max_monthly if available
	limit := plan.Limits.RefreshMaxMonthly
	if usage >= limit {
		s.logInfo("auto_refresh_limit_reached", map[string]any{
			"user_id": bot.UserID,
			"usage":   usage,
			"limit":   limit,
		})
		return nil // Silently skip - limit reached
	}

	// Get URL sources for this chatbot
	sources, err := db.GetURLSourcesForChatbot(ctx, s.DB, bot.ID)
	if err != nil {
		return pkgerrors.Wrapf(err, "getting url sources")
	}

	if len(sources) == 0 {
		s.logInfo("no_url_sources_to_refresh", map[string]any{"bot_id": bot.ID})
		return nil
	}

	// Queue each source for refresh
	refreshedCount := 0
	for _, src := range sources {
		// Update source for refresh (sets status to pending)
		if err := db.UpdateSourceForRefresh(ctx, s.DB, src.ID); err != nil {
			s.logWarn("update_source_for_refresh_error", map[string]any{
				"source_id": src.ID,
				"error":     err.Error(),
			})
			continue
		}

		// Queue the source for processing
		if s.Queue != nil {
			s.Queue.Enqueue(src.ID)
		}
		refreshedCount++
	}

	if refreshedCount > 0 {
		// Increment auto-refresh count
		if err := db.IncrementAutoRefreshCount(ctx, s.DB, bot.UserID, now, 1); err != nil {
			s.logWarn("increment_auto_refresh_count_error", map[string]any{
				"user_id": bot.UserID,
				"error":   err.Error(),
			})
		}

		s.logInfo("sources_queued_for_refresh", map[string]any{
			"bot_id":          bot.ID,
			"refreshed_count": refreshedCount,
			"total_sources":   len(sources),
		})
	}

	return nil
}

// CalculateNextRefresh calculates the next refresh time based on frequency
func CalculateNextRefresh(frequency string, from time.Time) time.Time {
	switch frequency {
	case RefreshFrequencyDaily:
		// Next day at midnight
		next := from.Add(24 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, from.Location())
	case RefreshFrequencyWeekly:
		// Next Sunday at midnight
		daysUntilSunday := (7 - int(from.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7 // If today is Sunday, go to next Sunday
		}
		next := from.Add(time.Duration(daysUntilSunday) * 24 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, from.Location())
	case RefreshFrequencyMonthly:
		// First day of next month at midnight
		next := from.AddDate(0, 1, 0)
		return time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, from.Location())
	default:
		// Default to weekly if frequency is not specified
		daysUntilSunday := (7 - int(from.Weekday())) % 7
		if daysUntilSunday == 0 {
			daysUntilSunday = 7
		}
		next := from.Add(time.Duration(daysUntilSunday) * 24 * time.Hour)
		return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, from.Location())
	}
}

// CalculateInitialNextRefresh calculates the initial next_refresh_at when enabling auto refresh
func CalculateInitialNextRefresh(frequency string) time.Time {
	return CalculateNextRefresh(frequency, time.Now())
}

func (s *RefreshScheduler) logInfo(event string, data map[string]any) {
	if s.Log != nil {
		s.Log.Info(event, data)
	}
}

func (s *RefreshScheduler) logWarn(event string, data map[string]any) {
	if s.Log != nil {
		s.Log.Warn(event, data)
	}
}
