package processing

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
	"github.com/onurceri/botla-co/pkg/logger"
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

func AggregateAndPersistChatbotSuggestions(ctx context.Context, pool *sql.DB, chatbotID string, log *logger.Logger) {
	AggregateAndPersistChatbotSuggestionsWithLimit(ctx, pool, chatbotID, DefaultMaxSuggestions, log)
}

func AggregateAndPersistChatbotSuggestionsWithLimit(ctx context.Context, pool *sql.DB, chatbotID string, maxQuestions int, log *logger.Logger) {
	if maxQuestions <= 0 {
		maxQuestions = DefaultMaxSuggestions
	}

	var existingJSON []byte
	var currentSuggestions []string
	_ = pool.QueryRowContext(ctx, `SELECT suggested_questions FROM chatbots WHERE id=$1`, chatbotID).Scan(&existingJSON)
	if len(existingJSON) > 0 {
		_ = json.Unmarshal(existingJSON, &currentSuggestions)
	}

	if len(currentSuggestions) >= maxQuestions {
		return
	}

	sources, err := fetchSourceQuestions(ctx, pool, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "fetch_source_questions_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		return
	}

	newSuggestions := AggregateWithWeightedSelection(sources, currentSuggestions, maxQuestions)

	if len(newSuggestions) != len(currentSuggestions) || !SlicesEqual(newSuggestions, currentSuggestions) {
		if err := db.UpdateChatbotSuggestedQuestions(ctx, pool, chatbotID, newSuggestions); err != nil {
			logWarnIfPresent(log, "update_chatbot_suggestions_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		}
	}
}

func ReAggregateSuggestionsForChatbotWithJob(ctx context.Context, pool *sql.DB, chatbotID, jobID string, log *logger.Logger) {
	if err := db.UpdateSuggestionJobStatus(ctx, pool, jobID, models.SuggestionJobStatusRunning); err != nil {
		logWarnIfPresent(log, "update_job_status_failed", map[string]any{"job_id": jobID, "error": err.Error()})
		_ = db.FailSuggestionJob(ctx, pool, jobID, "Failed to update job status")
		return
	}

	maxQuestions := fetchMaxSuggestionsForChatbot(ctx, pool, chatbotID)

	sources, err := fetchSourceQuestions(ctx, pool, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "reaggregate_fetch_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		_ = db.FailSuggestionJob(ctx, pool, jobID, err.Error())
		return
	}

	newSuggestions := AggregateWithWeightedSelection(sources, nil, maxQuestions)

	if err := db.UpdateChatbotSuggestedQuestions(ctx, pool, chatbotID, newSuggestions); err != nil {
		logWarnIfPresent(log, "reaggregate_update_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		_ = db.FailSuggestionJob(ctx, pool, jobID, err.Error())
		return
	}

	if err := db.CompleteSuggestionJob(ctx, pool, jobID, newSuggestions); err != nil {
		logWarnIfPresent(log, "complete_job_failed", map[string]any{"job_id": jobID, "error": err.Error()})
	}
}

func ReAggregateSuggestionsForChatbot(ctx context.Context, pool *sql.DB, chatbotID string, log *logger.Logger) {
	maxQuestions := fetchMaxSuggestionsForChatbot(ctx, pool, chatbotID)

	sources, err := fetchSourceQuestions(ctx, pool, chatbotID)
	if err != nil {
		logWarnIfPresent(log, "reaggregate_fetch_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		return
	}

	newSuggestions := AggregateWithWeightedSelection(sources, nil, maxQuestions)

	if err := db.UpdateChatbotSuggestedQuestions(ctx, pool, chatbotID, newSuggestions); err != nil {
		logWarnIfPresent(log, "reaggregate_update_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
	}
}

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
		return nil, pkgerrors.Wrapf(err, "query source questions")
	}
	defer func() {
		_ = rows.Close()
	}()

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
	if err := rows.Err(); err != nil {
		return nil, pkgerrors.Wrapf(err, "source questions rows err")
	}
	return sources, nil
}

func fetchMaxSuggestionsForChatbot(ctx context.Context, pool *sql.DB, chatbotID string) int {
	var limit int
	err := pool.QueryRowContext(ctx, `
		SELECT COALESCE(pl.chat_max_suggested_questions, $2)
		FROM chatbots c
		JOIN users u ON c.user_id = u.id
		JOIN plans p ON u.plan_id = p.id
		LEFT JOIN plan_limits pl ON pl.plan_id = p.id
		WHERE c.id = $1
	`, chatbotID, DefaultMaxSuggestions).Scan(&limit)
	if err != nil || limit <= 0 {
		return DefaultMaxSuggestions
	}
	return limit
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

func DeleteSourceVectors(ctx context.Context, vc rag.VectorClient, sourceID string) error {
	if err := vc.DeleteBySourceID(ctx, sourceID); err != nil {
		return pkgerrors.Wrapf(err, "delete source vectors")
	}
	return nil
}
