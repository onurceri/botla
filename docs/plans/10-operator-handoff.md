# Plan 2.3: Operatör Handoff (İnsan Temsilciye Devir)

## Özet

Chatbot'un cevaplayamadığı veya kullanıcının insan yardımı istediği durumlarda destek ekibine aktarım.

---

## Hedef Mimari

```
┌────────────────────────────────────────────────────────────────┐
│                      Handoff Options                            │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Option 1: Email Handoff (Basit)                               │
│  ─────────────────────────────────────                         │
│  - Konuşma dökümü e-posta olarak gönderilir                    │
│  - Düşük karmaşıklık, hızlı implementasyon                     │
│                                                                 │
│  Option 2: Crisp Integration                                    │
│  ─────────────────────────────────────                         │
│  - Mevcut Crisp panelinde konuşma görünür                      │
│  - Temsilci devralabilir                                       │
│                                                                 │
│  Option 3: Native Panel (Gelecek)                              │
│  ─────────────────────────────────────                         │
│  - Botla içinde canlı destek paneli                            │
│  - En karmaşık, en iyi UX                                      │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

**İlk fazda Option 1 (Email) ve Option 2 (Crisp) odaklı plan.**

---

## Uygulama Adımları

### Adım 1: Veritabanı Migration

**Dosya:** `db/migrations/000015_handoff_config.up.sql`

```sql
ALTER TABLE chatbots
ADD COLUMN IF NOT EXISTS handoff_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS handoff_type TEXT DEFAULT 'email', -- email, crisp, native
ADD COLUMN IF NOT EXISTS handoff_config JSONB DEFAULT '{}';

CREATE TABLE IF NOT EXISTS handoff_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    status TEXT DEFAULT 'pending', -- pending, assigned, resolved
    assigned_to TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    resolved_at TIMESTAMP
);
```

### Adım 2: Models

**Dosya:** `internal/models/handoff.go` (YENİ)

```go
type HandoffType string

const (
    HandoffTypeEmail  HandoffType = "email"
    HandoffTypeCrisp  HandoffType = "crisp"
    HandoffTypeNative HandoffType = "native"
)

type HandoffConfig struct {
    // Email
    EmailTo      string `json:"email_to,omitempty"`
    EmailSubject string `json:"email_subject,omitempty"`
    
    // Crisp
    CrispWebsiteID string `json:"crisp_website_id,omitempty"`
    CrispAPIKey    string `json:"crisp_api_key,omitempty"`
}

type HandoffRequest struct {
    ID             string     `json:"id"`
    ChatbotID      string     `json:"chatbot_id"`
    ConversationID string     `json:"conversation_id"`
    Status         string     `json:"status"`
    AssignedTo     *string    `json:"assigned_to,omitempty"`
    CreatedAt      time.Time  `json:"created_at"`
    ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}
```

### Adım 3: Handoff Service

**Dosya:** `internal/services/handoff_service.go` (YENİ)

```go
type HandoffService struct {
    DB         *sql.DB
    EmailSvc   EmailService
    CrispClient *crisp.Client
    Log        *logger.Logger
}

// RequestHandoff creates a handoff request and notifies operators
func (s *HandoffService) RequestHandoff(ctx context.Context, bot *models.Chatbot, convID string) (*models.HandoffRequest, error) {
    // Create handoff request
    req, err := db.CreateHandoffRequest(ctx, s.DB, bot.ID, convID)
    if err != nil {
        return nil, err
    }
    
    // Load conversation history
    messages, _ := db.GetConversationMessages(ctx, s.DB, convID)
    
    // Notify based on handoff type
    switch bot.HandoffType {
    case models.HandoffTypeEmail:
        return s.handleEmailHandoff(ctx, bot, req, messages)
    case models.HandoffTypeCrisp:
        return s.handleCrispHandoff(ctx, bot, req, messages)
    default:
        return nil, fmt.Errorf("unsupported handoff type: %s", bot.HandoffType)
    }
}

func (s *HandoffService) handleEmailHandoff(ctx context.Context, bot *models.Chatbot, req *models.HandoffRequest, messages []*models.Message) (*models.HandoffRequest, error) {
    var config models.HandoffConfig
    json.Unmarshal(bot.HandoffConfig, &config)
    
    // Build email content
    body := buildHandoffEmailBody(messages, req.ID)
    
    // Send email
    err := s.EmailSvc.Send(ctx, config.EmailTo, config.EmailSubject, body)
    if err != nil {
        s.Log.Warn("handoff_email_failed", map[string]any{"error": err})
    }
    
    return req, nil
}
```

### Adım 4: Crisp Integration

**Dosya:** `internal/integrations/crisp/client.go` (YENİ)

```go
package crisp

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
)

type Client struct {
    websiteID string
    apiKey    string
    http      *http.Client
}

func NewClient(websiteID, apiKey string) *Client

// CreateConversation creates a new conversation in Crisp
func (c *Client) CreateConversation(ctx context.Context, sessionID string) (string, error)

