## Mevcut Durum ve Boşluklar
- Upload: `internal/api/handlers/source.go` PDF/URL/TEXT kaynak oluşturuyor, temp dizine kaydedip DB'ye `status=pending` ile ekliyor ve kuyruğa atıyor.
- Queue/Worker: `internal/processing/sources_queue.go` status güncelliyor, PDF’de `pdf.ExtractPDFText` çağırıyor (OCR fallback şimdi mevcut) ama embedding/vektör kaydı ve `chunk_count` güncellemesi yapılmıyor.
- Qdrant: `internal/rag/qdrant.go` EnsureEmbeddingsCollection, UpsertEmbedding, DeleteBySourceID ve arama (`SearchSimilar`) hazır.
- OpenAI embedding: `internal/rag/openai.go` `CreateEmbedding` hazır.

## Hedefler
1. Upload sınırını dokümana uygun hale getirmek (PDF < 50MB).
2. Worker içinde kaynak tipine göre içerik çıkarımı → normalize → chunk.
3. Her chunk için embedding üretip Qdrant’a `payload` ile upsert etmek.
4. `chunk_count` ve `processed_at` güncellemesiyle `status=completed` yapmak; hatada anlamlı `failed`.

## Teknik Tasarım
- Chunking yardımcıları (`internal/processing/chunker.go`):
  - `ChunkText(input string) []string`
  - Heuristik: ~1200–1600 karakter hedef, ~200 karakter overlap; paragraf/satır sonu gözeterek böl.
- Worker akışı (`internal/processing/sources_queue.go`):
  - `url`: `scraper.ScrapeURLWithFallback` → `NormalizeText` → `ChunkText` → embedding + Qdrant Upsert
  - `pdf`: `pdf.ExtractPDFText` (OCR fallback dahil) → `ChunkText` → embedding + Upsert
  - `text`: temp dosyadan oku → normalize → `ChunkText` → embedding + Upsert
  - Upsert ID: `sourceID:chunkIndex` (string)
  - Payload: `EmbeddingPayload{ChatbotID, SourceID, ChunkIndex, OriginalText, SourceType, CreatedAt}`
  - Başlangıçta `EnsureEmbeddingsCollection` çağrısı zaten var; embedding istemcisi `OpenAIClientFromEnv`
- Durum Güncellemeleri:
  - Başarı: `UpdateSourceProcessing(..., "completed", nil, chunkCount, &now)`
  - Hata: `UpdateSourceProcessing(..., "failed", &msg, 0, nil)` ve continue

## Hata ve Dayanıklılık
- Embedding veya Qdrant hatalarında loglamayı sürdürüp kaynağı `failed` yap.
- Boş içerik durumunda chunking atlanır, `completed` + `chunk_count=0` (gelecekte yeniden işleme opsiyonu eklenebilir).

## İyileştirmeler
- Upload PDF limitini 50MB’a yükseltme (`source.go:68-99`).
- `ChunkText` için birim testi: kısa/uzun metin, paragraf koruma, overlap doğrulama.

## Teslim Edilecek Değişiklikler
- Yeni: `internal/processing/chunker.go` (+ test)
- Güncelleme: `internal/processing/sources_queue.go` (chunking + embedding + upsert + status)
- Güncelleme: `internal/api/handlers/source.go` (PDF boyut limiti: 50MB)

Onaylarsanız bu değişiklikleri uygulayıp testleri çalıştıracağım.