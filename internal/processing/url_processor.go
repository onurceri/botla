package processing

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/internal/text"
	"github.com/onurceri/botla-co/pkg/logger"
)

// URLProcessor handles URL source processing
type URLProcessor struct {
	DB           *sql.DB
	OpenAIClient rag.LLMClient
	Log          *logger.Logger
}

// NewURLProcessor creates a new URLProcessor
func NewURLProcessor(db *sql.DB, oai rag.LLMClient, log *logger.Logger) *URLProcessor {
	return &URLProcessor{
		DB:           db,
		OpenAIClient: oai,
		Log:          log,
	}
}

// ProcessResult contains the result of source processing
type ProcessResult struct {
	ChunkCount int
	Skipped    bool
	Error      error
}

// Process processes a URL source
func (p *URLProcessor) Process(ctx context.Context, s *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan) ProcessResult {
	if s.SourceURL == nil || *s.SourceURL == "" {
		return ProcessResult{Error: &ProcessingError{Msg: "empty_url"}}
	}

	// Create scrape config with CSS selectors if defined
	var scrapeConfig *scraper.ScrapeConfig
	if len(bot.SelectorWhitelist) > 0 {
		scrapeConfig = &scraper.ScrapeConfig{
			Selectors: bot.SelectorWhitelist,
		}
	}

	content, err := scraper.ScrapeURLWithFallback(
		scraper.ScrapingTask{URL: *s.SourceURL},
		scraper.DefaultCollectorConfig(),
		plan.Config.Scraping.DynamicEnabled,
		scrapeConfig,
	)
	if err != nil {
		return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
	}
	if content == "" {
		return ProcessResult{ChunkCount: 0}
	}

	// Crawler Logic - discover sub-pages
	p.discoverSubPages(ctx, s, bot, plan, content)

	content = text.NormalizeTR(content)

	// Compute hash of normalized content
	hashBytes := sha256.Sum256([]byte(content))
	newHash := hex.EncodeToString(hashBytes[:])

	// Check if content has changed (for refresh - skip re-embedding if unchanged)
	if s.Hash != nil && *s.Hash == newHash {
		p.logInfo("source_content_unchanged", map[string]any{"source_id": s.ID, "url": *s.SourceURL})
		return ProcessResult{ChunkCount: s.ChunkCount, Skipped: true}
	}

	// Content changed or new source - delete old vectors first if this is a refresh
	if s.Hash != nil {
		if err := DeleteSourceVectors(ctx, s.ID); err != nil {
			p.logWarn("source_delete_vectors_error", map[string]any{"source_id": s.ID, "error": err.Error()})
		}
	}

	// Extract and persist metadata
	p.persistIngestionMetadata(ctx, content, langCode, s)

	// Chunk and embed
	rc, rerr := rag.ChunkText(content, 512, langCode)
	if rerr != nil {
		return ProcessResult{Error: &ProcessingError{Msg: rerr.Error()}}
	}
	if err := rag.GenerateEmbeddingsForSource(rc, s.ChatbotID, s.ID, s.SourceType); err != nil {
		return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
	}

	// Update hash after successful embedding
	_ = db.UpdateSourceHash(ctx, p.DB, s.ID, newHash)

	// Calculate token usage
	var tokens int
	for _, ch := range rc {
		tokens += ch.TokenCount
	}
	_ = db.IncrementSuccessfulIngestion(ctx, p.DB, bot.UserID, time.Now(), 1)
	_ = db.AddEmbeddingTokens(ctx, p.DB, bot.UserID, time.Now(), tokens)

	return ProcessResult{ChunkCount: len(rc)}
}

// DiscoveryMode constants for URL discovery behavior
const (
	DiscoveryModeAuto     = "auto"     // Default: automatically add discovered URLs as sources
	DiscoveryModePending  = "pending"  // Add to pending table for user approval
	DiscoveryModeDisabled = "disabled" // Do not discover sub-pages
)

