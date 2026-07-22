> Extracted from: rag-pipeline.md — Section: 2.7 LLM Generation

## 2.7 LLM Generation

### 2.7.1 Provider Resolution (`client_factory.go`)

- [ ] Provider is resolved based on chatbot config: OpenAI or OpenRouter
- [ ] Model name is resolved via `services/model_registry.go`
- [ ] Default model is used if no model is configured
- [ ] Provider API keys are read from environment variables

### 2.7.2 OpenAI Integration (`openai.go`)

- [ ] Chat completions API is used
- [ ] System prompt includes context from retrieved chunks
- [ ] Token limits are respected (max_tokens, context window)
- [ ] Streaming is supported (SSE to client)
- [ ] Error handling for API errors (auth, rate limit, server error)

### 2.7.3 OpenRouter Integration (`openrouter.go`)

- [ ] Same interface as OpenAI provider
- [ ] Model fallback if primary model is unavailable
- [ ] OpenRouter-specific headers are set (referer, title)
- [ ] Error handling for OpenRouter-specific error codes

### 2.7.4 Circuit Breaker (`circuit_breaker.go`)

- [ ] Uses `sony/gobreaker` library
- [ ] Configuration: 50% failure threshold, 30s timeout, 3 requests in half-open, 60s window
- [ ] Circuit states: closed (normal), open (failing), half-open (testing)
- [ ] When open: requests fail fast with "Service temporarily unavailable"
- [ ] When half-open: limited requests are allowed to test recovery
- [ ] Successful half-open request closes the circuit
- [ ] Failed half-open request re-opens the circuit
- [ ] Metrics are exposed for monitoring (circuit state, request count, failure count)
- [ ] Circuit breaker wraps all LLM API calls
