> Extracted from: rag-pipeline.md — Section: 1. Feature Overview

# 1. Feature Overview

Retrieval-Augmented Generation pipeline for answering end-user questions. Ingests content (HTML, PDF, raw text), chunks it (sentence-aware, 512 tokens, 15% overlap), embeds via OpenAI `text-embedding-3-small`, stores vectors in Qdrant with deterministic point IDs, and answers queries through a tiered similarity search + LLM generation (OpenAI or OpenRouter) protected by a circuit breaker.
