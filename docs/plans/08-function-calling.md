# Plan 2.1: OpenAI Function Calling ve Actions Çerçevesi

## Özet

Chatbot'un dış sistemlerle etkileşime girmesini sağlayan Function Calling ve Custom Actions altyapısı.

---

## Mevcut Durum

| Dosya | Mevcut Durum |
|-------|--------------|
| `internal/rag/openai.go` | Basit completion, function call **yok** |
| `internal/services/chat_service.go` | Tek geçişli (single-turn) flow |

---

## Hedef Mimari

```
┌────────────────────────────────────────────────────────────────┐
│                      Chat Flow with Tools                       │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User Message                                                   │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              LLM with Tool Definitions                   │   │
│  │  - list_sources: Bilgi kaynaklarını listele              │   │
│  │  - get_weather: Hava durumu (custom action)              │   │
│  │  - create_ticket: Destek talebi oluştur                  │   │
│  └─────────────────────────────────────────────────────────┘   │
│       │                                                         │
│       ▼                                                         │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Response Types:                                         │   │
│  │  1. Direct answer → Return to user                       │   │
│  │  2. Tool call → Execute tool → Feed result → Continue    │   │
│  └─────────────────────────────────────────────────────────┘   │
│       │                                                         │
│       ▼                                                         │
│  Tool Execution                                                 │
│       │                                                         │
│       ▼                                                         │
│  Final Response                                                 │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

---

## Uygulama Adımları

### Adım 1: Tool/Action Model ve Migration

**Dosya:** `db/migrations/000014_chatbot_actions.up.sql`

```sql
CREATE TABLE IF NOT EXISTS chatbot_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chatbot_id UUID NOT NULL REFERENCES chatbots(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    action_type TEXT NOT NULL, -- 'builtin', 'http', 'zapier'
    config JSONB NOT NULL DEFAULT '{}',
    parameters JSONB NOT NULL DEFAULT '{}', -- JSON Schema
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(chatbot_id, name)
);

CREATE INDEX idx_chatbot_actions_bot ON chatbot_actions(chatbot_id) WHERE enabled = true;
```

### Adım 2: Models

**Dosya:** `internal/models/action.go` (YENİ)

```go
package models

type ActionType string

const (
    ActionTypeBuiltin ActionType = "builtin"
    ActionTypeHTTP    ActionType = "http"
    ActionTypeZapier  ActionType = "zapier"
)

