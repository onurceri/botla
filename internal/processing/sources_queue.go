package processing

import (
	"context"
	"database/sql"
	"encoding/json"

	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/pdf"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/internal/text"
	"github.com/onurceri/botla-co/pkg/storage"
	"github.com/onurceri/botla-co/pkg/logger"
)

type SourceQueue struct {
    ch           *chan string
    db           *sql.DB
    storage      storage.StorageService
    openaiClient rag.LLMClient
    log          *logger.Logger
}

func StartSourceQueue(dbpool *sql.DB, st storage.StorageService) (*SourceQueue, error) {
	c := make(chan string, 64)
	oa, err := rag.NewOpenAIClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create openai client: %w", err)
	}
    q := &SourceQueue{ch: &c, db: dbpool, storage: st, openaiClient: oa, log: logger.New("INFO")}
	go q.worker()
	// Ensure collection exists at startup (best-effort)
	if qc, err := rag.NewQdrantClientFromEnv(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = qc.EnsureEmbeddingsCollection(ctx)
		cancel()
	}
	return q, nil
}

func (q *SourceQueue) Enqueue(id string) {
	if q == nil || q.ch == nil {
		return
	}
	select {
	case *q.ch <- id:
	default:
		// drop if full
	}
}

func (q *SourceQueue) worker() {
	if q.ch == nil {
		return
	}
	for id := range *q.ch {
		q.markProcessing(id)
		s, langCode, ok := q.loadSourceAndLang(id)
		if !ok {
			continue
		}
		switch s.SourceType {
		case "url":
			q.processURLSource(id, s, langCode)
		case "pdf":
			q.processPDFSource(id, s, langCode)
		case "text":
			q.processTextSource(id, s, langCode)
		default:
			q.fail(id, "unsupported_type")
		}
	}
}

func defaultLang(code string) string {
	if strings.TrimSpace(code) == "" {
		return "tr"
	}
	return code
}

func (q *SourceQueue) markProcessing(id string) {
    if q.log != nil { q.log.Info("source_processing_start", map[string]any{"source_id": id}) }
    _ = db.UpdateSourceProcessing(context.Background(), q.db, id, "processing", nil, 0, nil)
}

func (q *SourceQueue) fail(id string, msg string) {
    if q.log != nil { q.log.Warn("source_processing_fail", map[string]any{"source_id": id, "reason": msg}) }
    _ = db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &msg, 0, nil)
}

func (q *SourceQueue) complete(id string, chunks int) {
    if q.log != nil { q.log.Info("source_processing_complete", map[string]any{"source_id": id, "chunks": chunks}) }
    now := time.Now()
    _ = db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, chunks, &now)
}

func (q *SourceQueue) loadSourceAndLang(id string) (*models.DataSource, string, bool) {
	s, err := db.GetSourceByID(context.Background(), q.db, id)
	if err != nil || s == nil {
		q.fail(id, "source_not_found")
		return nil, "", false
	}
	bot, err := db.GetChatbotByID(context.Background(), q.db, s.ChatbotID)
	if err != nil || bot == nil {
		q.fail(id, "chatbot_not_found")
		return nil, "", false
	}
	return s, defaultLang(bot.Language), true
}

func (q *SourceQueue) persistIngestionMetadata(content string, langCode string, s *models.DataSource) {
	if meta, err := rag.ExtractIngestionMetadata(context.Background(), q.openaiClient, content, langCode); err == nil {
		_ = db.UpdateSourceCapability(context.Background(), q.db, s.ID, meta.CapabilitySummary)
		_ = db.UpdateSourceSuggestions(context.Background(), q.db, s.ID, meta.SuggestedQuestions)
		aggregateAndPersistChatbotSuggestions(context.Background(), q.db, s.ChatbotID)
	}
}

