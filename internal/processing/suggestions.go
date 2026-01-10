package processing

import (
	"context"
	"sort"
	"strings"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/logger"
)

const DefaultMaxSuggestions = 6

type SourceQuestions struct {
	Questions  []string
	ChunkCount int
}

func AggregateWithWeightedSelection(sources []SourceQuestions, existingQuestions []string, maxQuestions int) []string {
	if maxQuestions <= 0 {
		maxQuestions = DefaultMaxSuggestions
	}

	seen := make(map[string]struct{}, len(existingQuestions))
	result := make([]string, 0, maxQuestions)

	for _, q := range existingQuestions {
		t := normalizeQuestion(q)
		if t == "" {
			continue
		}
		seen[strings.ToLower(t)] = struct{}{}
		result = append(result, t)
		if len(result) >= maxQuestions {
			return result
		}
	}

	sort.Slice(sources, func(i, j int) bool {
		return sources[i].ChunkCount > sources[j].ChunkCount
	})

	for _, src := range sources {
		for _, q := range src.Questions {
			t := normalizeQuestion(q)
			if t == "" {
				continue
			}
			key := strings.ToLower(t)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, t)
			if len(result) >= maxQuestions {
				return result
			}
		}
	}

	return result
}

func normalizeQuestion(q string) string {
	t := strings.TrimSpace(q)
	if t == "" {
		return ""
	}
	if len(t) > 120 {
		t = t[:120]
	}
	return t
}

// AggregateAndPersistChatbotSuggestions aggregates and persists suggestions for a chatbot.
// Uses ChatbotRepository to update suggestions.
func AggregateAndPersistChatbotSuggestions(ctx context.Context, chatbotRepo repository.ChatbotRepository, chatbotID string, log *logger.Logger) {
	AggregateAndPersistChatbotSuggestionsWithLimit(ctx, chatbotRepo, chatbotID, DefaultMaxSuggestions, log)
}

