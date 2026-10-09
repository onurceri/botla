package harness

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/onurceri/botla-app/internal/evals/cache"
	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/text"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
)

// ProductChunkTokens is the chunk size used by the product's text and URL
// processors (internal/processing). Keep in sync if the product changes it.
const ProductChunkTokens = 512

// ProductSourceType is the source type recorded for text sources.
const ProductSourceType = "text"

const capabilityNamespace = "capabilities"

// Index records what was ingested, so that the chunks the product placed in
// the LLM context can be mapped back to corpus documents and their text.
type Index struct {
	ChatbotID    string
	chunks       map[string]string // "<sourceID>#<chunkIndex>" -> chunk text
	ChunksPerDoc map[string]int
	// Capabilities are the per-source capability summaries the product
	// generates at ingestion time and injects into the system prompt.
	Capabilities []string
}

func chunkKey(sourceID string, idx int) string { return sourceID + "#" + strconv.Itoa(idx) }

// NewIndex returns an empty index.
func NewIndex(chatbotID string) *Index {
	return &Index{ChatbotID: chatbotID, chunks: map[string]string{}, ChunksPerDoc: map[string]int{}}
}

// AddChunk records a chunk.
func (ix *Index) AddChunk(sourceID string, idx int, txt string) {
	ix.chunks[chunkKey(sourceID, idx)] = txt
	if idx+1 > ix.ChunksPerDoc[sourceID] {
		ix.ChunksPerDoc[sourceID] = idx + 1
	}
}

// ChunkText returns the text of a chunk.
func (ix *Index) ChunkText(sourceID string, idx int) (string, bool) {
	t, ok := ix.chunks[chunkKey(sourceID, idx)]
	return t, ok
}

// TotalChunks returns the number of indexed chunks.
func (ix *Index) TotalChunks() int { return len(ix.chunks) }

// Ingestor loads the corpus through the product's ingestion code path.
type Ingestor struct {
	Embedder rag.EmbeddingClient
	Vector   rag.VectorClient
	// LLM is used for capability-summary extraction (rag.ExtractIngestionMetadata).
	// When nil, capability summaries are skipped and the system prompt has none.
	LLM      rag.LLMClient
	Cache    *cache.Cache
	Log      *logger.Logger
	Language string
}

// Ingest chunks, embeds and upserts every corpus document, mirroring
// processing.TextProcessor.ProcessWithSteps: text.NormalizeTR ->
// rag.ChunkText(512 tokens) -> rag.EmbeddingService.GenerateForSource.
// Each document's previous points are deleted first so that re-runs against a
// persistent Qdrant collection never see stale chunks.
func (in *Ingestor) Ingest(ctx context.Context, chatbotID string, corpus *dataset.Corpus) (*Index, error) {
	if err := in.Vector.EnsureEmbeddingsCollection(ctx); err != nil {
		return nil, fmt.Errorf("ensure vector collection: %w", err)
	}
	lang := in.Language
	if lang == "" {
		lang = "en"
	}
	svc := rag.NewEmbeddingService(in.Embedder, in.Vector, in.Log)
	ix := NewIndex(chatbotID)
	for _, doc := range corpus.Docs {
		if err := in.Vector.DeleteBySourceID(ctx, doc.ID); err != nil {
			return nil, fmt.Errorf("clear previous points for %s: %w", doc.ID, err)
		}
		content := text.NormalizeTR(doc.Text)
		chunks, err := rag.ChunkText(nil, content, ProductChunkTokens, lang)
		if err != nil {
			return nil, fmt.Errorf("chunk %s: %w", doc.ID, err)
		}
		if err := svc.GenerateForSource(ctx, chunks, chatbotID, doc.ID, ProductSourceType); err != nil {
			return nil, fmt.Errorf("embed %s: %w", doc.ID, err)
		}
		for i, ch := range chunks {
			ix.AddChunk(doc.ID, i, ch.Text)
		}
		if summary := in.capability(ctx, content, lang); summary != "" {
			ix.Capabilities = append(ix.Capabilities, summary)
		}
	}
	sort.Strings(ix.Capabilities) // order-independent system prompt across runs
	return ix, nil
}

// capability runs the product's metadata extraction (cached by content).
// Failures are non-fatal, exactly as in the product's ingestion path.
func (in *Ingestor) capability(ctx context.Context, content, lang string) string {
	if in.LLM == nil {
		return ""
	}
	key := cache.Key(config.DefaultModelName, lang, content)
	var cached string
	if ok, err := in.Cache.Get(capabilityNamespace, key, &cached); err == nil && ok && cached != "" {
		return cached
	}
	meta, err := rag.ExtractIngestionMetadata(ctx, in.LLM, content, lang, 3)
	if err != nil {
		if in.Log != nil {
			in.Log.Warn("eval_capability_extraction_failed", map[string]any{"error": err.Error()})
		}
		return ""
	}
	_ = in.Cache.Put(capabilityNamespace, key, meta.CapabilitySummary)
	return meta.CapabilitySummary
}
