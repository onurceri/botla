# Plan 1.6: Auto-Refresh Scheduler

## Özet

URL kaynaklarının otomatik olarak belirli periyotlarda yeniden taranması ve güncellenmesi için zamanlayıcı sistemi.

---

## Mevcut Durum

### Mevcut Altyapı

| Dosya | Mevcut Özellik |
|-------|----------------|
| `internal/processing/url_processor.go` | Hash kontrolü ile değişiklik tespiti ✅ |
| `internal/models/chatbot.go` | refresh_count alanı mevcut |
| `db/migrations/000007_add_refresh_tracking.up.sql` | Temel refresh tracking ✅ |

**Eksik:** Otomatik tetikleme mekanizması (cron/scheduler)

---

## Hedef Mimari

```
┌────────────────────────────────────────────────────────────┐
│                    Scheduler Service                        │
├────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │ Daily Job   │    │ Weekly Job  │    │ Monthly Job │     │
│  │   (00:00)   │    │   (Sun)     │    │   (1st)     │     │
│  └──────┬──────┘    └──────┬──────┘    └──────┬──────┘     │
│         │                  │                   │            │
│         └──────────────────┼───────────────────┘            │
│                            ▼                                │
│              ┌─────────────────────────┐                    │
│              │   RefreshWorker         │                    │
│              │   - Find eligible bots  │                    │
│              │   - Queue refresh jobs  │                    │
│              │   - Update last_refresh │                    │
│              └─────────────────────────┘                    │
│                            │                                │
│                            ▼                                │
│              ┌─────────────────────────┐                    │
│              │   Existing SourceQueue  │                    │
│              │   (processing pipeline) │                    │
│              └─────────────────────────┘                    │
│                                                             │
└────────────────────────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration - Refresh Config

**Dosya:** `db/migrations/000012_auto_refresh_config.up.sql`

```sql
-- Chatbot'a refresh policy ekle
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS refresh_policy TEXT DEFAULT 'manual',
ADD COLUMN IF NOT EXISTS refresh_frequency TEXT DEFAULT NULL,
ADD COLUMN IF NOT EXISTS next_refresh_at TIMESTAMP DEFAULT NULL,
ADD COLUMN IF NOT EXISTS last_refresh_at TIMESTAMP DEFAULT NULL;

-- refresh_policy: 'manual', 'auto'
-- refresh_frequency: 'daily', 'weekly', 'monthly'

COMMENT ON COLUMN chatbots.refresh_policy IS 'manual or auto';
COMMENT ON COLUMN chatbots.refresh_frequency IS 'daily, weekly, or monthly (only for auto)';

-- Index for scheduler queries
CREATE INDEX idx_chatbots_next_refresh ON chatbots(next_refresh_at) 
WHERE refresh_policy = 'auto' AND deleted_at IS NULL;
```

### Adım 2: Model Güncelleme

**Dosya:** `internal/models/chatbot.go`

**Eklenecek alanlar:**
```go
RefreshPolicy    string     `json:"refresh_policy"`    // manual, auto
RefreshFrequency *string    `json:"refresh_frequency"` // daily, weekly, monthly
NextRefreshAt    *time.Time `json:"next_refresh_at"`
LastRefreshAt    *time.Time `json:"last_refresh_at"`
```

### Adım 3: Refresh Scheduler Service

**Dosya:** `internal/services/refresh_scheduler.go` (YENİ)

```go
type RefreshScheduler struct {
    DB        *sql.DB
    Queue     *processing.SourceQueue
    Log       *logger.Logger
    interval  time.Duration
    stopChan  chan struct{}
}

func NewRefreshScheduler(db *sql.DB, queue *processing.SourceQueue, log *logger.Logger) *RefreshScheduler

// Start scheduler loop
func (s *RefreshScheduler) Start(ctx context.Context)

// Stop scheduler
func (s *RefreshScheduler) Stop()

// FindDueForRefresh finds chatbots that need refresh
func (s *RefreshScheduler) FindDueForRefresh(ctx context.Context) ([]*models.Chatbot, error)

// QueueRefreshForChatbot enqueues all URL sources for refresh
func (s *RefreshScheduler) QueueRefreshForChatbot(ctx context.Context, botID string) error

// CalculateNextRefresh calculates next refresh time based on frequency
func CalculateNextRefresh(frequency string, from time.Time) time.Time
```

### Adım 4: Scheduler Görevleri

**Mantık:**

```go
func (s *RefreshScheduler) Start(ctx context.Context) {
    ticker := time.NewTicker(s.interval) // Her 5 dakikada kontrol
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-s.stopChan:
            return
        case <-ticker.C:
            bots, _ := s.FindDueForRefresh(ctx)
            for _, bot := range bots {
                if err := s.QueueRefreshForChatbot(ctx, bot.ID); err != nil {
                    s.Log.Warn("refresh_queue_error", map[string]any{"bot_id": bot.ID, "error": err})
                    continue
                }
                // Update next_refresh_at
                next := CalculateNextRefresh(bot.RefreshFrequency, time.Now())
                db.UpdateChatbotNextRefresh(ctx, s.DB, bot.ID, next, time.Now())
            }
        }
    }
}
```

### Adım 5: Server Başlatma Entegrasyonu

**Dosya:** `cmd/server/main.go`

```go
// Scheduler başlat
scheduler := services.NewRefreshScheduler(db, sourceQueue, log)
go scheduler.Start(ctx)

// Graceful shutdown
defer scheduler.Stop()
```

### Adım 6: API Endpoint'leri

**Dosya:** `internal/api/handlers/chatbot.go`

**Güncellemeler:**

```
PATCH /api/chatbots/:id
Body:
{
    "refresh_policy": "auto",
    "refresh_frequency": "weekly"
}