func (q *SourceQueue) processURLSource(id string, s *models.DataSource, langCode string) {
	if s.SourceURL == nil || *s.SourceURL == "" {
		q.fail(id, "empty_url")
		return
	}
	content, err := scraper.ScrapeURLWithFallback(
		scraper.ScrapingTask{URL: *s.SourceURL},
		scraper.DefaultCollectorConfig(),
	)
	if err != nil {
		q.fail(id, err.Error())
		return
	}
	if content == "" {
		q.complete(id, 0)
		return
	}
	content = text.NormalizeTR(content)
	q.persistIngestionMetadata(content, langCode, s)
	rc, rerr := rag.ChunkText(content, 512, langCode)
	if rerr != nil {
		q.fail(id, rerr.Error())
		return
	}
	if err := rag.GenerateEmbeddingsForSource(rc, s.ChatbotID, s.ID, s.SourceType); err != nil {
		q.fail(id, err.Error())
		return
	}
	q.complete(id, len(rc))
}

func (q *SourceQueue) processPDFSource(id string, s *models.DataSource, langCode string) {
	if s.FilePath == nil || *s.FilePath == "" {
		q.fail(id, "empty_file_path")
		return
	}
	localPath := *s.FilePath
	if q.storage != nil {
		rc, err := q.storage.DownloadFile(context.Background(), *s.FilePath)
		if err != nil {
			q.fail(id, err.Error())
			return
		}
		tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
		if err != nil {
			_ = rc.Close()
			q.fail(id, err.Error())
			return
		}
		_, err = io.Copy(tmpFile, rc)
		_ = rc.Close()
		_ = tmpFile.Close()
		if err != nil {
			q.fail(id, err.Error())
			return
		}
		localPath = tmpFile.Name()
		go func(p string) { _ = os.Remove(p) }(localPath)
	}
	content, err := pdf.ExtractPDFText(localPath, langCode)
	if err != nil {
		q.fail(id, err.Error())
		return
	}
	if content == "" {
		q.complete(id, 0)
		return
	}
	content = text.NormalizeTR(content)
	q.persistIngestionMetadata(content, langCode, s)
	rc, rerr := rag.ChunkText(content, 512, langCode)
	if rerr != nil {
		q.fail(id, rerr.Error())
		return
	}
	if err := rag.GenerateEmbeddingsForSource(rc, s.ChatbotID, s.ID, s.SourceType); err != nil {
		q.fail(id, err.Error())
		return
	}
	q.complete(id, len(rc))
}

func (q *SourceQueue) processTextSource(id string, s *models.DataSource, langCode string) {
	if s.FilePath == nil || *s.FilePath == "" {
		q.fail(id, "empty_file_path")
		return
	}
	var content string
	if q.storage != nil {
		rc, err := q.storage.DownloadFile(context.Background(), *s.FilePath)
		if err != nil {
			q.fail(id, err.Error())
			return
		}
		b, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			q.fail(id, err.Error())
			return
		}
		content = string(b)
	} else {
		q.fail(id, "storage_required")
		return
	}
	if content == "" {
		q.complete(id, 0)
		return
	}
	content = text.NormalizeTR(content)
	q.persistIngestionMetadata(content, langCode, s)
	rc, rerr := rag.ChunkText(content, 512, langCode)
	if rerr != nil {
		q.fail(id, rerr.Error())
		return
	}
	if err := rag.GenerateEmbeddingsForSource(rc, s.ChatbotID, s.ID, s.SourceType); err != nil {
		q.fail(id, err.Error())
		return
	}
	q.complete(id, len(rc))
}

// aggregateAndPersistChatbotSuggestions queries data_sources for the chatbot and writes unique suggestions to chatbots.suggested_questions.
func aggregateAndPersistChatbotSuggestions(ctx context.Context, pool *sql.DB, chatbotID string) {
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

	rows, err := pool.QueryContext(ctx, `SELECT suggested_questions FROM data_sources WHERE chatbot_id=$1 AND suggested_questions IS NOT NULL`, chatbotID)
	if err != nil {
		return
	}
	defer func() { _ = rows.Close() }()

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
			if len(out) >= 6 {
				break
			}
		}
		if len(out) >= 6 {
			break
		}
	}

	// Only update if we have changes or if we just want to ensure consistency
	_ = db.UpdateChatbotSuggestions(ctx, pool, chatbotID, out)
}

func DeleteSourceVectors(ctx context.Context, sourceID string) error {
	qc, err := rag.NewQdrantClientFromEnv()
	if err != nil {
		return err
	}
	return qc.DeleteBySourceID(ctx, sourceID)
}
