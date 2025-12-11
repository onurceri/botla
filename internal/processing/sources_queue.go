package processing

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

// SourceQueue manages background processing of data sources
type SourceQueue struct {
	ch           chan string
	stopCh       chan struct{}
	db           *sql.DB
	storage      storage.StorageService
	openaiClient rag.LLMClient
	log          *logger.Logger

	// Processors
	urlProcessor  *URLProcessor
	pdfProcessor  *PDFProcessor
	textProcessor *TextProcessor
}

// StartSourceQueue creates and starts a new source processing queue
func StartSourceQueue(dbpool *sql.DB, st storage.StorageService) (*SourceQueue, error) {
	c := make(chan string, 64)
	stop := make(chan struct{})

	oa, err := rag.NewOpenAIClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create openai client: %w", err)
	}

	log := logger.New("INFO")

	q := &SourceQueue{
		ch:           c,
		stopCh:       stop,
		db:           dbpool,
		storage:      st,
		openaiClient: oa,
		log:          log,

		// Initialize processors
		urlProcessor:  NewURLProcessor(dbpool, oa, log),
		pdfProcessor:  NewPDFProcessor(dbpool, st, oa, log),
		textProcessor: NewTextProcessor(dbpool, st, oa, log),
	}

	go q.worker()

	// Ensure collection exists at startup (best-effort)
	if qc, err := rag.NewQdrantClientFromEnv(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_ = qc.EnsureEmbeddingsCollection(ctx)
		cancel()
	}

	// Recover pending sources at startup
	go q.recoverPendingSources()

	return q, nil
}

// Enqueue adds a source ID to the processing queue
func (q *SourceQueue) Enqueue(id string) {
	if q == nil || q.ch == nil {
		return
	}
	select {
	case q.ch <- id:
	default:
		// drop if full
		if q.log != nil {
			q.log.Warn("source_queue_full", map[string]any{"dropped_id": id})
		}
	}
}

// Stop gracefully shuts down the queue worker
func (q *SourceQueue) Stop() {
	if q == nil || q.stopCh == nil {
		return
	}
	close(q.stopCh)
}

// recoverPendingSources finds and enqueues sources stuck in 'pending' status at startup
func (q *SourceQueue) recoverPendingSources() {
	defer func() {
		if r := recover(); r != nil {
			if q.log != nil {
				q.log.Error("recover_pending_sources_panic", map[string]any{"panic": r})
			}
		}
	}()

	if q == nil || q.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if DB is reachable to avoid panic in QueryContext with uninitialized DB
	if err := q.db.PingContext(ctx); err != nil {
		if q.log != nil {
			q.log.Warn("recover_pending_sources_db_unreachable", map[string]any{"error": err.Error()})
		}
		return
	}

	// Find sources with pending status (stuck from previous runs)
	rows, err := q.db.QueryContext(ctx, `
		SELECT id FROM data_sources 
		WHERE status = 'pending' AND deleted_at IS NULL 
		ORDER BY created_at ASC
		LIMIT 100
	`)
	if err != nil {
		if q.log != nil {
			q.log.Warn("recover_pending_sources_query_failed", map[string]any{"error": err.Error()})
		}
		return
	}
	defer func() { _ = rows.Close() }()

	var recovered int
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		q.Enqueue(id)
		recovered++
	}

	if recovered > 0 && q.log != nil {
		q.log.Info("recover_pending_sources_completed", map[string]any{
			"recovered_count": recovered,
		})
	}
}

// worker processes sources from the queue
func (q *SourceQueue) worker() {
	if q.ch == nil {
		return
	}
	for {
		select {
		case <-q.stopCh:
			if q.log != nil {
				q.log.Info("source_queue_shutdown", nil)
			}
			return
		case id := <-q.ch:
			q.processSource(id)
		}
	}
}

