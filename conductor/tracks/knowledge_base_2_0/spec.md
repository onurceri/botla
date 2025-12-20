# Specification: Knowledge Base 2.0 - Transparency & Control

## 1. Overview
The goal of this track is to elevate the "Knowledge Base" (Source Management) experience from "functional" to "transparent." Users currently ingest data but lack visibility into *what* was actually indexed or *why* a source failed. This track implements a "Chunk Inspector" and detailed error reporting.

## 2. User Stories
- **As a User**, I want to see the specific text chunks extracted from my PDF or URL so that I can verify the AI is reading my content correctly.
- **As a User**, I want to see detailed error messages (e.g., "403 Forbidden", "PDF Encrypted") if a source fails to sync, so I can fix the issue.
- **As a User**, I want to search within my indexed chunks to find specific information.

## 3. Functional Requirements

### 3.1. Source Status & Error Reporting
- **Backend:** Update the `sources` table or status logic to capture detailed error messages from the scraper/ingestor.
- **Frontend:** Update the Source List UI to display a tooltip or status message with the specific error (e.g., instead of just red "Error" badge, show "Error: Timeout crawling URL").

### 3.2. Chunk Inspector
- **Backend:** Create a new API endpoint `GET /api/v1/sources/{source_id}/chunks` that returns the paginated list of chunks (text content + metadata) for a given source from Qdrant/Postgres.
- **Frontend:**
    - Add a "View Chunks" button to each source in the dashboard.
    - Create a modal or slide-over panel that lists the chunks.
    - Include a search bar in this panel to filter chunks by text content.

## 4. Technical Constraints
- **Performance:** Chunk retrieval must be paginated (e.g., 20 chunks per page) to avoid overloading the browser or API.
- **Security:** Ensure users can only inspect chunks for sources belonging to their own chatbots/organization.
- **Database:** Reuse existing `chatbot_source_chunks` or Qdrant scrolling API.

## 5. UI/UX Design
- **Source List:** Enhance the table row to include "Status", "Last Synced", "Chunk Count", and "Actions" (View Chunks, Delete).
- **Chunk Inspector Modal:** A clean, scrollable list. Each chunk should show the raw text and its associated "Score" (if applicable from a search context) or just the indexed text.

## 6. Metrics & Analytics
- Track how often users open the Chunk Inspector (engagement).
