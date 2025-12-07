# Plan 3.3: Gelişmiş Analytics

## Özet

Kaynak başına performans, müşteri memnuniyeti, yanıtlanamayan sorular ve detaylı kullanım metrikleri.

---

## Mevcut Durum

| Dosya | Mevcut Özellik |
|-------|----------------|
| `internal/db/analytics.go` | Temel analytics (mesaj sayısı, token) ✅ |
| `internal/api/handlers/analytics.go` | Temel endpoint'ler ✅ |
| Frontend | Basit dashboard ✅ |

---

## Hedef Metrikler

```
┌────────────────────────────────────────────────────────────────┐
│                      Advanced Analytics                         │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  📊 Kaynak Performansı                                          │
│  ────────────────────                                          │
│  - Hangi kaynak ne kadar kullanılıyor?                         │
│  - Hangi kaynaktan düşük skor geliyor?                         │
│  - Kaynak başına yanıt kalitesi                                │
│                                                                 │
│  😊 Kullanıcı Memnuniyeti                                       │
│  ────────────────────────                                      │
│  - Thumbs up/down oranı                                        │
│  - Konuşma uzunluğu (engagement)                               │
│  - Handoff oranı (memnuniyetsizlik göstergesi)                 │
│                                                                 │
│  ❓ Yanıtlanamayan Sorular                                      │
│  ─────────────────────────                                     │
│  - Düşük confidence cevaplar                                   │
│  - "Bilmiyorum" dönen sorular                                  │
│  - En sık sorulan ama cevaplanmayan konular                    │
│                                                                 │
│  📈 Trend Analizi                                               │
│  ───────────────                                               │
│  - Günlük/haftalık/aylık karşılaştırma                         │
│  - Peak saatler                                                │
│  - Mevsimsel desenler                                          │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration

**Dosya:** `db/migrations/000019_advanced_analytics.up.sql`

```sql
-- Message-level analytics
ALTER TABLE messages
ADD COLUMN IF NOT EXISTS confidence_score FLOAT,
ADD COLUMN IF NOT EXISTS sources_used UUID[] DEFAULT '{}',
ADD COLUMN IF NOT EXISTS feedback_rating INT, -- -1, 0, 1
ADD COLUMN IF NOT EXISTS feedback_text TEXT;

-- Source usage tracking
CREATE TABLE IF NOT EXISTS source_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    hit_count INT DEFAULT 0,
    avg_score FLOAT DEFAULT 0,
    
    UNIQUE(chatbot_id, source_id, date)
);

-- Unanswered queries tracking
CREATE TABLE IF NOT EXISTS unanswered_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    query_embedding VECTOR(1536),
    occurrence_count INT DEFAULT 1,
    last_occurred_at TIMESTAMP DEFAULT NOW(),
    addressed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Aggregated daily stats
CREATE TABLE IF NOT EXISTS analytics_daily (
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    total_messages INT DEFAULT 0,
    unique_sessions INT DEFAULT 0,
    avg_confidence FLOAT,
    positive_feedback INT DEFAULT 0,
    negative_feedback INT DEFAULT 0,
    handoff_count INT DEFAULT 0,
    avg_response_time_ms INT,
    unanswered_count INT DEFAULT 0,
    
    PRIMARY KEY (chatbot_id, date)
);

CREATE INDEX idx_source_usage_date ON source_usage(chatbot_id, date);
CREATE INDEX idx_unanswered_chatbot ON unanswered_queries(chatbot_id, addressed);
```

### Adım 2: Models

**Dosya:** `internal/models/analytics.go` (YENİ/GÜNCELLE)

```go
type MessageAnalytics struct {
    ConfidenceScore float64  `json:"confidence_score"`
    SourcesUsed     []string `json:"sources_used"`
    FeedbackRating  *int     `json:"feedback_rating"` // -1, 0, 1
    FeedbackText    *string  `json:"feedback_text"`
}

type SourceUsage struct {
    SourceID  string  `json:"source_id"`
    SourceURL *string `json:"source_url"`
    HitCount  int     `json:"hit_count"`
    AvgScore  float64 `json:"avg_score"`
}