// processSource handles processing of a single source
func (q *SourceQueue) processSource(id string) {
	q.markProcessing(id)

	s, bot, langCode, plan, ok := q.loadSourceAndLang(id)
	if !ok {
		return
	}

	if q.log != nil {
		q.log.Info("source_processing_dispatch", map[string]any{
			"source_id":   id,
			"source_type": s.SourceType,
			"chatbot_id":  s.ChatbotID,
			"lang_code":   langCode,
		})
	}

	var result ProcessResult

	switch s.SourceType {
	case "url":
		result = q.urlProcessor.Process(context.Background(), s, bot, langCode, plan)
		// Enqueue discovered sub-pages (URL processor creates them but doesn't enqueue)
		q.enqueueNewURLSources(s.ChatbotID)
	case "pdf":
		result = q.pdfProcessor.Process(context.Background(), s, bot, langCode, plan)
	case "text":
		result = q.textProcessor.Process(context.Background(), s, bot, langCode)
	default:
		q.fail(id, "unsupported_type")
		return
	}

	if result.Error != nil {
		q.fail(id, result.Error.Error())
		return
	}

	// Log warning when source completes with 0 chunks
	if result.ChunkCount == 0 && !result.Skipped {
		if q.log != nil {
			q.log.Warn("source_processing_empty_content", map[string]any{
				"source_id":   id,
				"source_type": s.SourceType,
				"hint":        "Source completed but extracted 0 chunks - content may be empty or inaccessible",
			})
		}
	}

	q.complete(id, result.ChunkCount)
}

// enqueueNewURLSources discovers and enqueues pending URL sources
func (q *SourceQueue) enqueueNewURLSources(chatbotID string) {
	ctx := context.Background()
	sources, err := db.ListSourcesByChatbotID(ctx, q.db, chatbotID)
	if err != nil {
		return
	}
	for _, s := range sources {
		if s.SourceType == "url" && s.Status == "pending" {
			q.Enqueue(s.ID)
		}
	}
}

// loadSourceAndLang loads source, chatbot, and plan data
func (q *SourceQueue) loadSourceAndLang(id string) (*models.DataSource, *models.Chatbot, string, *models.Plan, bool) {
	ctx := context.Background()

	s, err := db.GetSourceByID(ctx, q.db, id)
	if err != nil || s == nil {
		q.fail(id, "source_not_found")
		return nil, nil, "", nil, false
	}

	bot, err := db.GetChatbotByID(ctx, q.db, s.ChatbotID)
	if err != nil || bot == nil {
		q.fail(id, "chatbot_not_found")
		return nil, nil, "", nil, false
	}

	plan, err := db.GetPlanByUserID(ctx, q.db, bot.UserID)
	if err != nil {
		q.fail(id, "plan_error")
		return nil, nil, "", nil, false
	}

	// Fallback to empty plan if nil
	if plan == nil {
		plan = &models.Plan{}
	}

	return s, bot, defaultLang(bot.LanguageCode), plan, true
}

// markProcessing marks a source as processing
func (q *SourceQueue) markProcessing(id string) {
	if q.log != nil {
		q.log.Info("source_processing_start", map[string]any{"source_id": id})
	}
	_ = db.UpdateSourceProcessing(context.Background(), q.db, id, "processing", nil, 0, nil)
}

// fail marks a source as failed
func (q *SourceQueue) fail(id string, msg string) {
	if q.log != nil {
		q.log.Warn("source_processing_fail", map[string]any{"source_id": id, "reason": msg})
	}
	_ = db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &msg, 0, nil)
}

// complete marks a source as completed
func (q *SourceQueue) complete(id string, chunks int) {
	if q.log != nil {
		q.log.Info("source_processing_complete", map[string]any{"source_id": id, "chunks": chunks})
	}
	now := time.Now()
	_ = db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, chunks, &now)
}

// defaultLang extracts base language code
func defaultLang(code string) string {
	s := strings.TrimSpace(code)
	if s == "" {
		return "tr"
	}
	if i := strings.Index(s, "-"); i > 0 {
		s = s[:i]
	}
	return s
}
