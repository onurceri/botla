# Plan: Knowledge Base 2.0 - Transparency & Control

## Phase 1: Backend - Chunk Inspection & Re-Sync

- [ ] Task: Create `GetSourceChunks` API Handler
    - **Goal:** Expose the internal chunks for a specific source to the frontend.
    - **Implementation:**
        - Create a new handler method `GetSourceChunks` in `internal/api/handlers/source.go`.
        - If chunks are stored in Postgres (check `chatbot_source_chunks`), query them with pagination.
        - If chunks are only in Qdrant, use the Qdrant Scroll API to retrieve them.
        - Ensure strict ownership checks (User -> Org -> Chatbot -> Source).
    - **Test:** Unit test the handler with mocked DB/Qdrant.

- [ ] Task: Create `RefreshSource` API Handler
    - **Goal:** Allow manual triggering of the ingestion process.
    - **Implementation:**
        - Create `RefreshSource` in `internal/api/handlers/source.go`.
        - Reuse the existing `processing.ProcessSource` logic but trigger it on demand.
        - Update the source status to `pending` immediately.
    - **Test:** Unit test to verify the job is enqueued/started.

- [ ] Task: Enhance Error Reporting in Ingestion
    - **Goal:** Capture specific errors during scraping/PDF processing.
    - **Implementation:**
        - Modify `internal/scraper` and `internal/pdf` to return typed/detailed errors.
        - Update `internal/processing/processor.go` to save the specific error message to the `sources` table (add `last_error` column if missing, or use a status details JSON field).

- [ ] Task: Conductor - User Manual Verification 'Backend - Chunk Inspection & Re-Sync' (Protocol in workflow.md)

## Phase 2: Frontend - Source List & Status

- [ ] Task: Update `useSources` Query & Types
    - **Goal:** Fetch the new error details and handle the new API endpoints.
    - **Implementation:**
        - Update `frontend/src/api/sources.ts` (or equivalent) to add `getChunks(sourceId)` and `refreshSource(sourceId)`.
        - Update the `Source` interface to include `last_error` or detailed status fields.

- [ ] Task: Enhance Source List UI
    - **Goal:** Display detailed status and add action buttons.
    - **Implementation:**
        - Modify `frontend/src/features/chatbot/pages/tabs/SourcesTab.tsx`.
        - Update the status badge to show a tooltip with `last_error` if the status is "failed".
        - Add a "Sync Now" button for URL sources (disabled if status is "pending").

- [ ] Task: Conductor - User Manual Verification 'Frontend - Source List & Status' (Protocol in workflow.md)

## Phase 3: Frontend - Chunk Inspector

- [ ] Task: Create `ChunkInspector` Component
    - **Goal:** A modal/slide-over to view and search chunks.
    - **Implementation:**
        - Create `frontend/src/features/chatbot/components/ChunkInspector.tsx`.
        - Use a Dialog or Sheet component from Radix UI.
        - Implement a paginated list of text chunks.
        - Add a simple text search filter (client-side or server-side depending on API).

- [ ] Task: Integrate Inspector into Source List
    - **Goal:** Connect the "View Chunks" button to the Inspector.
    - **Implementation:**
        - Add state to `SourcesTab` to track the `selectedSourceForInspection`.
        - Render the `ChunkInspector` when a source is selected.

- [ ] Task: Conductor - User Manual Verification 'Frontend - Chunk Inspector' (Protocol in workflow.md)

## Phase 4: Polish & Integration

- [ ] Task: E2E Testing of Ingestion Flow
    - **Goal:** Verify the full loop: Add Source -> Manual Sync -> View Chunks.
    - **Implementation:**
        - Write a Playwright test that adds a test URL, clicks "Sync Now", waits for success, and then opens the Chunk Inspector to verify content exists.

- [ ] Task: Conductor - User Manual Verification 'Polish & Integration' (Protocol in workflow.md)
