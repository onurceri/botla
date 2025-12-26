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
	VectorClient rag.VectorClient
	Log          *logger.Logger
	Scraper      scraper.Scraper
}

// NewURLProcessor creates a new URLProcessor.
// If scraper is nil, a DefaultScraper is used.
func NewURLProcessor(db *sql.DB, oai rag.LLMClient, vc rag.VectorClient, log *logger.Logger, s scraper.Scraper) *URLProcessor {
	if s == nil {
		s = scraper.NewDefaultScraper()
	}
	return &URLProcessor{
		DB:           db,
		OpenAIClient: oai,
		VectorClient: vc,
		Log:          log,
		Scraper:      s,
	}
}

// ProcessResult contains the result of source processing
type ProcessResult struct {
	ChunkCount   int
	Skipped      bool
	Error        error
	NewSourceIDs []string
}

// Process processes a URL source
func (p *URLProcessor) Process(ctx context.Context, s *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan) ProcessResult {
	if s.SourceURL == nil || *s.SourceURL == "" {
		p.logWarn("url_processing_empty_url", map[string]any{"source_id": s.ID})
		return ProcessResult{Error: &ProcessingError{Msg: "empty_url"}}
	}

	p.logInfo("url_processing_started", map[string]any{
		"source_id":      s.ID,
		"url":            *s.SourceURL,
		"discovery_mode": bot.DiscoveryMode,
		"chatbot_id":     s.ChatbotID,
	})

	// Create scrape config with CSS selectors if defined
	var scrapeConfig *scraper.ScrapeConfig
	if len(bot.SelectorWhitelist) > 0 {
		scrapeConfig = &scraper.ScrapeConfig{
			Selectors: bot.SelectorWhitelist,
		}
		p.logInfo("url_processing_selectors_configured", map[string]any{
			"source_id": s.ID,
			"selectors": bot.SelectorWhitelist,
		})
	}

	// Step 1: Fetch raw HTML for link discovery (always, regardless of text extraction)
	rawHTML, htmlErr := p.Scraper.FetchRawHTML(*s.SourceURL, scraper.DefaultCollectorConfig())
	switch {
	case htmlErr != nil:
		p.logWarn("url_processing_fetch_html_failed", map[string]any{
			"source_id": s.ID,
			"url":       *s.SourceURL,
			"error":     htmlErr.Error(),
		})
	case rawHTML != "":
		p.logInfo("url_processing_html_fetched", map[string]any{
			"source_id":  s.ID,
			"html_bytes": len(rawHTML),
		})
		// Attempt sub-page discovery using raw HTML
		p.discoverSubPages(ctx, s, bot, plan, rawHTML)
	default:
		p.logWarn("url_processing_html_empty", map[string]any{
			"source_id": s.ID,
			"url":       *s.SourceURL,
		})
	}

	// Step 2: Extract text content for embedding
	content, err := p.Scraper.ScrapeURLWithFallback(
		scraper.ScrapingTask{URL: *s.SourceURL},
		scraper.DefaultCollectorConfig(),
		plan.Config.Scraping.DynamicEnabled,
		scrapeConfig,
	)
	if err != nil {
		p.logWarn("url_processing_scrape_failed", map[string]any{
			"source_id": s.ID,
			"url":       *s.SourceURL,
			"error":     err.Error(),
		})
		return ProcessResult{Error: &ProcessingError{Msg: err.Error()}}
	}

	// Step 3: Handle empty content case
	if content == "" {
		p.logWarn("url_processing_content_empty", map[string]any{
			"source_id":       s.ID,
			"url":             *s.SourceURL,
			"dynamic_enabled": plan.Config.Scraping.DynamicEnabled,
			"selectors_used":  len(bot.SelectorWhitelist) > 0,
			"hint":            "Website may require JavaScript rendering or has anti-bot protection",
		})
		// Return 0 chunks but not an error - discovery may have still succeeded
		return ProcessResult{ChunkCount: 0}
	}

	p.logInfo("url_processing_content_extracted", map[string]any{
		"source_id":     s.ID,
		"content_bytes": len(content),
	})

	content = text.NormalizeTR(content)

	// Compute hash of normalized content
	hashBytes := sha256.Sum256([]byte(content))
	newHash := hex.EncodeToString(hashBytes[:])

	// Check if content has changed (for refresh - skip re-embedding if unchanged)
	// Also ensure we have chunks - if count is 0, we should reprocess even if hash matches
	if s.Hash != nil && *s.Hash == newHash && s.ChunkCount > 0 {
		p.logInfo("url_processing_content_unchanged", map[string]any{
			"source_id": s.ID,
			"url":       *s.SourceURL,
			"hash":      newHash[:16], // First 16 chars of hash for brevity
		})
		return ProcessResult{ChunkCount: s.ChunkCount, Skipped: true}
	}

	// Content changed or new source - delete old vectors first if this is a refresh
	if s.Hash != nil {
		p.logInfo("url_processing_content_changed", map[string]any{
			"source_id": s.ID,
			"old_hash":  (*s.Hash)[:16],
			"new_hash":  newHash[:16],
		})
		if err := DeleteSourceVectors(ctx, p.VectorClient, s.ID); err != nil {
			p.logWarn("url_processing_delete_vectors_error", map[string]any{
				"source_id": s.ID,
				"error":     err.Error(),
			})
		}
	}

	// Extract and persist metadata
	maxQuestions := 0
	if plan != nil && plan.Config.Chat.MaxSuggestedQuestions > 0 {
		maxQuestions = plan.Config.Chat.MaxSuggestedQuestions
	}
	p.persistIngestionMetadata(ctx, content, langCode, s, maxQuestions)

	// Chunk and embed
	p.logInfo("url_processing_chunking_started", map[string]any{
		"source_id":     s.ID,
		"content_bytes": len(content),
		"lang_code":     langCode,
	})

	rc, rerr := rag.ChunkText(content, 512, langCode)
	if rerr != nil {
		p.logWarn("url_processing_chunking_failed", map[string]any{
			"source_id": s.ID,
			"error":     rerr.Error(),
		})
		return ProcessResult{Error: &ProcessingError{Msg: rerr.Error()}}
	}

	p.logInfo("url_processing_embedding_started", map[string]any{
		"source_id":   s.ID,
		"chunk_count": len(rc),
	})

	emb, ok := p.OpenAIClient.(rag.EmbeddingClient)
	if !ok {
		return ProcessResult{Error: &ProcessingError{Msg: "llm_client_does_not_support_embeddings"}}
	}

	if err := rag.GenerateEmbeddingsForSource(ctx, emb, p.VectorClient, rc, s.ChatbotID, s.ID, s.SourceType); err != nil {
		p.logWarn("url_processing_embedding_failed", map[string]any{
			"source_id": s.ID,
			"error":     err.Error(),
		})
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

	p.logInfo("url_processing_completed", map[string]any{
		"source_id":    s.ID,
		"url":          *s.SourceURL,
		"chunk_count":  len(rc),
		"total_tokens": tokens,
	})

	return ProcessResult{ChunkCount: len(rc)}
}

// DiscoveryMode constants for URL discovery behavior
const (
	DiscoveryModeAuto     = "auto"     // Default: automatically add discovered URLs as sources
	DiscoveryModePending  = "pending"  // Add to pending table for user approval
	DiscoveryModeDisabled = "disabled" // Do not discover sub-pages
)

// discoverSubPages crawls and discovers sub-pages from the raw HTML content
func (p *URLProcessor) discoverSubPages(ctx context.Context, s *models.DataSource, bot *models.Chatbot, plan *models.Plan, rawHTML string) []string {
	var newIDs []string
	// Skip discovery for sources that were themselves discovered via crawling
	// This implements 1-level depth crawling (only user-added URLs discover sub-pages)
	if s.IsDiscovered {
		p.logInfo("url_discovery_skipped_is_discovered", map[string]any{
			"source_id": s.ID,
			"url":       *s.SourceURL,
			"reason":    "Source was discovered via crawling, not manually added",
		})
		return nil
	}

	// Check discovery mode - default to auto
	discoveryMode := bot.DiscoveryMode
	if discoveryMode == "" {
		discoveryMode = DiscoveryModeAuto
	}

	p.logInfo("url_discovery_started", map[string]any{
		"source_id":      s.ID,
		"url":            *s.SourceURL,
		"discovery_mode": discoveryMode,
		"html_length":    len(rawHTML),
	})

	// If disabled, skip discovery entirely
	if discoveryMode == DiscoveryModeDisabled {
		p.logInfo("url_discovery_disabled", map[string]any{
			"source_id": s.ID,
		})
		return nil
	}

	if plan.Config.Scraping.MaxPagesPerCrawl <= 0 {
		p.logInfo("url_discovery_skipped_no_crawl_limit", map[string]any{
			"source_id":           s.ID,
			"max_pages_per_crawl": plan.Config.Scraping.MaxPagesPerCrawl,
		})
		return nil
	}

	// Only crawl if we haven't reached the limit
	cnt, err := db.CountSourcesByType(ctx, p.DB, s.ChatbotID, "url")
	if err != nil {
		p.logWarn("url_discovery_count_sources_failed", map[string]any{
			"source_id": s.ID,
			"error":     err.Error(),
		})
		return nil
	}

	maxTotal := plan.Config.Scraping.MaxPagesPerCrawl + plan.Config.Scraping.MaxURLsPerBot
	if cnt >= maxTotal {
		p.logInfo("url_discovery_skipped_limit_reached", map[string]any{
			"source_id":     s.ID,
			"current_count": cnt,
			"max_total":     maxTotal,
		})
		return nil
	}

	// Create path filter from chatbot settings
	var filter *scraper.PathFilter
	if len(bot.IncludePaths) > 0 || len(bot.ExcludePaths) > 0 {
		filter, err = scraper.NewPathFilter(bot.IncludePaths, bot.ExcludePaths)
		if err != nil {
			p.logWarn("url_discovery_path_filter_failed", map[string]any{
				"source_id":     s.ID,
				"include_paths": bot.IncludePaths,
				"exclude_paths": bot.ExcludePaths,
				"error":         err.Error(),
			})
			// Continue without filter rather than failing completely
			filter = nil
		} else {
			p.logInfo("url_discovery_path_filter_created", map[string]any{
				"source_id":     s.ID,
				"include_paths": bot.IncludePaths,
				"exclude_paths": bot.ExcludePaths,
			})
		}
	}

	// Extract links with filter
	links, lerr := p.Scraper.ExtractLinks(rawHTML, *s.SourceURL, filter)
	if lerr != nil {
		p.logWarn("url_discovery_extract_links_failed", map[string]any{
			"source_id": s.ID,
			"url":       *s.SourceURL,
			"error":     lerr.Error(),
		})
		return nil
	}

	p.logInfo("url_discovery_links_extracted", map[string]any{
		"source_id":   s.ID,
		"links_found": len(links),
		"filter_used": filter != nil,
	})

	if len(links) == 0 {
		p.logInfo("url_discovery_no_links_found", map[string]any{
			"source_id": s.ID,
			"url":       *s.SourceURL,
			"hint":      "No internal links found on this page",
		})
		return nil
	}

	added := 0
	skipped := 0
	limit := plan.Config.Scraping.MaxPagesPerCrawl

	for _, link := range links {
		if added >= limit {
			p.logInfo("url_discovery_limit_reached", map[string]any{
				"source_id": s.ID,
				"added":     added,
				"limit":     limit,
			})
			break
		}

		switch discoveryMode {
		case DiscoveryModeAuto:
			// Create discovered source (will not crawl further due to is_discovered=true)
			exists, _ := db.SourceExists(ctx, p.DB, s.ChatbotID, link)
			if !exists {
				newID, cerr := db.CreateDiscoveredSource(ctx, p.DB, s.ChatbotID, link)
				if cerr == nil {
					added++
					newIDs = append(newIDs, newID)
				} else {
					p.logWarn("url_discovery_create_source_failed", map[string]any{
						"source_id": s.ID,
						"link":      link,
						"error":     cerr.Error(),
					})
				}
			} else {
				skipped++
			}
		case DiscoveryModePending:
			// New behavior: add to pending table for approval
			// Check if already exists as source or in pending
			exists, _ := db.SourceExists(ctx, p.DB, s.ChatbotID, link)
			if !exists {
				sourceID := s.ID
				if err := db.InsertPendingURL(ctx, p.DB, s.ChatbotID, &sourceID, link); err == nil {
					added++
				} else {
					p.logWarn("url_discovery_insert_pending_failed", map[string]any{
						"source_id": s.ID,
						"link":      link,
						"error":     err.Error(),
					})
				}
			} else {
				skipped++
			}
		}
	}

	p.logInfo("url_discovery_completed", map[string]any{
		"source_id":   s.ID,
		"mode":        discoveryMode,
		"links_found": len(links),
		"added":       added,
		"skipped":     skipped,
		"limit":       limit,
	})
	return newIDs
}

// persistIngestionMetadata extracts and saves metadata for the source
func (p *URLProcessor) persistIngestionMetadata(ctx context.Context, content, langCode string, s *models.DataSource, maxQuestions int) {
	meta, err := rag.ExtractIngestionMetadata(ctx, p.OpenAIClient, content, langCode, maxQuestions)
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
