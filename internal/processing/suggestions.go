package processing

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
)

// DefaultMaxSuggestions is the fallback limit when plan config is not available.
const DefaultMaxSuggestions = 6

// SourceQuestions represents questions from a source with its weight (chunk count).
type SourceQuestions struct {
	Questions  []string
	ChunkCount int
}

// AggregateWithWeightedSelection selects questions from sources weighted by chunk count.
// Sources with more chunks get priority. Returns up to maxQuestions unique questions.
func AggregateWithWeightedSelection(sources []SourceQuestions, existingQuestions []string, maxQuestions int) []string {
	if maxQuestions <= 0 {
		maxQuestions = DefaultMaxSuggestions
	}

	// Build seen set from existing questions
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

	// Sort sources by chunk count descending (larger sources first)
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].ChunkCount > sources[j].ChunkCount
	})

	// Collect questions from sorted sources
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

// normalizeQuestion trims, validates, and caps question length.
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

// AggregateAndPersistChatbotSuggestions aggregates suggestions from all sources.
// Uses DefaultMaxSuggestions when limit is not specified.
func AggregateAndPersistChatbotSuggestions(ctx context.Context, pool *sql.DB, chatbotID string, log *logger.Logger) {
	AggregateAndPersistChatbotSuggestionsWithLimit(ctx, pool, chatbotID, DefaultMaxSuggestions, log)
}

// AggregateAndPersistChatbotSuggestionsWithLimit queries data_sources for the chatbot
// and writes unique suggestions to chatbots.suggested_questions respecting the limit.
func AggregateAndPersistChatbotSuggestionsWithLimit(ctx context.Context, pool *sql.DB, chatbotID string, maxQuestions int, log *logger.Logger) {
	if maxQuestions <= 0 {
		maxQuestions = DefaultMaxSuggestions
	}

	// Fetch existing chatbot suggestions (manual questions take priority)
	var existingJSON []byte
	var currentSuggestions []string
	_ = pool.QueryRowContext(ctx, `SELECT suggested_questions FROM chatbots WHERE id=$1`, chatbotID).Scan(&existingJSON)
	if len(existingJSON) > 0 {
		_ = json.Unmarshal(existingJSON, &currentSuggestions)
	}

	// If already at limit, nothing to do
	if len(currentSuggestions) >= maxQuestions {
		return
	}

	// Fetch source questions with chunk counts
	sources, err := fetchSourceQuestions(ctx, pool, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "fetch_source_questions_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		return
	}

	// Aggregate with weighted selection
	newSuggestions := AggregateWithWeightedSelection(sources, currentSuggestions, maxQuestions)

	// Only update if changed
	if len(newSuggestions) != len(currentSuggestions) || !slicesEqual(newSuggestions, currentSuggestions) {
		if err := db.UpdateChatbotSuggestions(ctx, pool, chatbotID, newSuggestions); err != nil {
			logWarnIfPresent(log, "update_chatbot_suggestions_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		}
	}
}

// ReAggregateSuggestionsForChatbot re-aggregates suggestions after source changes.
// Call this when a source is deleted or updated.
func ReAggregateSuggestionsForChatbot(ctx context.Context, pool *sql.DB, chatbotID string, log *logger.Logger) {
	// Fetch plan limit for this chatbot
	maxQuestions := fetchMaxSuggestionsForChatbot(ctx, pool, chatbotID)

	// Clear existing auto-generated suggestions and rebuild
	sources, err := fetchSourceQuestions(ctx, pool, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "reaggregate_fetch_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		return
	}

	// Rebuild from scratch (no existing questions)
	newSuggestions := AggregateWithWeightedSelection(sources, nil, maxQuestions)

	if err := db.UpdateChatbotSuggestions(ctx, pool, chatbotID, newSuggestions); err != nil {
		logWarnIfPresent(log, "reaggregate_update_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
	}
}

// fetchSourceQuestions retrieves all source questions with their chunk counts.
func fetchSourceQuestions(ctx context.Context, pool *sql.DB, chatbotID string) ([]SourceQuestions, error) {
	rows, err := pool.QueryContext(ctx, `
		SELECT suggested_questions, chunk_count 
		FROM data_sources 
		WHERE chatbot_id=$1 
		  AND suggested_questions IS NOT NULL 
		  AND deleted_at IS NULL
		ORDER BY chunk_count DESC
	`, chatbotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []SourceQuestions
	for rows.Next() {
		var qJSON []byte
		var chunkCount int
		if err := rows.Scan(&qJSON, &chunkCount); err != nil {
			continue
		}
		var questions []string
		if err := json.Unmarshal(qJSON, &questions); err != nil {
			continue
		}
		if len(questions) > 0 {
			sources = append(sources, SourceQuestions{
				Questions:  questions,
				ChunkCount: chunkCount,
			})
		}
	}
	return sources, rows.Err()
}

// fetchMaxSuggestionsForChatbot gets the plan-based limit for suggestions.
func fetchMaxSuggestionsForChatbot(ctx context.Context, pool *sql.DB, chatbotID string) int {
	var limit int
	err := pool.QueryRowContext(ctx, `
		SELECT COALESCE((p.config->'chat'->>'max_suggested_questions')::int, $2)
		FROM chatbots c
		JOIN users u ON c.user_id = u.id
		JOIN plans p ON u.plan_id = p.id
		WHERE c.id = $1
	`, chatbotID, DefaultMaxSuggestions).Scan(&limit)
	if err != nil || limit <= 0 {
		return DefaultMaxSuggestions
	}
	return limit
}

// slicesEqual checks if two string slices are equal.
func slicesEqual(a, b []string) bool {
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

// logWarnIfPresent logs a warning if logger is available.
func logWarnIfPresent(log *logger.Logger, event string, data map[string]any) {
	if log != nil {
		log.Warn(event, data)
	}
}

// DeleteSourceVectors deletes vectors associated with a source from Qdrant.
func DeleteSourceVectors(ctx context.Context, sourceID string) error {
	qc, err := rag.NewQdrantClientFromEnv()
	if err != nil {
		return err
	}
	return qc.DeleteBySourceID(ctx, sourceID)
}