type UnansweredQuery struct {
    ID              string    `json:"id"`
    Query           string    `json:"query"`
    OccurrenceCount int       `json:"occurrence_count"`
    LastOccurredAt  time.Time `json:"last_occurred_at"`
    Addressed       bool      `json:"addressed"`
}

type DailyAnalytics struct {
    Date              string  `json:"date"`
    TotalMessages     int     `json:"total_messages"`
    UniqueSessions    int     `json:"unique_sessions"`
    AvgConfidence     float64 `json:"avg_confidence"`
    PositiveFeedback  int     `json:"positive_feedback"`
    NegativeFeedback  int     `json:"negative_feedback"`
    HandoffCount      int     `json:"handoff_count"`
    AvgResponseTimeMs int     `json:"avg_response_time_ms"`
    UnansweredCount   int     `json:"unanswered_count"`
}
```

### Adım 3: Analytics Service Güncelleme

**Dosya:** `internal/services/analytics_service.go` (YENİ)

```go
type AnalyticsService struct {
    DB  *sql.DB
    Log *logger.Logger
}

// TrackMessageAnalytics tracks message-level analytics
func (s *AnalyticsService) TrackMessageAnalytics(ctx context.Context, msgID string, confidence float64, sourcesUsed []string)

// RecordFeedback records user feedback for a message
func (s *AnalyticsService) RecordFeedback(ctx context.Context, msgID string, rating int, text *string)

// TrackSourceUsage updates source usage statistics
func (s *AnalyticsService) TrackSourceUsage(ctx context.Context, chatbotID string, sources []SourceHit)

// TrackUnansweredQuery tracks or updates an unanswered query
func (s *AnalyticsService) TrackUnansweredQuery(ctx context.Context, chatbotID, query string, embedding []float32)

// GetSourcePerformance returns performance metrics per source
func (s *AnalyticsService) GetSourcePerformance(ctx context.Context, chatbotID string, days int) ([]SourceUsage, error)

// GetUnansweredQueries returns unanswered queries grouped by similarity
func (s *AnalyticsService) GetUnansweredQueries(ctx context.Context, chatbotID string, limit int) ([]UnansweredQuery, error)

