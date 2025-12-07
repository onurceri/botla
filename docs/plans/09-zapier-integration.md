# Plan 2.2: Zapier Entegrasyonu

## Özet

Chatbot'un Zapier webhook'larına veri göndermesi ve 6000+ uygulamayla entegrasyon kurması.

---

## Mevcut Durum

Plan 2.1 (Function Calling) tamamlandığını varsayarak, Zapier action tipi için spesifik implementasyon.

---

## Hedef Mimari

```
┌────────────────────────────────────────────────────────────┐
│                    Zapier Integration Flow                  │
├────────────────────────────────────────────────────────────┤
│                                                             │
│  Chat Conversation:                                         │
│  "E-posta adresim test@example.com"                        │
│       │                                                     │
│       ▼                                                     │
│  LLM detects: send_to_zapier(email: "test@example.com")    │
│       │                                                     │
│       ▼                                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Zapier Webhook                                      │   │
│  │  POST https://hooks.zapier.com/...                   │   │
│  │  Body: { "email": "test@example.com", ... }          │   │
│  └─────────────────────────────────────────────────────┘   │
│       │                                                     │
│       ▼                                                     │
│  Zapier → Mailchimp, Google Sheets, CRM, etc.              │
│                                                             │
└────────────────────────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Zapier Webhook URL Validasyonu

**Dosya:** `internal/rag/tool_executor.go`

```go
func (e *ToolExecutor) executeZapier(ctx context.Context, toolCall ToolCall, action *models.ChatbotAction) (*ToolResult, error) {
    var config models.ZapierActionConfig
    if err := json.Unmarshal(action.Config, &config); err != nil {
        return nil, err
    }
    
    // Validate Zapier URL
    if !isValidZapierWebhook(config.WebhookURL) {
        return nil, fmt.Errorf("invalid Zapier webhook URL")
    }
    
    // Parse arguments
    var args map[string]interface{}
    if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
        return nil, err
    }
    
    // Send to Zapier
    body, _ := json.Marshal(args)
    req, _ := http.NewRequestWithContext(ctx, "POST", config.WebhookURL, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := e.httpClient.Do(req)
    if err != nil {
        return &ToolResult{ToolCallID: toolCall.ID, Error: err.Error()}, nil
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return &ToolResult{ToolCallID: toolCall.ID, Error: fmt.Sprintf("Zapier returned %d", resp.StatusCode)}, nil
    }
    
    return &ToolResult{
        ToolCallID: toolCall.ID,
        Result:     `{"status": "success", "message": "Data sent to Zapier"}`,
    }, nil
}

func isValidZapierWebhook(url string) bool {
    return strings.HasPrefix(url, "https://hooks.zapier.com/")
}
```

### Adım 2: Zapier Action Template

**Dosya:** `internal/api/handlers/action.go`

```go
// Predefined Zapier action templates
var ZapierTemplates = map[string]struct {
    Name        string
    Description string
    Parameters  json.RawMessage
}{
    "collect_lead": {
        Name:        "collect_lead",
        Description: "Collect lead information and send to Zapier",
        Parameters: json.RawMessage(`{
            "type": "object",
            "properties": {
                "email": { "type": "string", "description": "Email address" },
                "name": { "type": "string", "description": "Full name" },
                "phone": { "type": "string", "description": "Phone number" },
                "message": { "type": "string", "description": "Additional message" }
            },
            "required": ["email"]
        }`),
    },
    "create_ticket": {
        Name:        "create_ticket",
        Description: "Create a support ticket",
        Parameters: json.RawMessage(`{
            "type": "object",
            "properties": {
                "subject": { "type": "string" },
                "description": { "type": "string" },
                "priority": { "type": "string", "enum": ["low", "medium", "high"] }
            },
            "required": ["subject", "description"]
        }`),
    },
}
```

### Adım 3: Frontend - Zapier Action UI

**Dosya:** `frontend/src/features/actions/ZapierActionForm.tsx` (YENİ)

**UI Tasarımı:**

```
┌─────────────────────────────────────────────────────────────┐
│ ⚡ Yeni Zapier Action                                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ Şablon Seç:                                                 │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 📧 Lead Topla          │ 🎫 Destek Talebi               │ │
│ │ 📅 Randevu Oluştur     │ 🔧 Özel Şablon                 │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Action Adı:                                                 │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ collect_visitor_email                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Zapier Webhook URL:                                         │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ https://hooks.zapier.com/hooks/catch/...               │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ ℹ️ Zapier'da bir Webhook trigger oluşturup URL'i buraya     │
│    yapıştırın. Detaylı rehber için tıklayın.               │
│                                                             │
│                              [🧪 Test Et]  [💾 Kaydet]       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Adım 4: Webhook Test Endpoint

**Dosya:** `internal/api/handlers/action.go`

```go
// POST /api/chatbots/:id/actions/:actionId/test
func (h *ActionHandlers) TestAction(c *gin.Context) {
    action, err := h.getAction(c)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Action not found"})
        return
    }
    
    var testData map[string]interface{}
    if err := c.ShouldBindJSON(&testData); err != nil {
        // Use sample data based on parameters
        testData = generateSampleData(action.Parameters)
    }
    
    executor := &rag.ToolExecutor{...}
    result, err := executor.Execute(c.Request.Context(), rag.ToolCall{
        Function: struct{ Name, Arguments string }{
            Name:      action.Name,
            Arguments: mustMarshal(testData),
        },
    }, action)
    
    c.JSON(http.StatusOK, gin.H{
        "success": err == nil && result.Error == "",
        "result":  result,
        "error":   err,
    })
}
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `internal/rag/tool_executor.go` | GÜNCELLE | Zapier executor |
| `internal/api/handlers/action.go` | GÜNCELLE | Templates, test endpoint |
| `frontend/src/features/actions/ZapierActionForm.tsx` | YENİ | UI |

---

## Test Planı

### Unit Testler

```go
func TestExecuteZapier_ValidWebhook(t *testing.T) {
    // Mock HTTP server
    // Execute Zapier action
    // Verify request body
}

func TestExecuteZapier_InvalidWebhook(t *testing.T) {
    // URL that doesn't start with hooks.zapier.com
    // Should return error
}
```

### Manuel Test

1. Zapier hesabında webhook trigger oluştur
2. Botla'da Zapier action ekle
3. Test Et butonuna bas
4. Zapier'da veri geldiğini doğrula
5. Chat'te action tetikleyecek mesaj yaz
6. Zapier'da yeni veri oluştuğunu doğrula

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Zapier executor | 2-3 saat |
| Templates | 1-2 saat |
| Frontend UI | 3-4 saat |
| Test endpoint | 1-2 saat |
| Testler | 2-3 saat |
| **TOPLAM** | **~1 hafta** |

---

## Bağımlılıklar

**Önceki:** Plan 2.1 (Function Calling) tamamlanmış olmalı

**Sonraki:** Bağımsız