type ChatbotAction struct {
    ID          string          `json:"id"`
    ChatbotID   string          `json:"chatbot_id"`
    Name        string          `json:"name"`
    Description string          `json:"description"`
    ActionType  ActionType      `json:"action_type"`
    Config      json.RawMessage `json:"config"`
    Parameters  json.RawMessage `json:"parameters"` // JSON Schema
    Enabled     bool            `json:"enabled"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
}

// HTTP Action config
type HTTPActionConfig struct {
    URL         string            `json:"url"`
    Method      string            `json:"method"`
    Headers     map[string]string `json:"headers"`
    AuthType    string            `json:"auth_type"` // none, bearer, api_key
    AuthConfig  json.RawMessage   `json:"auth_config"`
}

// Zapier Action config
type ZapierActionConfig struct {
    WebhookURL string `json:"webhook_url"`
}
```

### Adım 3: Tool Definitions for OpenAI

**Dosya:** `internal/rag/tools.go` (YENİ)

```go
package rag

// OpenAI Function Calling tool format
type Tool struct {
    Type     string       `json:"type"` // "function"
    Function ToolFunction `json:"function"`
}

type ToolFunction struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    Parameters  json.RawMessage `json:"parameters"` // JSON Schema
}

type ToolCall struct {
    ID       string `json:"id"`
    Type     string `json:"type"` // "function"
    Function struct {
        Name      string `json:"name"`
        Arguments string `json:"arguments"` // JSON string
    } `json:"function"`
}

// ConvertActionsToTools converts ChatbotActions to OpenAI tool format
func ConvertActionsToTools(actions []*models.ChatbotAction) []Tool

// Built-in tools
func GetBuiltinTools() []Tool {
    return []Tool{
        {
            Type: "function",
            Function: ToolFunction{
                Name:        "list_sources",
                Description: "Lists the available knowledge sources and their capabilities",
                Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
            },
        },
    }
}
```

### Adım 4: OpenAI Client Function Calling Desteği

**Dosya:** `internal/rag/openai.go`

**Değişiklikler:**

```go
type ChatRequestWithTools struct {
    Model       string        `json:"model"`
    Messages    []ChatMessage `json:"messages"`
    Tools       []Tool        `json:"tools,omitempty"`
    ToolChoice  string        `json:"tool_choice,omitempty"` // "auto", "none"
    Temperature float32       `json:"temperature,omitempty"`
    MaxTokens   int           `json:"max_tokens,omitempty"`
}

type ChatResponseWithTools struct {
    Choices []struct {
        Message struct {
            Role       string     `json:"role"`
            Content    *string    `json:"content"`
            ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
        } `json:"message"`
        FinishReason string `json:"finish_reason"`
    } `json:"choices"`
    Usage struct {
        TotalTokens int `json:"total_tokens"`
    } `json:"usage"`
}

// CreateCompletionWithTools sends a completion request with tool support
func (c *OpenAIClient) CreateCompletionWithTools(
    ctx context.Context,
    messages []ChatMessage,
    tools []Tool,
    model string,
    temperature float32,
    maxTokens int,
) (*ChatResponseWithTools, error)
```

### Adım 5: Tool Executor

**Dosya:** `internal/rag/tool_executor.go` (YENİ)

```go
package rag

type ToolExecutor struct {
    DB  *sql.DB
    Log *logger.Logger
}

type ToolResult struct {
    ToolCallID string `json:"tool_call_id"`
    Result     string `json:"result"` // JSON string
    Error      string `json:"error,omitempty"`
}

// Execute executes a tool call and returns the result
func (e *ToolExecutor) Execute(ctx context.Context, toolCall ToolCall, action *models.ChatbotAction) (*ToolResult, error) {
    switch action.ActionType {
    case models.ActionTypeBuiltin:
        return e.executeBuiltin(ctx, toolCall)
    case models.ActionTypeHTTP:
        return e.executeHTTP(ctx, toolCall, action)
    case models.ActionTypeZapier:
        return e.executeZapier(ctx, toolCall, action)
    default:
        return nil, fmt.Errorf("unknown action type: %s", action.ActionType)
    }
}

func (e *ToolExecutor) executeBuiltin(ctx context.Context, toolCall ToolCall) (*ToolResult, error) {
    switch toolCall.Function.Name {
    case "list_sources":
        // Return capability summaries
    default:
        return nil, fmt.Errorf("unknown builtin tool: %s", toolCall.Function.Name)
    }
}

func (e *ToolExecutor) executeHTTP(ctx context.Context, toolCall ToolCall, action *models.ChatbotAction) (*ToolResult, error) {
    // Parse config
    // Build HTTP request
    // Execute with timeout
    // Return result
}
```

### Adım 6: ChatService Agentic Loop

**Dosya:** `internal/services/chat_service.go`

**Yeni Akış:**

```go
func (s *ChatService) ProcessChatWithTools(ctx context.Context, req ChatRequest, bot *models.Chatbot, ragConfig models.RAGConfig) (*ChatResult, error) {
    // Get enabled actions for this chatbot
    actions, _ := db.GetEnabledActions(ctx, s.DB, bot.ID)
    tools := rag.ConvertActionsToTools(actions)
    tools = append(tools, rag.GetBuiltinTools()...)
    
    // Initial context retrieval
    embedding, _ := s.OAI.CreateEmbedding(ctx, req.Message)
    contextText, _, _ := rag.SearchContext(embedding, bot.ID, ragConfig.TopK, ragConfig.MaxContextTokens)
    
    messages := []rag.ChatMessage{
        {Role: "system", Content: systemPrompt},
        {Role: "user", Content: "Context:\n" + contextText + "\n\nQuestion:\n" + req.Message},
    }
    
    // Agentic loop (max 5 iterations to prevent infinite loops)
    executor := &rag.ToolExecutor{DB: s.DB, Log: s.Log}
    
    for i := 0; i < 5; i++ {
        response, err := s.OAI.CreateCompletionWithTools(ctx, messages, tools, bot.Model, bot.Temperature, bot.MaxTokens)
        if err != nil {
            return nil, err
        }
        
        choice := response.Choices[0]
        
        // If no tool calls, we have final answer
        if len(choice.Message.ToolCalls) == 0 {
            return &ChatResult{
                Response:   *choice.Message.Content,
                TokensUsed: response.Usage.TotalTokens,
            }, nil
        }
        
        // Execute tool calls
        messages = append(messages, rag.ChatMessage{Role: "assistant", Content: "", ToolCalls: choice.Message.ToolCalls})
        
        for _, tc := range choice.Message.ToolCalls {
            action := findActionByName(actions, tc.Function.Name)
            result, _ := executor.Execute(ctx, tc, action)
            messages = append(messages, rag.ChatMessage{
                Role:       "tool",
                ToolCallID: tc.ID,
                Content:    result.Result,
            })
        }
    }
    
    // Max iterations reached
    return &ChatResult{Response: "İşlem tamamlanamadı"}, nil
}
```

### Adım 7: API Endpoints - Action CRUD

**Dosya:** `internal/api/handlers/action.go` (YENİ)

```
GET    /api/chatbots/:id/actions
POST   /api/chatbots/:id/actions
GET    /api/chatbots/:id/actions/:actionId
PUT    /api/chatbots/:id/actions/:actionId
DELETE /api/chatbots/:id/actions/:actionId
POST   /api/chatbots/:id/actions/:actionId/test
```

### Adım 8: Frontend - Actions UI

**Dosya:** `frontend/src/features/actions/ActionsManager.tsx` (YENİ)

**UI Tasarımı:**

```
┌─────────────────────────────────────────────────────────────┐
│ ⚡ Actions (2 aktif)                           [+ Yeni Ekle]│
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 📋 list_sources (Yerleşik)              [✓] Aktif       │ │
│ │ Bilgi kaynaklarını ve yeteneklerini listeler            │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ 🌐 get_order_status (HTTP)             [✓] Aktif [⚙️]    │ │
│ │ Sipariş durumunu sorgular                                │ │
│ │ POST https://api.example.com/orders/status               │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ⚡ send_to_zapier (Zapier)              [✗] Devre dışı   │ │
│ │ Lead bilgilerini Zapier'a gönderir                       │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `db/migrations/000014_*.sql` | YENİ | Actions tablosu |
| `internal/models/action.go` | YENİ | Action modeli |
| `internal/rag/tools.go` | YENİ | Tool definitions |
| `internal/rag/openai.go` | GÜNCELLE | Function calling |
| `internal/rag/tool_executor.go` | YENİ | Tool execution |
| `internal/services/chat_service.go` | GÜNCELLE | Agentic loop |
| `internal/api/handlers/action.go` | YENİ | CRUD API |
| `frontend/src/features/actions/*` | YENİ | UI |

---

## Test Planı

### Unit Testler

```go
func TestConvertActionsToTools(t *testing.T) {
    // Action'ların doğru tool formatına dönüştürüldüğünü test et
}

func TestToolExecutor_Builtin(t *testing.T) {
    // list_sources gibi builtin tool'ların çalıştığını test et
}

func TestToolExecutor_HTTP(t *testing.T) {
    // Mock HTTP server ile HTTP action testi
}

func TestProcessChatWithTools_NoTools(t *testing.T) {
    // Tool yokken normal akışın çalıştığını test et
}

func TestProcessChatWithTools_WithToolCall(t *testing.T) {
    // Tool call yapıldığında execute edilip sonucun kullanıldığını test et
}
```

### Manuel Test

1. Chatbot oluştur
2. HTTP action ekle (test URL ile)
3. Chat gönder, action'ın tetiklendiğini loglardan gör
4. Sonucun yanıtta kullanıldığını doğrula

---

## Doğrulama Kriterleri

| Kriter | Doğrulama Yöntemi |
|--------|-------------------|
| ✅ Actions tablosu | Migration |
| ✅ Tool format doğru | Unit test |
| ✅ Function calling çalışıyor | Integration test |
| ✅ Agentic loop | Unit test |
| ✅ HTTP executor | Mock test |
| ✅ API endpoints | API test |
| ✅ %90 coverage | `make cover-gate` |

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Migration + Models | 1-2 saat |
| Tool definitions | 2-3 saat |
| OpenAI function calling | 4-6 saat |
| Tool executor | 4-6 saat |
| Agentic loop | 4-6 saat |
| API endpoints | 3-4 saat |
| Frontend UI | 6-8 saat |
| Testler | 4-6 saat |
| **TOPLAM** | **~1.5 hafta** |

---

## Bağımlılıklar

**Önceki:** Plan 1.1 (LLM Client Abstraction) - interface kullanımı

**Sonraki:** Plan 2.2 (Zapier Integration) bu altyapıyı kullanır
