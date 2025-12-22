# Specification: Stability, Resiliency, and Cost-Efficient Testing

## Overview
This track focuses on improving the stability and trustworthiness of the Botla-co platform by significantly increasing test coverage while implementing a robust mocking strategy. The goal is to enable frequent, safe code changes without incurring high costs from external LLM (OpenAI/OpenRouter) or Storage (Cloudflare R2) providers.

## Functional Requirements

### 1. Mocking Infrastructure
- **Interface-based Mocks:** Transition all external service dependencies (LLMs, Vector DB, Object Storage) to Go interfaces if not already done.
- **Mock Implementations:** Create standard mock implementations for:
    - `LLMService` (OpenAI/OpenRouter)
    - `EmbeddingService` (OpenAI)
    - `StorageService` (Cloudflare R2/S3)
    - `VectorDBService` (Qdrant) - though Qdrant runs in Docker, mocking is preferred for fast unit tests.

### 2. Enhanced Testing Suite
- **Unit Tests:** Achieve >= 90% coverage for the RAG pipeline, document processing logic, and business services.
- **Integration Tests:** Implement comprehensive integration tests for:
    - Authentication & Identity (JWT, multi-tenant isolation).
    - Data Ingestion & Processing (Scraping/Uploading -> Processing -> Vector Storage).
    - Chatbot Management (CRUD).
    - AI Chat Execution (Query -> Retrieval -> LLM Response).
- **Cost-Free Test Environment:** Ensure `make test-all` can run entirely offline (or against local Docker containers) without requiring real API keys for LLMs or Storage.

## Non-Functional Requirements
- **Performance:** Tests should be fast; unit tests should rely solely on mocks.
- **Determinism:** Tests must be reliable and produce the same result every time.
- **Maintainability:** Use `testify/mock` and standard Go patterns to keep tests readable.

## Acceptance Criteria
- [ ] All external service calls are abstracted behind interfaces.
- [ ] Mocks exist for LLM, Storage, and Vector DB services.
- [ ] Integration tests cover the full lifecycle of a chatbot (Create -> Ingest -> Chat).
- [ ] `make test-all` passes without valid `OPENAI_API_KEY` or `AWS_ACCESS_KEY_ID` (using mocks instead).
- [ ] Overall backend code coverage reaches or exceeds 90%.

## Out of Scope
- Implementing new product features.
- Performance optimization of the production RAG pipeline (unless required for stability).
- UI/UX changes to the dashboard.