// SendMessage sends a message to a Crisp conversation
func (c *Client) SendMessage(ctx context.Context, convID string, content string, from string) error

// TransferToOperator marks conversation for operator pickup
func (c *Client) TransferToOperator(ctx context.Context, convID string, notes string) error
```

### Adım 5: Chat Service'e Handoff Desteği

**Dosya:** `internal/services/chat_service.go`

```go
// ProcessChat içinde handoff detection
func (s *ChatService) ProcessChat(...) (*ChatResult, error) {
    // ... mevcut akış
    
    // Handoff trigger kontrolü
    if s.shouldTriggerHandoff(ans, bot) {
        if bot.HandoffEnabled {
            handoffSvc := NewHandoffService(s.DB, ...)
            _, err := handoffSvc.RequestHandoff(ctx, bot, conv.ID)
            if err != nil {
                s.Log.Warn("handoff_failed", ...)
            }
            
            // Add handoff acknowledgment to response
            ans = cfg.ResponseTemplates.HandoffMessage + "\n\n" + ans
        }
    }
    
    return &ChatResult{...}, nil
}

func (s *ChatService) shouldTriggerHandoff(response string, bot *models.Chatbot) bool {
    // LLM "bilmiyorum" tipi cevap verdiyse
    // veya kullanıcı açıkça "insan istiyorum" dediyse
    // veya belirli keywords içeriyorsa
    return containsHandoffTriggers(response)
}
```

### Adım 6: Widget'ta "İnsan Yardımı İste" Butonu

**Dosya:** `widget/src/components/ChatDrawer.tsx`

```tsx
const ChatDrawer = ({ handoffEnabled, onRequestHandoff }) => {
    return (
        <div className="chat-drawer">
            {/* ... mevcut chat */}
            
            {handoffEnabled && (
                <button 
                    className="handoff-button"
                    onClick={onRequestHandoff}
                >
                    👤 İnsan Desteği İste
                </button>
            )}
        </div>
    );
};
```

### Adım 7: API Endpoints

**Dosya:** `internal/api/handlers/handoff.go` (YENİ)

```
POST /api/public/:botId/handoff
Body: { "session_id": "...", "message": "Optional note" }
Response: { "request_id": "...", "status": "pending" }

GET /api/chatbots/:id/handoff-requests
Response: { "requests": [...] }

PATCH /api/chatbots/:id/handoff-requests/:requestId
Body: { "status": "resolved" }
```

### Adım 8: Frontend - Handoff Settings

**Dosya:** `frontend/src/features/chatbot/HandoffSettings.tsx` (YENİ)

**UI:**

```
┌─────────────────────────────────────────────────────────────┐
│ 👤 İnsan Desteği Ayarları                                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ☑️ İnsan desteği aktif                                       │
│                                                             │
│ Handoff Tipi:                                               │
│ ○ E-posta - Konuşma e-posta olarak gönderilir               │
│ ● Crisp - Crisp paneline aktarılır                          │
│                                                             │
│ ─────────────────────────────────────────────────────────── │
│                                                             │
│ [E-posta Ayarları]                                          │
│ E-posta Adresi: support@example.com                         │
│ E-posta Başlığı: Yeni Destek Talebi - {bot_name}            │
│                                                             │
│ [Crisp Ayarları]                                            │
│ Website ID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx            │
│ API Key: ********                                           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000015_*.sql` | YENİ | Migration |
| `internal/models/handoff.go` | YENİ | Models |
| `internal/services/handoff_service.go` | YENİ | Service |
| `internal/integrations/crisp/client.go` | YENİ | Crisp API |
| `internal/services/chat_service.go` | GÜNCELLE | Handoff trigger |
| `internal/api/handlers/handoff.go` | YENİ | API |
| `widget/src/components/ChatDrawer.tsx` | GÜNCELLE | Buton |
| `frontend/src/features/chatbot/HandoffSettings.tsx` | YENİ | UI |

---

## Test Planı

### Unit Testler

```go
func TestShouldTriggerHandoff(t *testing.T) {
    // "Bilmiyorum" içeren yanıt → true
    // Normal yanıt → false
    // "insan istiyorum" → true
}

func TestHandoffService_Email(t *testing.T) {
    // Mock email service
    // Verify email sent with correct content
}
```

### Manuel Test

1. Chatbot → Ayarlar → İnsan Desteği: Aktif et (E-posta)
2. Widget'ta "İnsan Desteği İste" butonuna bas
3. E-posta aldığını doğrula
4. E-postada konuşma dökümü olduğunu doğrula

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration + Models | 1-2 saat |
| Handoff service | 4-6 saat |
| Email handoff | 2-3 saat |
| Crisp integration | 4-6 saat |
| Widget button | 1-2 saat |
| API endpoints | 2-3 saat |
| Frontend UI | 4-6 saat |
| Testler | 3-4 saat |
| **TOPLAM** | **~1.5 hafta** |

---

## Bağımlılıklar

**Önceki:** Bağımsız

**Sonraki:** Plan 3.1 (Multi-Tenant) - Gelişmiş native panel için
