package processing

import (
	"context"
	"crypto/md5"
	"database/sql"

	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/pdf"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/scraper"
	"github.com/onurceri/botla-co/pkg/storage"
)

type SourceQueue struct {
	ch      *chan string
	db      *sql.DB
	storage storage.StorageService
}

func StartSourceQueue(dbpool *sql.DB, st storage.StorageService) (*SourceQueue, error) {
	c := make(chan string, 64)
	q := &SourceQueue{ch: &c, db: dbpool, storage: st}
	go q.worker()
	// Ensure collection exists at startup (best-effort)
	if qc, err := rag.NewQdrantClientFromEnv(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		qc.EnsureEmbeddingsCollection(ctx)
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
		// mark processing
		db.UpdateSourceProcessing(context.Background(), q.db, id, "processing", nil, 0, nil)
		// fetch source
		s, err := db.GetSourceByID(context.Background(), q.db, id)
		if err != nil || s == nil {
			msg := "source_not_found"
			db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &msg, 0, nil)
			continue
		}
		switch s.SourceType {
		case "url":
			if s.SourceURL == nil || *s.SourceURL == "" {
				msg := "empty_url"
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &msg, 0, nil)
				continue
			}
			content, err := scraper.ScrapeURLWithFallback(
				scraper.ScrapingTask{URL: *s.SourceURL},
				scraper.DefaultCollectorConfig(),
			)
			if err != nil {
				m := err.Error()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
				continue
			}
			if content == "" {
				now := time.Now()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, 0, &now)
				continue
			}
			chunks := ChunkText(content, 1500, 200)
			chunkCount := len(chunks)
			qc, err := rag.NewQdrantClientFromEnv()
			if err != nil {
				m := err.Error()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
				continue
			}
			oai, err := rag.NewOpenAIClientFromEnv()
			if err != nil {
				m := err.Error()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			for ci, ch := range chunks {
				emb, eerr := oai.CreateEmbedding(ctx, ch)
				if eerr != nil {
					m := eerr.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					cancel()
					continue
				}
				if err := qc.UpsertEmbedding(ctx, makePointID(id, ci), emb, rag.EmbeddingPayload{
					ChatbotID:    s.ChatbotID,
					SourceID:     s.ID,
					ChunkIndex:   ci,
					OriginalText: ch,
					SourceType:   s.SourceType,
					CreatedAt:    time.Now(),
				}); err != nil {
					m := err.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					cancel()
					continue
				}
			}
			cancel()
			now := time.Now()
			db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, chunkCount, &now)
		case "pdf":
			if s.FilePath == nil || *s.FilePath == "" {
				msg := "empty_file_path"
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &msg, 0, nil)
				continue
			}

			localPath := *s.FilePath
			// If storage is available, download it
			if q.storage != nil {
				rc, err := q.storage.DownloadFile(context.Background(), *s.FilePath)
				if err != nil {
					m := err.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					continue
				}
				// Save to temp file for PDF processing
				tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
				if err != nil {
					rc.Close()
					m := err.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					continue
				}
				_, err = io.Copy(tmpFile, rc)
				rc.Close()
				tmpFile.Close()
				if err != nil {
					m := err.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					continue
				}
				localPath = tmpFile.Name()
				defer os.Remove(localPath)
			}

			content, err := pdf.ExtractPDFText(localPath)
			if err != nil {
				m := err.Error()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
				continue
			}
			if content == "" {
				now := time.Now()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, 0, &now)
				continue
			}
			chunks := ChunkText(content, 1500, 200)
			chunkCount := len(chunks)
			qc, err := rag.NewQdrantClientFromEnv()
			if err != nil {
				m := err.Error()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
				continue
			}
			oai, err := rag.NewOpenAIClientFromEnv()
			if err != nil {
				m := err.Error()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			for ci, ch := range chunks {
				emb, eerr := oai.CreateEmbedding(ctx, ch)
				if eerr != nil {
					m := eerr.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					cancel()
					continue
				}
				if err := qc.UpsertEmbedding(ctx, makePointID(id, ci), emb, rag.EmbeddingPayload{
					ChatbotID:    s.ChatbotID,
					SourceID:     s.ID,
					ChunkIndex:   ci,
					OriginalText: ch,
					SourceType:   s.SourceType,
					CreatedAt:    time.Now(),
				}); err != nil {
					m := err.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					cancel()
					continue
				}
			}
			cancel()
			now := time.Now()
			db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, chunkCount, &now)
		case "text":
			if s.FilePath == nil || *s.FilePath == "" {
				msg := "empty_file_path"
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &msg, 0, nil)
				continue
			}
			var content string
			if q.storage != nil {
				rc, err := q.storage.DownloadFile(context.Background(), *s.FilePath)
				if err != nil {
					m := err.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					continue
				}
				b, err := io.ReadAll(rc)
				rc.Close()
				if err != nil {
					m := err.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					continue
				}
				content = string(b)
			} else {
				b, rerr := os.ReadFile(*s.FilePath)
				if rerr != nil {
					m := rerr.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					continue
				}
				content = string(b)
			}
			if content == "" {
				now := time.Now()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, 0, &now)
				continue
			}
			chunks := ChunkText(content, 1500, 200)
			chunkCount := len(chunks)
			qc, err := rag.NewQdrantClientFromEnv()
			if err != nil {
				m := err.Error()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
				continue
			}
			oai, err := rag.NewOpenAIClientFromEnv()
			if err != nil {
				m := err.Error()
				db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			for ci, ch := range chunks {
				emb, eerr := oai.CreateEmbedding(ctx, ch)
				if eerr != nil {
					m := eerr.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					cancel()
					continue
				}
				if err := qc.UpsertEmbedding(ctx, makePointID(id, ci), emb, rag.EmbeddingPayload{
					ChatbotID:    s.ChatbotID,
					SourceID:     s.ID,
					ChunkIndex:   ci,
					OriginalText: ch,
					SourceType:   s.SourceType,
					CreatedAt:    time.Now(),
				}); err != nil {
					m := err.Error()
					db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &m, 0, nil)
					cancel()
					continue
				}
			}
			cancel()
			now := time.Now()
			db.UpdateSourceProcessing(context.Background(), q.db, id, "completed", nil, chunkCount, &now)
		default:
			msg := "unsupported_type"
			db.UpdateSourceProcessing(context.Background(), q.db, id, "failed", &msg, 0, nil)
		}
	}
}

func DeleteSourceVectors(ctx context.Context, sourceID string) error {
	qc, err := rag.NewQdrantClientFromEnv()
	if err != nil {
		return err
	}
	return qc.DeleteBySourceID(ctx, sourceID)
}

func makePointID(sourceID string, index int) string {
	s := sourceID + ":" + strconv.Itoa(index)
	h := md5.Sum([]byte(s))
	// version 3 (MD5)
	h[6] = (h[6] & 0x0f) | 0x30
	h[8] = (h[8] & 0x3f) | 0x80
	u := h[:]
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}
