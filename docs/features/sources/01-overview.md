> Extracted from: sources.md — Section: 1. Feature Overview

# 1. Feature Overview

The Sources subsystem handles data ingestion for the botla.app AI chatbot platform. Users add content sources (URLs, PDFs, raw text, sitemaps) to their chatbots, which are then asynchronously processed through a pipeline: fetch → parse → chunk → embed → store vectors in Qdrant.

**Key components:**
- **Source Management** — CRUD for data sources scoped to chatbots
- **Sitemap Discovery** — Parse XML sitemaps to discover URLs for ingestion
