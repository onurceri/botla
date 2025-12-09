# 04. Temperature & Model Configuration Tests

> **Priority**: High  
> **Test Count**: 15  
> **Source Files**: `internal/services/chat_service.go`, `internal/rag/client_factory.go`

---

## 4.1 Temperature Parameter

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| TMP-001 | Temperature 0.0 (deterministic) | Consistent output for same input | ✅ |
| TMP-002 | Temperature 1.0 (creative) | Valid but varied output | ✅ |
| TMP-003 | Temperature 2.0 (max) | Valid output, no errors | ✅ |
| TMP-004 | Temperature passed to OpenAI | Correct value in API request | ✅ |
| TMP-005 | Temperature passed to Anthropic | Correct value in API request | ✅ |
| TMP-006 | Temperature passed to Google AI | Correct value in API request | ✅ |
| TMP-007 | Default temperature (0.7) | Applied when not specified | ✅ |

### Technical Notes

```go
// internal/services/chat_service.go:130-137
params := models.CompletionParams{
    SystemPrompt: sp,
    Context:      contextText,
    UserMessage:  req.Message,
    Model:        modelName,
    Temperature:  bot.Temperature,  // ← Test this
    MaxTokens:    bot.MaxTokens,
}
```

---

## 4.2 Model Configuration

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| MDL-001 | Model `gpt-4o-mini` (default) | Uses OpenAI client | ✅ |
| MDL-002 | Model `gpt-4o` | Uses OpenAI client | ✅ |
| MDL-003 | Model prefix `anthropic:claude-3-sonnet` | Uses Anthropic client | ✅ |
| MDL-004 | Model prefix `google:gemini-pro` | Uses Google AI client | ✅ |
| MDL-005 | Model prefix `openrouter:meta-llama/llama-3` | Uses OpenRouter client | ✅ |
| MDL-006 | Invalid model name | Fallback to gpt-4o-mini | ✅ |
| MDL-007 | Model not in plan's `allowed_models` | Fallback to allowed model | ✅ |

### Technical Notes

```go
// internal/rag/client_factory.go
// Model prefixes:
// - "anthropic:" → AnthropicClient
// - "google:" → GoogleAIClient
// - "openrouter:" → OpenRouterClient
// - default → OpenAIClient
```

---

## 4.3 MaxTokens Configuration

| ID | Test Case | Expected Result | Status |
|----|-----------|-----------------|--------|
| MTK-001 | MaxTokens 256 (low) | Response truncated appropriately | ✅ |
| MTK-002 | MaxTokens 4096 (high) | Full response returned | ✅ |
| MTK-003 | MaxTokens 0 (default) | Default 512 | ✅ |
| MTK-004 | MaxTokens passed to LLM | Correct value in API request | ✅ |

---

## Existing Test Coverage

| File | Coverage |
|------|----------|
| `internal/rag/openai_test.go` | OpenAI client |
| `internal/rag/openrouter_test.go` | OpenRouter client |
| `internal/integration/temperature_model_test.go` | Integration tests for Temperature, Models, and MaxTokens |
