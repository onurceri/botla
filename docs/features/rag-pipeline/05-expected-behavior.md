> Extracted from: rag-pipeline.md — Section: 5. Expected Behavior

# 5. Expected Behavior — Success/Failure Conditions

## 5.1 RAG Pipeline

| Condition | Success Indicator | Failure Indicator |
|---|---|---|
| Content ingested | Chunks embedded, Qdrant has points | Error in ingestion queue, empty collection |
| Query answered | Response with citations, correct answer | Hallucination, irrelevant answer, fallback message |
| Strong match (≥ 0.50) | Confident answer with sources | Unsure answer, low-quality citations |
| Medium match (0.30–0.49) | Qualified answer with confidence caveat | Overconfident incorrect answer |
| No match (< 0.30) | Appropriate fallback response | Hallucination, made-up answer |
| Circuit breaker closed | Fast response, LLM called | Latency from repeated failures |
| Circuit breaker open | Fast failure ("Service unavailable") | Attempted call to failing LLM |
| Circuit breaker half-open | Limited test requests; recovery or re-open | All-or-nothing behavior |
