## Kapsam
- Chat endpoint’i ve RAG boru hattını ekleyerek sohbeti çalışır hale getirme.
- OpenAI embedding ve completion entegrasyonu.
- Qdrant arama ile kaynak bağlamı oluşturma.
- Conversation/message kayıtları ve token kullanımı takibi.

## Yeni/Düzenlenecek Dosyalar
- `internal/models/message.go`: Message modeli.
- `internal/models/conversation.go`: Conversation modeli.
- `internal/db/conversation.go`: Conversation ve Message DB işlemleri.
- `internal/api/handlers/chat.go`: `Chat(w, r)` handler.
- `internal/rag/openai.go`: `CreateEmbedding` ve `CreateCompletion`.
- `cmd/server/main.go`: Route kaydı (`POST /api/v1/chatbots/:chatbot_id/chat`).
- Gerekirse: `internal/rag/prompt.go` (prompt biçimleme yardımcıları).

## Mevcut Yapıya Uyum
- Router örüntüsü `cmd/server/main.go:36-61` ile aynı ServeMux + path-parsing mantığı.
- Auth koruması `pkg/middleware/auth.go:15-37` ile aynı.
- Qdrant istemcisi mevcut: `internal/rag/qdrant.go:140` (`SearchSimilar`).

## Model ve Şema
- Message:
  - `ID string, ConversationID string, Role string, Content string, TokensUsed int, ThumbsUp *bool, CreatedAt time.Time`.
- Conversation:
  - `ID string, ChatbotID string, SessionID *string, MessageCount int, CreatedAt time.Time, UpdatedAt time.Time`.
- Şema hazır: `db/migrations/0001_init.up.sql:62-89` (conversations, messages).

## DB Katmanı
- `GetOrCreateConversationBySessionID(ctx, db, chatbotID, sessionID) (*Conversation, error)`.
- `CreateMessage(ctx, db, msg *Message) (string, error)`.
- `IncrementConversationMessageCount(ctx, db, conversationID) error`.
- `ListRecentMessages(ctx, db, conversationID, limit int) ([]Message, error)` (gerekirse kısa geçmişi prompt’a katmak için).

## OpenAI Entegrasyonu
- `CreateEmbedding(text string) ([]float32, error)`
  - HTTP POST: `https://api.openai.com/v1/embeddings`.
  - Model: `text-embedding-3-small`.
  - Dönüş: `[]float32` vektör.
  - Not: Qdrant koleksiyon vektör boyutu 384 yerine 1536 olmalı; `internal/rag/qdrant.go:66-68` güncellenecek ve yeni koleksiyon oluşturma/yeniden yaratma stratejisi belirlenecek.
- `CreateCompletion(systemPrompt, context, userMessage string, model string, temperature float32, maxTokens int) (text string, tokens int, error)`
  - HTTP POST: `https://api.openai.com/v1/chat/completions` (messages format).
  - `usage.total_tokens` ile token sayısı.
  - Exponential backoff (max 3-5 deneme).

## Qdrant Arama ve Bağlam
- `SearchSimilar(embedding, chatbotID, topK)` kullanılır (`internal/rag/qdrant.go:140`).
- Sonuçlardaki `payload.original_text` birleştirilerek `context` oluşturulur.
- `sources_used`: `chunk_index` ve `source_type` alanlarından liste üretilir.

## Chat Handler Akışı
- Endpoint: `POST /api/v1/chatbots/:chatbot_id/chat` (`internal/api/handlers/chat.go`).
- Body: `{ "message": string, "session_id": string }`.
- Adımlar:
  1) Auth → kullanıcı doğrula (`UserIDFromContext`).
  2) Chatbot sahipliği kontrolü (`internal/db/chatbot.go:GetChatbotByID`).
  3) `GetOrCreateConversationBySessionID` ile conversation al/oluştur.
  4) Kullanıcı mesajını DB’ye kaydet.
  5) Mesaj embedding’i oluştur.
  6) Qdrant’ta top_k=5 benzer arama (chatbot_id filtresi).
  7) Context’i oluştur (benzer parçaları birleştir + gerektiğinde son n mesajı).
  8) System + Context + User mesaj ile OpenAI completion.
  9) Asistan mesajını DB’ye kaydet (tokens_used ile).
  10) `sources_used` listesini doldur.
  11) JSON response döndür.

## Hata ve Dayanıklılık
- OpenAI rate limit / geçici hata: exponential backoff.
- Qdrant hatası: bağlam olmadan sadece system+user ile yanıt dene; yine hata ise fallback mesaj: "Şu an bir hata oluştu, lütfen tekrar deneyin.".
- Token limit aşımı: `max_tokens` ve model sınırlarına göre truncate; yanıtı güvenli şekilde kısalt.
- Timeout’lar: `context.WithTimeout` çağrıları; HTTP client’lar `Timeout` ile.

## Güvenlik ve Konfigürasyon
- API anahtarları env’den: `OPENAI_API_KEY`, `QDRANT_URL`, `QDRANT_API_KEY` (opsiyonel), `JWT_SECRET`.
- Yanıt JSON’u:
  - `{ "response": string, "tokens_used": number, "sources_used": [{"chunk_index": number, "source_type": string}] }`.

## Route Kaydı
- `cmd/server/main.go` içine ek handler:
  - `/api/v1/chatbots/` altında composite handler’a `.../chat` dalı eklenir (ör. `internal/api/handlers/source.go:50-57` örüntüsüyle).

## Doğrulama
- Unit test: DB katmanı (conversation/message oluşturma), OpenAI istemci mock’u, Qdrant arama mock’u.
- Manuel test: Lokal server ile örnek istek; auth header ile.
- Loglama: Hata ve ölçümler `pkg/logger` ile.

## Teslimat
- Endpoint çalışır, yanıt üretir, `messages` ve `conversations` tablolarına kayıt ekler.
- Toplam token sayısı ve kullanılan kaynaklar dönülür.

Onay verirseniz uygulamaya başlayayım.