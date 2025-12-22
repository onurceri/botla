# Plan: Stability, Resiliency, and Cost-Efficient Testing

## Phase 1: Mocking Infrastructure & Service Abstraction
Goal: Ensure all external dependencies are behind interfaces and create mocks to enable offline, zero-cost testing.

- [x] Task: Audit and Refactor LLM Service Abstraction
    - [x] Identify direct OpenAI/OpenRouter calls in `internal/rag` and `internal/services`.
    - [x] Create or update `LLMService` and `EmbeddingService` interfaces.
    - [x] Ensure all callers use the interfaces instead of direct client instances.
- [x] Task: Audit and Refactor Storage & Vector DB Abstractions
    - [x] Identify direct AWS/R2 calls and Qdrant calls.
    - [x] Create or update `StorageService` and `VectorService` interfaces.
- [x] Task: Generate/Implement Mocks using `testify/mock`
    - [x] Implement `MockLLMService`.
    - [x] Implement `MockEmbeddingService`.
    - [x] Implement `MockStorageService`.
    - [x] Implement `MockVectorService`.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Mocking Infrastructure & Service Abstraction' (Protocol in workflow.md)

## Phase 2: Unit Testing Core Services
Goal: Reach high coverage for business logic using the new mocking infrastructure.

- [x] Task: Implement Unit Tests for RAG Pipeline
    - [x] Test retrieval logic with `MockVectorService`.
    - [x] Test response generation logic with `MockLLMService`.
- [x] Task: Implement Unit Tests for Document Processing
    - [x] Test PDF/Text parsing and splitting logic.
    - [x] Mock `StorageService` for ingestion tests.
- [x] Task: Implement Unit Tests for Web Scraper
    - [x] Use `httptest` to mock website responses.
    - [x] Test edge cases (timeouts, invalid content).
- [x] Task: Conductor - User Manual Verification 'Phase 2: Unit Testing Core Services' (Protocol in workflow.md)

## Phase 3: Integration Testing for API Handlers
Goal: Verify end-to-end API flows using a test database and mocked external services.

- [x] Task: Implement Auth & Identity Integration Tests
    - [x] Test JWT generation, validation, and multi-tenant isolation.
- [x] Task: Implement Data Ingestion Integration Tests
    - [x] Test full flow from scraping/uploading to vector storage (using mocks).
- [x] Task: Implement AI Chat Execution Integration Tests
    - [x] Test the `/chat` endpoint flow: query processing -> retrieval -> LLM response.
- [x] Task: Conductor - User Manual Verification 'Phase 3: Integration Testing for API Handlers' (Protocol in workflow.md)

## Phase 4: Coverage & Finalization
Goal: Ensure the project meets the 90% coverage target and can run tests without external costs.

- [x] Task: Verify Zero-Cost Testing
    - [x] Run `make test-all` with invalid/empty credentials for OpenAI/R2 and verify it passes.
- [x] Task: Coverage Audit & Final Improvements
    - [x] Increase coverage through unit and integration tests.
- [x] Task: Conductor - User Manual Verification 'Phase 4: Coverage & Finalization' (Protocol in workflow.md)

