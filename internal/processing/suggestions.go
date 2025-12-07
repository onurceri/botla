package processing

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
)

// AggregateAndPersistChatbotSuggestions queries data_sources for the chatbot
// and writes unique suggestions to chatbots.suggested_questions.
func AggregateAndPersistChatbotSuggestions(ctx context.Context, pool *sql.DB, chatbotID string, log *logger.Logger) {
	// Fetch existing chatbot suggestions
	var existing []byte
	var currentSuggestions []string
	_ = pool.QueryRowContext(ctx, `SELECT suggested_questions FROM chatbots WHERE id=$1`, chatbotID).Scan(&existing)

	if len(existing) > 0 {
		_ = json.Unmarshal(existing, &currentSuggestions)
	}

	// Initialize set with existing suggestions to avoid duplicates
	uniq := map[string]struct{}{}
	out := make([]string, 0, 6)

	for _, q := range currentSuggestions {
		t := strings.TrimSpace(q)
		if t == "" {
			continue
		}
		if len(t) > 120 {
			t = t[:120]
		}
		k := strings.ToLower(t)
		if _, ok := uniq[k]; ok {
			continue
		}
		uniq[k] = struct{}{}
		out = append(out, t)
	}

	// If we already have 6 or more questions, we don't need to fetch more
	if len(out) >= 6 {
		return
	}

	rows, err := pool.QueryContext(ctx, `SELECT suggested_questions FROM data_sources WHERE chatbot_id=$1 AND suggested_questions IS NOT NULL AND deleted_at IS NULL`, chatbotID)
	if err != nil {
		return
	}
	defer func() { _ = rows.Close() }()

	foundNew := false
	for rows.Next() {
		var arr []byte
		if err := rows.Scan(&arr); err != nil {
			continue
		}
		var items []string
		if err := json.Unmarshal(arr, &items); err != nil {
			continue
		}
		for _, it := range items {
			t := strings.TrimSpace(it)
			if t == "" {
				continue
			}
			if len(t) > 120 {
				t = t[:120]
			}
			k := strings.ToLower(t)
			if _, ok := uniq[k]; ok {
				continue
			}
			uniq[k] = struct{}{}
			out = append(out, t)
			foundNew = true
			if len(out) >= 6 {
				break
			}
		}
		if len(out) >= 6 {
			break
		}
	}

	// Only update if we have changes
	if foundNew || len(out) != len(currentSuggestions) {
		err := db.UpdateChatbotSuggestions(ctx, pool, chatbotID, out)
		if err != nil && log != nil {
			log.Warn("update_chatbot_suggestions_failed", map[string]any{"chatbot_id": chatbotID, "error": err.Error()})
		}
	}
}

// DeleteSourceVectors deletes vectors associated with a source from Qdrant
func DeleteSourceVectors(ctx context.Context, sourceID string) error {
	qc, err := rag.NewQdrantClientFromEnv()
	if err != nil {
		return err
	}
	return qc.DeleteBySourceID(ctx, sourceID)
}