GET /api/chatbots/:id
Response (ek):
{
    "refresh_policy": "auto",
    "refresh_frequency": "weekly",
    "next_refresh_at": "2024-12-10T00:00:00Z",
    "last_refresh_at": "2024-12-03T02:15:00Z"
}
```

### Adım 7: Frontend - Refresh Ayarları

**Dosya:** `frontend/src/features/chatbot/RefreshSettings.tsx` (YENİ)

**UI Tasarımı:**

```
┌─────────────────────────────────────────────────────────────┐
│ 🔄 Otomatik Yenileme Ayarları                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Yenileme Politikası:                                        │
│ ○ Manuel - Sadece elle tetikleme                            │
│ ● Otomatik - Belirli aralıklarla yenile                     │
│                                                             │
│ ─────────────────────────────────────────────────────────── │
│                                                             │
│ Yenileme Sıklığı:                  (Otomatik seçiliyse)     │
│ ┌────────────────────────────────────────────────────────┐  │
│ │ ⏰ Günlük    │  📅 Haftalık  │  📆 Aylık            │  │
│ └────────────────────────────────────────────────────────┘  │
│                                                             │
│ ─────────────────────────────────────────────────────────── │
│                                                             │
│ ℹ️ Son Yenileme: 3 Aralık 2024, 02:15                        │
│ ⏱️ Sonraki Yenileme: 10 Aralık 2024, 00:00                   │
│                                                             │
│ Planınız: Pro (Aylık 5 otomatik yenileme hakkı)             │
│ Kullanılan: 2/5                                             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Adım 8: Plan Limitleri Kontrolü

**Dosya:** `internal/services/refresh_scheduler.go`

```go
func (s *RefreshScheduler) QueueRefreshForChatbot(ctx context.Context, botID string) error {
    // Plan kontrolü
    user, _ := db.GetUserByBotID(ctx, s.DB, botID)
    plan, _ := plans.GetPlan(user.PlanID)
    usage, _ := db.GetMonthlyRefreshUsage(ctx, s.DB, user.ID, time.Now())
    
    if usage >= plan.Config.Scraping.MonthlyAutoRefreshLimit {
        s.Log.Info("refresh_limit_reached", map[string]any{"user_id": user.ID, "limit": plan.Config.Scraping.MonthlyAutoRefreshLimit})
        return nil // Sessizce atla
    }
    
    // Refresh işlemine devam et...
}
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000012_*.sql` | YENİ | Migration |
| `internal/models/chatbot.go` | GÜNCELLE | Refresh alanları |
| `internal/services/refresh_scheduler.go` | YENİ | Scheduler |
| `internal/services/refresh_scheduler_test.go` | YENİ | Testler |
| `cmd/server/main.go` | GÜNCELLE | Scheduler başlatma |
| `internal/api/handlers/chatbot.go` | GÜNCELLE | API |
| `frontend/src/features/chatbot/RefreshSettings.tsx` | YENİ | UI |

---

## Test Planı

### Unit Testler

**Dosya:** `internal/services/refresh_scheduler_test.go` (YENİ)

```go
func TestCalculateNextRefresh_Daily(t *testing.T) {
    from := time.Date(2024, 12, 5, 10, 30, 0, 0, time.UTC)
    next := CalculateNextRefresh("daily", from)
    expected := time.Date(2024, 12, 6, 0, 0, 0, 0, time.UTC)
    assert.Equal(t, expected, next)
}

func TestCalculateNextRefresh_Weekly(t *testing.T) {
    // Test weekly calculation (next Sunday)
}

func TestCalculateNextRefresh_Monthly(t *testing.T) {
    // Test monthly calculation (1st of next month)
}

func TestFindDueForRefresh(t *testing.T) {
    // Test DB query for due bots
}
```

**Komut:** `go test ./internal/services/... -v -run TestRefresh`

### Integration Test

**Dosya:** `internal/integration/refresh_scheduler_test.go`

```go
func TestRefreshScheduler_FullCycle(t *testing.T) {
    // 1. Chatbot oluştur, auto refresh ayarla
    // 2. next_refresh_at'i geçmişe ayarla
    // 3. Scheduler'ı tetikle
    // 4. Queue'ya job eklendiğini doğrula
    // 5. next_refresh_at güncellendiğini doğrula
}
```

### Manuel Test Prosedürü

1. **Scheduler Test:**
   - Chatbot → Ayarlar → Otomatik Yenileme: "Günlük" yap
   - DB'de `next_refresh_at`'i 1 dakika öncesine ayarla
   - Server loglarında refresh tetiklendiğini gör
   - Sources'ların yeniden işlendiğini doğrula

2. **Plan Limit Test:**
   - Kullanıcının refresh limitini doldur
   - Yeni refresh tetiklendiğinde atlandığını doğrula

---

## Doğrulama Kriterleri

| Kriter | Doğrulama Yöntemi |
|--------|-------------------|
| ✅ Migration başarılı | `make migrate-up` |
| ✅ Next refresh hesaplama | Unit test |
| ✅ Scheduler döngüsü | Integration test |
| ✅ Plan limitleri | Unit test |
| ✅ API çalışıyor | API test |
| ✅ Frontend UI | Manuel test |
| ✅ %90 coverage | `make cover-gate` |

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration | 30 dk |
| Scheduler service | 4-6 saat |
| Server entegrasyonu | 1-2 saat |
| API updates | 2-3 saat |
| Frontend UI | 3-4 saat |
| Plan limit kontrolü | 2-3 saat |
| Testler | 3-4 saat |
| **TOPLAM** | **~1 hafta** |

---

## Bağımlılıklar

**Önceki:** Plan 1.5 (URL Checkbox UI) - queue altyapısı

**Sonraki:** Bağımsız