// AggregateAndPersistChatbotSuggestionsWithLimit aggregates and persists suggestions with a custom limit.
func AggregateAndPersistChatbotSuggestionsWithLimit(ctx context.Context, chatbotRepo repository.ChatbotRepository, chatbotID string, maxQuestions int, log *logger.Logger) {
	if maxQuestions <= 0 {
		maxQuestions = DefaultMaxSuggestions
	}

	// Get existing chatbot to access current suggestions
	bot, err := chatbotRepo.GetByID(ctx, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "fetch_chatbot_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		return
	}
	if bot == nil {
		logWarnIfPresent(log, "chatbot_not_found", map[string]any{"chatbot_id": chatbotID})
		return
	}

	currentSuggestions := bot.SuggestedQuestions

	if len(currentSuggestions) >= maxQuestions {
		return
	}

	// Sources will be fetched in fetchSourceQuestions via repository
	sources, err := fetchSourceQuestions(ctx, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "fetch_source_questions_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		return
	}

	newSuggestions := AggregateWithWeightedSelection(sources, currentSuggestions, maxQuestions)

	if len(newSuggestions) != len(currentSuggestions) || !SlicesEqual(newSuggestions, currentSuggestions) {
		if err := chatbotRepo.UpdateSuggestedQuestions(ctx, chatbotID, newSuggestions); err != nil {
			logWarnIfPresent(log, "update_chatbot_suggestions_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		}
	}
}

// ReAggregateSuggestionsForChatbotWithJob re-aggregates suggestions for a chatbot with job tracking.
// Uses ChatbotRepository, SourceRepository, and SuggestionJobRepository to manage the process.
func ReAggregateSuggestionsForChatbotWithJob(ctx context.Context, chatbotRepo repository.ChatbotRepository, sourceRepo repository.SourceRepository, suggestionJobRepo repository.SuggestionJobRepository, chatbotID, jobID string, log *logger.Logger) {
	// Update job status to running
	if err := suggestionJobRepo.UpdateStatus(ctx, jobID, models.SuggestionJobStatusRunning); err != nil {
		logWarnIfPresent(log, "update_job_status_failed", map[string]any{"job_id": jobID, "error": err.Error()})
		_ = suggestionJobRepo.Fail(ctx, jobID, "Failed to update job status")
		return
	}

	maxQuestions := fetchMaxSuggestionsForChatbot(ctx, chatbotRepo, chatbotID)

	sources, err := fetchSourceQuestionsFromRepo(ctx, sourceRepo, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "reaggregate_fetch_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		_ = suggestionJobRepo.Fail(ctx, jobID, err.Error())
		return
	}

	newSuggestions := AggregateWithWeightedSelection(sources, nil, maxQuestions)

	if err := chatbotRepo.UpdateSuggestedQuestions(ctx, chatbotID, newSuggestions); err != nil {
		logWarnIfPresent(log, "reaggregate_update_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		_ = suggestionJobRepo.Fail(ctx, jobID, err.Error())
		return
	}

	if err := suggestionJobRepo.Complete(ctx, jobID, newSuggestions); err != nil {
		logWarnIfPresent(log, "complete_job_failed", map[string]any{"job_id": jobID, "error": err.Error()})
	}
}

// ReAggregateSuggestionsForChatbot re-aggregates suggestions for a chatbot.
// This is called after source processing completes to update the chatbot's suggested_questions
// from all available sources.
func ReAggregateSuggestionsForChatbot(ctx context.Context, chatbotRepo repository.ChatbotRepository, sourceRepo repository.SourceRepository, chatbotID string, log *logger.Logger) {
	logInfoIfPresent(log, "reaggregate_start", map[string]any{"chatbot_id": chatbotID})

	maxQuestions := fetchMaxSuggestionsForChatbot(ctx, chatbotRepo, chatbotID)
	logInfoIfPresent(log, "reaggregate_max_questions", map[string]any{"chatbot_id": chatbotID, "max_questions": maxQuestions})

	sources, err := fetchSourceQuestionsFromRepo(ctx, sourceRepo, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "reaggregate_fetch_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		return
	}

	// Log source count and total questions
	var totalSourceQuestions int
	for _, s := range sources {
		totalSourceQuestions += len(s.Questions)
	}
	logInfoIfPresent(log, "reaggregate_sources_fetched", map[string]any{
		"chatbot_id":             chatbotID,
		"source_count":           len(sources),
		"total_source_questions": totalSourceQuestions,
	})

	newSuggestions := AggregateWithWeightedSelection(sources, nil, maxQuestions)

	logInfoIfPresent(log, "reaggregate_result", map[string]any{
		"chatbot_id":            chatbotID,
		"new_suggestions_count": len(newSuggestions),
		"suggestions":           newSuggestions,
	})

	if err := chatbotRepo.UpdateSuggestedQuestions(ctx, chatbotID, newSuggestions); err != nil {
		logWarnIfPresent(log, "reaggregate_update_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		return
	}

	logInfoIfPresent(log, "reaggregate_success", map[string]any{"chatbot_id": chatbotID, "questions_count": len(newSuggestions)})
}

// fetchSourceQuestionsFromRepo retrieves source questions from the chatbot's sources using the repository.
func fetchSourceQuestionsFromRepo(ctx context.Context, sourceRepo repository.SourceRepository, chatbotID string) ([]SourceQuestions, error) {
	if sourceRepo == nil {
		return []SourceQuestions{}, nil
	}

	suggestions, err := sourceRepo.GetSourceSuggestions(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	result := make([]SourceQuestions, 0, len(suggestions))
	for _, s := range suggestions {
		result = append(result, SourceQuestions{
			Questions:  s.Questions,
			ChunkCount: s.ChunkCount,
		})
	}
	return result, nil
}

// fetchSourceQuestions retrieves source questions from the chatbot's sources.
// Deprecated: Use fetchSourceQuestionsFromRepo with a SourceRepository instead.
func fetchSourceQuestions(ctx context.Context, chatbotID string) ([]SourceQuestions, error) {
	// This is a legacy placeholder for backward compatibility
	// New code should use fetchSourceQuestionsFromRepo
	return []SourceQuestions{}, nil
}

// fetchMaxSuggestionsForChatbot fetches the max suggestions limit for a chatbot from its plan.
func fetchMaxSuggestionsForChatbot(ctx context.Context, chatbotRepo repository.ChatbotRepository, chatbotID string) int {
	return DefaultMaxSuggestions
}

func SlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func logWarnIfPresent(log *logger.Logger, event string, data map[string]any) {
	if log != nil {
		log.Warn(event, data)
	}
}

func logInfoIfPresent(log *logger.Logger, event string, data map[string]any) {
	if log != nil {
		log.Info(event, data)
	}
}

func DeleteSourceVectors(ctx context.Context, vc rag.VectorClient, sourceID string) error {
	if err := vc.DeleteBySourceID(ctx, sourceID); err != nil {
		return pkgerrors.Wrapf(err, "delete source vectors")
	}
	return nil
}
