> Extracted from: background-processing.md — Section: 1. Feature Overview

# 1. Feature Overview

The Background Processing subsystem handles asynchronous ingestion of content sources through a bounded worker pool with per-source-type processors (URL, PDF, Text). Each source goes through a pipeline: fetch → parse → chunk → embed → store vectors in Qdrant. The system includes suggestion generation, duplicate detection, automatic retries with exponential backoff, and startup recovery for stuck jobs.

**Key components:**
- **Source Queue** — Bounded worker pool (max 16 workers, 64-job buffer) with per-type processors
- **URL Processor** — Web scraping with static/dynamic rendering, link discovery, content normalization
- **PDF Processor** — Text extraction via fitz CGO, OCR support for Pro+ plans
- **Text Processor** — Raw text chunking and embedding
- **Suggestion Generation** — Capability summary and suggested questions extracted during ingestion
- **Duplicate Detection** — Content hash-based and URL-normalization-based deduplication
- **Worker Pool** — Configurable concurrency with rate limiting between jobs
