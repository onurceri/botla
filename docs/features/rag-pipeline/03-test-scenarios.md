> Extracted from: rag-pipeline.md — Section: 3. Test Scenarios

# 3. Test Scenarios

## 3.1 RAG Pipeline Scenarios

### 3.1.1 Happy Path — Question with Strong Match

1. Ingest content about "product pricing" into a chatbot
2. Ask: "What are your prices?"
3. **Expected:** Query is embedded
4. **Expected:** Search returns results with score ≥ 0.50 (high tier)
5. **Expected:** LLM generates answer from retrieved context
6. **Expected:** Response includes citations/sources from matching content
7. **Expected:** Response latency is within acceptable range (< 5s)

### 3.1.2 Weak Match (Medium Tier)

1. Ask a question loosely related to ingested content
2. Search returns results with score 0.30–0.49
3. **Expected:** Response includes a confidence warning
4. **Expected:** LLM qualifies the answer ("Based on available information...")
5. **Expected:** Citations are still included

### 3.1.3 No Match (Low Tier) — Static Fallback

1. Ask a question completely unrelated to ingested content
2. Search returns all results with score < 0.30
3. **Expected:** Static fallback message is returned
4. **Expected:** Response does not attempt to answer the question
5. **Expected:** Message is user-friendly, not a technical error

### 3.1.4 No Match (Low Tier) — Smart Fallback

1. Chatbot is configured with smart fallback mode
2. Ask unrelated question
3. **Expected:** AI generates a redirect response
4. **Expected:** Response suggests topics the chatbot CAN help with

### 3.1.5 No Match (Low Tier) — Escalate Fallback

1. Chatbot is configured with escalate fallback mode
2. Ask unrelated question
3. **Expected:** Response offers human handoff
4. **Expected:** Handoff flow is triggered if user accepts

### 3.1.6 Circuit Breaker — Closed (Normal Operation)

1. All LLM API calls succeed
2. **Expected:** Circuit breaker remains closed
3. **Expected:** All requests pass through normally
4. **Expected:** Metrics show 0 failures, closed state

### 3.1.7 Circuit Breaker — Open (Failure Threshold Exceeded)

1. LLM API starts returning 5xx errors (or timing out)
2. Failure rate exceeds 50% in the 60s window
3. **Expected:** Circuit breaker transitions to open state
4. **Expected:** Subsequent requests fail immediately with "Service temporarily unavailable"
5. **Expected:** LLM is not called (fast fail)

### 3.1.8 Circuit Breaker — Half-Open (Recovery Test)

1. Circuit is open
2. After 30s timeout, circuit transitions to half-open
3. **Expected:** 3 test requests are allowed through
4. **Expected:** If test requests succeed → circuit closes
5. **Expected:** If test requests fail → circuit re-opens, timer restarts

### 3.1.9 Embedding — Batch Processing

1. Ingest a large document (100+ chunks)
2. **Expected:** Chunks are embedded in batches of 25
3. **Expected:** All chunks are successfully embedded and stored
4. **Expected:** No rate-limit errors from OpenAI

### 3.1.10 Embedding — Rate Limit Handling

1. OpenAI returns 429 rate limit error
2. **Expected:** Embedding process backs off and retries
3. **Expected:** Eventually all chunks are embedded
4. **Expected:** No data loss

### 3.1.11 Ingestion — Deterministic Point IDs

1. Ingest the same source twice
2. **Expected:** Point IDs are the same (deterministic from `MakePointID(sourceID, index)`)
3. **Expected:** Upsert is idempotent — no duplicate vectors
4. **Expected:** Qdrant collection has same number of points after re-ingestion

### 3.1.12 Chunking — Sentence Awareness

1. Ingest content with sentences of varying lengths
2. **Expected:** No chunk splits mid-sentence
3. **Expected:** Overlap regions align with sentence boundaries where possible

### 3.1.13 Ingestion — PDF Content

1. Upload a multi-page PDF
2. **Expected:** PDF text is extracted (via fitz/CGO)
3. **Expected:** Text is chunked and embedded
4. **Expected:** Questions about PDF content return accurate answers
5. **Expected:** PDF metadata (filename, page numbers) is preserved

### 3.1.14 Ingestion — URL Content

1. Submit a URL for ingestion
2. **Expected:** Content is fetched via HTTP
3. **Expected:** HTML is cleaned (boilerplate removed)
4. **Expected:** Main content is extracted and chunked
5. **Expected:** SSRF protection is applied to the URL
