# Plan 1.1: LLM Client Soyutlaması (Multi-Model Desteği)

## Özet

Mevcut OpenAI-only yapıyı genişleterek Claude, Gemini ve diğer LLM'leri destekleyen soyut bir client katmanı oluşturma.

---

## Mevcut Durum

### Dosya: `internal/rag/openai.go`

```go
type OpenAIClient struct {
    apiKey       string
    http         *http.Client
    base         string
    defaultModel string
}
```

**Sorun:** Doğrudan OpenAI'a bağlı. Başka model eklemek için yapısal değişiklik gerekli.

---

## Hedef Mimari

```
┌─────────────────────────────────────────────────────────────────┐
│                         LLMClient Interface                     │
├─────────────────────────────────────────────────────────────────┤
│ + CreateEmbedding(ctx, text) ([]float32, error)                 │
│ + CreateEmbeddingsBatch(ctx, texts) ([][]float32, error)        │
│ + CreateCompletion(ctx, params) (CompletionResult, error)       │
│ + GetSupportedModels() []ModelInfo                              │
└─────────────────────────────────────────────────────────────────┘
                              ▲
            ┌─────────────────┼─────────────────┐
            │                 │                 │
     ┌──────┴──────┐   ┌──────┴──────┐   ┌──────┴──────┐
     │ OpenAI      │   │ Anthropic   │   │ Google      │
     │ Client      │   │ Client      │   │ AI Client   │
     └─────────────┘   └─────────────┘   └─────────────┘
```

---

## Uygulama Adımları

### Adım 1: Interface Tanımlama

**Dosya:** `internal/rag/llm_client.go` (YENİ)

```
Oluşturulacak yapılar:
- LLMClient interface
- CompletionParams struct
- CompletionResult struct
- ModelInfo struct
- EmbeddingClient interface (ayrı, çünkü her provider desteklemeyebilir)
```

**İçerik:**
- `CreateCompletion(ctx, CompletionParams) (*CompletionResult, error)`
- `GetModelInfo() ModelInfo`
- Model bilgisi: name, provider, maxTokens, supportedFeatures

### Adım 2: OpenAI Client'ı Refactor Et

**Dosya:** `internal/rag/openai.go`

**Değişiklikler:**
1. `OpenAIClient` struct'ına interface implement ettir
2. `NewOpenAIClient(config OpenAIConfig)` constructor ekle
3. Mevcut `NewOpenAIClientFromEnv()` fonksiyonunu koru (geriye uyumluluk)
4. `CreateCompletion` signature'ını yeni params ile güncelle

### Adım 3: Anthropic (Claude) Client Ekle

**Dosya:** `internal/rag/anthropic.go` (YENİ)

**Gereksinimler:**
- Anthropic API formatı OpenAI'dan farklı (messages API)
- `claude-3-5-sonnet-20241022` varsayılan model
- Rate limiting ve retry mekanizması