// GetDailyTrends returns daily analytics for a date range
func (s *AnalyticsService) GetDailyTrends(ctx context.Context, chatbotID string, from, to time.Time) ([]DailyAnalytics, error)
```

### Adım 4: ChatService Entegrasyonu

**Dosya:** `internal/services/chat_service.go`

```go
func (s *ChatService) ProcessChat(...) (*ChatResult, error) {
    startTime := time.Now()
    
    // ... mevcut akış
    
    // Track analytics
    go func() {
        ctx := context.Background()
        
        // Message analytics
        s.Analytics.TrackMessageAnalytics(ctx, msgID, maxScore, usedSourceIDs)
        
        // Source usage
        s.Analytics.TrackSourceUsage(ctx, bot.ID, sourceHits)
        
        // Unanswered tracking
        if maxScore < bot.ConfidenceThreshold {
            s.Analytics.TrackUnansweredQuery(ctx, bot.ID, req.Message, embedding)
        }
        
        // Response time
        responseTimeMs := int(time.Since(startTime).Milliseconds())
        s.Analytics.TrackResponseTime(ctx, bot.ID, responseTimeMs)
    }()
    
    return &ChatResult{...}, nil
}
```

### Adım 5: Feedback API

**Dosya:** `internal/api/handlers/public.go`

```go
// POST /api/public/:botId/feedback
func (h *PublicHandlers) SubmitFeedback(c *gin.Context) {
    var req struct {
        MessageID string  `json:"message_id" binding:"required"`
        Rating    int     `json:"rating" binding:"required,min=-1,max=1"`
        Text      *string `json:"text"`
    }
    
    // Validate and save
    h.Analytics.RecordFeedback(c.Request.Context(), req.MessageID, req.Rating, req.Text)
}
```

### Adım 6: Analytics API Endpoints

```
GET /api/chatbots/:id/analytics/overview
GET /api/chatbots/:id/analytics/sources
GET /api/chatbots/:id/analytics/unanswered
GET /api/chatbots/:id/analytics/trends?from=2024-01-01&to=2024-01-31
GET /api/chatbots/:id/analytics/feedback
POST /api/chatbots/:id/analytics/unanswered/:queryId/address
```

### Adım 7: Frontend Dashboard

**UI:**

```
┌─────────────────────────────────────────────────────────────────┐
│ 📊 Analytics Dashboard                     📅 Son 30 gün   ▼    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐    │
│  │   1,234   │  │   87%     │  │   92%     │  │    12     │    │
│  │  Mesaj    │  │ Güven Sk. │  │ Olumlu FB │  │ Handoff   │    │
│  │  ↑ 12%    │  │  ↓ 2%     │  │  = 0%     │  │  ↑ 3      │    │
│  └───────────┘  └───────────┘  └───────────┘  └───────────┘    │
│                                                                 │
│  ─────────────────────────────────────────────────────────────  │
│                                                                 │
│  📈 Mesaj Trendi                                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │     *                                                    │   │
│  │   *   *     *                                            │   │
│  │  *     *   * *                                    *      │   │
│  │ *       * *   *                                  * *     │   │
│  │*         *     *          *                     *   *    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  📚 Kaynak Performansı                      🔍 Cevaplanamayan   │
│  ┌───────────────────────────┐  ┌───────────────────────────┐  │
│  │ Kaynak        Kull.  Skor │  │ "iade politikası nedir?" │  │
│  │ /faq          234   0.89  │  │ 23 kez soruldu           │  │
│  │ /blog/*       156   0.82  │  │ [📝 İçerik Ekle]         │  │
│  │ /products     89    0.75  │  │                          │  │
│  └───────────────────────────┘  │ "kargo ücreti ne kadar?" │  │
│                                  │ 18 kez soruldu           │  │
│                                  │ [📝 İçerik Ekle]         │  │
│                                  └───────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Adım 8: Widget Feedback UI

**Dosya:** `widget/src/components/MessageFeedback.tsx` (YENİ)

```tsx
const MessageFeedback = ({ messageId, onFeedback }) => {
    return (
        <div className="message-feedback">
            <button onClick={() => onFeedback(messageId, 1)}>👍</button>
            <button onClick={() => onFeedback(messageId, -1)}>👎</button>
        </div>
    );
};
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000019_*.sql` | YENİ | Tablolar |
| `internal/models/analytics.go` | GÜNCELLE | Yeni modeller |
| `internal/services/analytics_service.go` | YENİ | Analytics service |
| `internal/services/chat_service.go` | GÜNCELLE | Tracking |
| `internal/api/handlers/analytics.go` | GÜNCELLE | API endpoints |
| `internal/api/handlers/public.go` | GÜNCELLE | Feedback API |
| `widget/src/components/MessageFeedback.tsx` | YENİ | Feedback UI |
| `frontend/src/pages/Analytics.tsx` | GÜNCELLE | Dashboard |

---

## Test Planı

### Unit Testler

```go
func TestTrackMessageAnalytics(t *testing.T) {
    // Track and verify stored
}

func TestTrackUnansweredQuery_NewQuery(t *testing.T) {
    // New query → insert
}

func TestTrackUnansweredQuery_ExistingQuery(t *testing.T) {
    // Similar query → increment count
}

func TestGetSourcePerformance(t *testing.T) {
    // Aggregated results
}
```

### Manuel Test

1. Widget'ta mesaj gönder
2. Thumbs up/down tıkla
3. Analytics dashboard'da görüntüle
4. Kaynak performansını kontrol et
5. Cevaplanamayan sorular listesine bak

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration | 2-3 saat |
| Models + Service | 4-6 saat |
| ChatService integration | 2-3 saat |
| API endpoints | 3-4 saat |
| Widget feedback | 2-3 saat |
| Frontend dashboard | 6-8 saat |
| Testler | 3-4 saat |
| **TOPLAM** | **~1 hafta** |

---

## Bağımlılıklar

**Önceki:** Plan 3.1 (Multi-Tenant) - Org bazlı aggregation için

**Sonraki:** Bağımsız
