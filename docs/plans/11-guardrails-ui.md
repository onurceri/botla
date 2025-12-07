# Plan 2.4: Guardrails UI

## Özet

RAG güven skoru, halüsinasyon kontrolü ve fallback mesajları için kullanıcı arayüzü.

---

## Mevcut Durum

| Dosya | Mevcut Özellik |
|-------|----------------|
| `internal/rag/search.go` | `RAG_SCORE_THRESHOLD` mevcut |
| `pkg/langconfig/*.go` | `NoInfoFound` mesajı mevcut |
| Frontend | Bu ayarlar için UI **yok** |

---

## Hedef Mimari

```
┌────────────────────────────────────────────────────────────┐
│                    Guardrails Configuration                 │
├────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Güven Eşiği (Confidence Threshold)                     │
│     - Score < threshold → "Bilmiyorum" döndür              │
│                                                             │
│  2. Fallback Mesajları                                      │
│     - no_info_found: "Bu konuda bilgi bulamadım"           │
│     - error_message: "Bir hata oluştu"                     │
│     - handoff_message: "Sizi temsilciye bağlıyorum"        │
│                                                             │
│  3. Konu Kısıtlamaları (Topic Restrictions)                │
│     - Belirli konularda cevap verme/verme                   │
│                                                             │
└────────────────────────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration

**Dosya:** `db/migrations/000016_guardrails_config.up.sql`

```sql
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS confidence_threshold FLOAT DEFAULT 0.7,
ADD COLUMN IF NOT EXISTS fallback_messages JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS topic_restrictions JSONB DEFAULT '{}';

COMMENT ON COLUMN chatbots.confidence_threshold IS 'Minimum RAG score to provide answer (0.0-1.0)';
```

### Adım 2: Model Güncelleme

**Dosya:** `internal/models/chatbot.go`

```go
ConfidenceThreshold float64          `json:"confidence_threshold"`
FallbackMessages    FallbackMessages `json:"fallback_messages"`
TopicRestrictions   TopicConfig      `json:"topic_restrictions,omitempty"`

type FallbackMessages struct {
    NoInfoFound    string `json:"no_info_found"`
    ErrorMessage   string `json:"error_message"`
    HandoffMessage string `json:"handoff_message"`
}

type TopicConfig struct {
    AllowedTopics  []string `json:"allowed_topics,omitempty"`
    BlockedTopics  []string `json:"blocked_topics,omitempty"`
    BlockedMessage string   `json:"blocked_message,omitempty"`
}
```

### Adım 3: ChatService'te Guardrails Kontrolü

**Dosya:** `internal/services/chat_service.go`

```go
func (s *ChatService) ProcessChat(...) (*ChatResult, error) {
    // ... embedding ve search
    
    // Confidence check
    maxScore := getMaxScore(searchResults)
    if maxScore < bot.ConfidenceThreshold {
        fallback := bot.FallbackMessages.NoInfoFound
        if fallback == "" {
            fallback = cfg.ResponseTemplates.NoInfoFound
        }
        return &ChatResult{Response: fallback, TokensUsed: 0}, nil
    }
    
    // Topic restriction check (opsiyonel)
    if len(bot.TopicRestrictions.BlockedTopics) > 0 {
        if isBlockedTopic(req.Message, bot.TopicRestrictions.BlockedTopics) {
            return &ChatResult{
                Response: bot.TopicRestrictions.BlockedMessage,
            }, nil
        }
    }
    
    // Continue with normal flow...
}
```

### Adım 4: API Güncellemesi

**Dosya:** `internal/api/handlers/chatbot.go`

```go
// PATCH /api/chatbots/:id
// Body:
{
    "confidence_threshold": 0.75,
    "fallback_messages": {
        "no_info_found": "Bu konuda size yardımcı olamıyorum.",
        "error_message": "Teknik bir sorun oluştu, lütfen tekrar deneyin."
    }
}
```

### Adım 5: Frontend - Guardrails Settings

**Dosya:** `frontend/src/features/chatbot/GuardrailsSettings.tsx` (YENİ)

**UI:**

```
┌─────────────────────────────────────────────────────────────┐
│ 🛡️ Guardrails Ayarları                                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Güven Eşiği:                                                │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ●───────────────────────────○                     0.7   │ │
│ │ 0.0                                               1.0   │ │
│ └─────────────────────────────────────────────────────────┘ │
│ ℹ️ Düşük skor, daha fazla "bilmiyorum" yanıtı verir.        │
│    Yüksek skor, daha riskli yanıtlara izin verir.          │
│                                                             │
│ ─────────────────────────────────────────────────────────── │
│                                                             │
│ Fallback Mesajları:                                         │
│                                                             │
│ 📭 Bilgi Bulunamadı:                                        │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Bu konuda size yardımcı olamıyorum. Başka bir konuda   │ │
│ │ sormak ister misiniz?                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ ⚠️ Hata Mesajı:                                              │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Teknik bir sorun oluştu. Lütfen daha sonra tekrar      │ │
│ │ deneyin veya destek ekibimize ulaşın.                   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ 👤 Handoff Mesajı:                                           │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Sizi bir müşteri temsilcisine yönlendiriyorum...       │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│                                        [Varsayılana Dön]    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000016_*.sql` | YENİ | Migration |
| `internal/models/chatbot.go` | GÜNCELLE | Yeni alanlar |
| `internal/services/chat_service.go` | GÜNCELLE | Guardrails logic |
| `internal/api/handlers/chatbot.go` | GÜNCELLE | API |
| `frontend/src/features/chatbot/GuardrailsSettings.tsx` | YENİ | UI |

---

## Test Planı

### Unit Testler

```go
func TestProcessChat_BelowThreshold(t *testing.T) {
    // RAG score < threshold
    // Expect fallback message
}

func TestProcessChat_AboveThreshold(t *testing.T) {
    // RAG score > threshold
    // Expect normal response
}

func TestProcessChat_CustomFallback(t *testing.T) {
    // Custom fallback message configured
    // Expect custom message, not default
}
```

### Manuel Test

1. Chatbot → Ayarlar → Guardrails
2. Güven eşiğini 0.9 yap (çok yüksek)
3. Bilgi tabanında olmayan bir soru sor
4. Fallback mesajının göründüğünü doğrula
5. Eşiği 0.3 yap (çok düşük)
6. Aynı soruyu sor
7. Yanıt üretildiğini doğrula

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration + Model | 1-2 saat |
| ChatService update | 2-3 saat |
| API update | 1-2 saat |
| Frontend UI | 3-4 saat |
| Testler | 2-3 saat |
| **TOPLAM** | **~3-4 gün** |

---

## Bağımlılıklar

**Önceki:** Bağımsız

**Sonraki:** Bağımsız