// discoverSubPages crawls and discovers sub-pages from the content
func (p *URLProcessor) discoverSubPages(ctx context.Context, s *models.DataSource, bot *models.Chatbot, plan *models.Plan, content string) {
	// Check discovery mode - default to auto for backward compatibility
	discoveryMode := bot.DiscoveryMode
	if discoveryMode == "" {
		discoveryMode = DiscoveryModeAuto
	}

	// If disabled, skip discovery entirely
	if discoveryMode == DiscoveryModeDisabled {
		return
	}

	if plan.Config.Scraping.MaxPagesPerCrawl <= 0 {
		return
	}

	// Only crawl if we haven't reached the limit
	cnt, err := db.CountSourcesByType(ctx, p.DB, s.ChatbotID, "url")
	if err != nil || cnt >= plan.Config.Scraping.MaxPagesPerCrawl+plan.Config.Scraping.MaxURLsPerBot {
		return
	}

	// Create path filter from chatbot settings
	var filter *scraper.PathFilter
	if len(bot.IncludePaths) > 0 || len(bot.ExcludePaths) > 0 {
		filter, err = scraper.NewPathFilter(bot.IncludePaths, bot.ExcludePaths)
		if err != nil {
			p.logWarn("path_filter_creation_failed", map[string]any{
				"source_id": s.ID,
				"error":     err.Error(),
			})
			// Continue without filter rather than failing completely
			filter = nil
		}
	}

	// Extract links with filter
	links, lerr := scraper.ExtractLinks(content, *s.SourceURL, filter)
	if lerr != nil {
		return
	}

	added := 0
	limit := plan.Config.Scraping.MaxPagesPerCrawl

	for _, link := range links {
		if added >= limit {
			break
		}

		switch discoveryMode {
		case DiscoveryModeAuto:
			// Current behavior: directly create source
			exists, _ := db.SourceExists(ctx, p.DB, s.ChatbotID, link)
			if !exists {
				_, cerr := db.CreateSource(ctx, p.DB, s.ChatbotID, "url", &link, nil, nil)
				if cerr == nil {
					added++
				}
			}
		case DiscoveryModePending:
			// New behavior: add to pending table for approval
			// Check if already exists as source or in pending
			exists, _ := db.SourceExists(ctx, p.DB, s.ChatbotID, link)
			if !exists {
				sourceID := s.ID
				if err := db.InsertPendingURL(ctx, p.DB, s.ChatbotID, &sourceID, link); err == nil {
					added++
				}
			}
		}
	}

	if added > 0 {
		p.logInfo("urls_discovered", map[string]any{
			"source_id":      s.ID,
			"mode":           discoveryMode,
			"discovered":     added,
			"total_found":    len(links),
		})
	}
}

// persistIngestionMetadata extracts and saves metadata for the source
func (p *URLProcessor) persistIngestionMetadata(ctx context.Context, content, langCode string, s *models.DataSource) {
	meta, err := rag.ExtractIngestionMetadata(ctx, p.OpenAIClient, content, langCode)
	if err != nil {
		p.logWarn("extract_metadata_failed", map[string]any{"source_id": s.ID, "error": err.Error()})
		return
	}

	if len(meta.SuggestedQuestions) == 0 {
		p.logWarn("extract_metadata_empty_questions", map[string]any{"source_id": s.ID})
	} else {
		p.logInfo("extract_metadata_success", map[string]any{
			"source_id":       s.ID,
			"questions_count": len(meta.SuggestedQuestions),
			"questions":       meta.SuggestedQuestions,
		})
	}

	if err := db.UpdateSourceCapability(ctx, p.DB, s.ID, meta.CapabilitySummary); err != nil {
		p.logWarn("update_source_capability_failed", map[string]any{"source_id": s.ID, "error": err.Error()})
	}
	if err := db.UpdateSourceSuggestions(ctx, p.DB, s.ID, meta.SuggestedQuestions); err != nil {
		p.logWarn("update_source_suggestions_failed", map[string]any{"source_id": s.ID, "error": err.Error()})
	}

	go AggregateAndPersistChatbotSuggestions(ctx, p.DB, s.ChatbotID, p.Log)
}

func (p *URLProcessor) logInfo(event string, data map[string]any) {
	if p.Log != nil {
		p.Log.Info(event, data)
	}
}

func (p *URLProcessor) logWarn(event string, data map[string]any) {
	if p.Log != nil {
		p.Log.Warn(event, data)
	}
}

// ProcessingError represents a processing error
type ProcessingError struct {
	Msg string
}

func (e *ProcessingError) Error() string { return e.Msg }