**Environment Variables:**
- `ANTHROPIC_API_KEY`
- `ANTHROPIC_API_BASE` (opsiyonel, varsayılan: https://api.anthropic.com)

### Adım 4: Google AI (Gemini) Client Ekle

**Dosya:** `internal/rag/googleai.go` (YENİ)

**Gereksinimler:**
- Google AI API formatı
- `gemini-1.5-flash` varsayılan model
- Farklı authentication (API key vs Service Account)

**Environment Variables:**
- `GOOGLE_AI_API_KEY`
- `GOOGLE_AI_PROJECT_ID` (opsiyonel)

### Adım 5: Client Factory Oluştur

**Dosya:** `internal/rag/client_factory.go` (YENİ)

```
Fonksiyonlar:
- NewLLMClient(provider string, config map[string]string) (LLMClient, error)
- GetAvailableProviders() []string
- IsProviderConfigured(provider string) bool
```

**Provider mapping:**
- `openai` → OpenAIClient
- `anthropic` → AnthropicClient
- `google` → GoogleAIClient

### Adım 6: Model Yapılandırmasını Güncelle

**Dosya:** `pkg/config/config.go`

**Eklenecekler:**
```
- AllowedModels: map[string][]string (plan bazlı)
- DefaultModels: map[string]string (provider → default model)
- ModelMappings: map[string]ModelConfig
```

### Adım 7: Chatbot Modelini Güncelle

**Dosya:** `internal/models/chatbot.go`

**Değişiklik:**
```
Model alanını provider:model formatına çevir
Örnek: "openai:gpt-4o-mini", "anthropic:claude-3-5-sonnet"
```

### Adım 8: ChatService'i Güncelle

**Dosya:** `internal/services/chat_service.go`

**Değişiklikler:**
1. `ProcessChat` fonksiyonunda model parsing
2. Provider'a göre client seçimi
3. Fallback mekanizması (provider hata verirse)

### Adım 9: Veritabanı Migration

**Dosya:** `db/migrations/000009_model_provider.up.sql`

```sql
-- Model formatını güncelle
UPDATE chatbots 
SET model = 'openai:' || model 
WHERE model NOT LIKE '%:%';

-- Plan config'e allowed_models ekle
ALTER TABLE plans 
ADD COLUMN IF NOT EXISTS allowed_models JSONB DEFAULT '["openai:gpt-4o-mini"]';
```

### Adım 10: Frontend Güncellemesi

**Dosya:** `frontend/src/features/chatbot/ChatbotSettings.tsx`

**Değişiklikler:**
1. Model seçim dropdown'ını provider gruplarına ayır
2. Plan bazlı model filtreleme
3. Provider ikonu gösterimi

---

## Dosya Değişiklikleri Özeti

| Dosya | İşlem | Açıklama |
|-------|-------|----------|
| `internal/rag/llm_client.go` | YENİ | Interface tanımları |
| `internal/rag/openai.go` | GÜNCELLE | Interface implement |
| `internal/rag/anthropic.go` | YENİ | Claude client |
| `internal/rag/googleai.go` | YENİ | Gemini client |
| `internal/rag/client_factory.go` | YENİ | Factory pattern |
| `pkg/config/config.go` | GÜNCELLE | Model yapılandırması |
| `internal/models/chatbot.go` | GÜNCELLE | Model format |
| `internal/services/chat_service.go` | GÜNCELLE | Client seçimi |
| `db/migrations/000009_*.sql` | YENİ | Migration |
| `frontend/src/features/chatbot/*` | GÜNCELLE | UI değişiklikleri |

---

## Test Planı

### Unit Testler

**Dosya:** `internal/rag/llm_client_test.go` (YENİ)

1. **Interface Compliance Test:**
   - Her client'ın interface'i doğru implement ettiğini kontrol et

2. **OpenAI Client Tests:**
   - Mevcut testleri koru
   - Yeni signature ile uyumluluğu test et

3. **Mock Client Test:**
   - Test için mock LLMClient oluştur
   - Integration testlerde kullan

**Komut:** `go test ./internal/rag/... -v`

### Integration Testler

**Dosya:** `internal/integration/llm_integration_test.go`

1. **Multi-Provider Test:**
   - Aynı prompt'u farklı provider'lara gönder
   - Response formatlarını karşılaştır

2. **Fallback Test:**
   - Provider hata verdiğinde fallback çalışıyor mu?

**Komut:** `go test ./internal/integration/... -v -tags=integration`

### Manuel Test

1. Dashboard'dan chatbot oluştur
2. Model olarak "anthropic:claude-3-5-sonnet" seç
3. Chat gönder, yanıt al
4. Logs'ta doğru provider'ın kullanıldığını doğrula

---

## Doğrulama Kriterleri

| Kriter | Doğrulama Yöntemi |
|--------|-------------------|
| ✅ Interface tanımlı | `go build` başarılı |
| ✅ OpenAI çalışıyor | Mevcut testler geçiyor |
| ✅ Claude çalışıyor | Integration test |
| ✅ Gemini çalışıyor | Integration test |
| ✅ Plan bazlı kısıtlama | Unit test + manuel |
| ✅ %90 coverage | `make cover-gate` |
| ✅ Lint temiz | `make lint` |

---

## Riskler ve Mitigasyon

| Risk | Olasılık | Etki | Mitigasyon |
|------|----------|------|------------|
| API format uyumsuzluğu | Orta | Yüksek | Response normalizer |
| Rate limiting farklılıkları | Düşük | Orta | Provider-specific backoff |
| Maliyet artışı | Orta | Orta | Usage tracking per provider |

---

## Tahmini Süre

| Görev | Süre |
|-------|------|
| Interface tasarımı | 2-3 saat |
| OpenAI refactor | 3-4 saat |
| Anthropic client | 4-6 saat |
| Google AI client | 4-6 saat |
| Factory + Config | 2-3 saat |
| Migration + DB | 1-2 saat |
| Frontend | 4-6 saat |
| Testler | 4-6 saat |
| **TOPLAM** | **~1 hafta** |

---

## Sonraki Adımlar

Bu plan tamamlandığında:
- ✅ [2.1 Function Calling](./08-function-calling.md) için temel hazır
- ✅ Mevcut tüm chatbot'lar geriye uyumlu çalışır
