# Comprehensive System Analysis & Documentation

## 1. System Overview

**Botla.co** is a SaaS platform that allows users to create AI-powered chatbots trained on their own data (PDFs, URLs, Text). The system consists of a Go backend and a React frontend.

### Architecture
- **Backend**: Go (Golang) using `net/http` for the API server.
- **Database**: PostgreSQL for relational data (Users, Chatbots, Sources, Conversations).
- **Vector Database**: Qdrant for storing embeddings.
- **Frontend**: React (Vite) with TypeScript.
- **AI Integration**: OpenAI API for embeddings and chat completions.
- **Background Processing**: Asynchronous queue for processing sources (PDF parsing, scraping, chunking, embedding).

---

## 2. User Flows

### 2.1. Authentication
**Flows:**
1.  **Registration**:
    -   **Endpoint**: `POST /api/v1/auth/register`
    -   **Input**: `email`, `password`, `full_name`
    -   **Process**: Hashes password, creates user in DB.
    -   **Output**: Returns `token` (Access Token) and `refresh_token`.
2.  **Login**:
    -   **Endpoint**: `POST /api/v1/auth/login`
    -   **Input**: `email`, `password`
    -   **Process**: Verifies credentials.
    -   **Output**: Returns `token` and `refresh_token`.
3.  **Token Refresh**:
    -   **Endpoint**: `POST /api/v1/auth/refresh`
    -   **Input**: `refresh_token`
    -   **Process**: Verifies refresh token, checks revocation, rotates token (revokes old, issues new).
    -   **Output**: Returns new `token` and `refresh_token`.
    -   **Frontend Behavior**: Axios interceptor catches 401s, attempts refresh, and retries original request.
4.  **Logout**:
    -   **Endpoint**: `POST /api/v1/auth/logout`
    -   **Input**: `refresh_token`
    -   **Process**: Revokes the refresh token in DB.

### 2.2. Chatbot Management
**Flows:**
1.  **List Chatbots**:
    -   **Endpoint**: `GET /api/v1/chatbots`
    -   **Process**: Returns all chatbots belonging to the authenticated user.
2.  **Create Chatbot**:
    -   **Endpoint**: `POST /api/v1/chatbots`
    -   **Input**: `name` (required), optional settings (`system_prompt`, `model`, etc.).
    -   **Process**: Creates chatbot with default values if not provided.
3.  **Get Chatbot Details**:
    -   **Endpoint**: `GET /api/v1/chatbots/:id`
    -   **Process**: Returns full chatbot configuration.
4.  **Update Chatbot**:
    -   **Endpoint**: `PUT /api/v1/chatbots/:id`
    -   **Input**: Any subset of chatbot fields.
    -   **Process**: Updates specified fields.
5.  **Delete Chatbot**:
    -   **Endpoint**: `DELETE /api/v1/chatbots/:id`
    -   **Process**: Soft deletes the chatbot (sets `deleted_at`).

### 2.3. Data Source Management
**Flows:**
1.  **List Sources**:
    -   **Endpoint**: `GET /api/v1/chatbots/:id/sources`
    -   **Process**: Returns all sources for a specific chatbot.
2.  **Add Source**:
    -   **Endpoint**: `POST /api/v1/chatbots/:id/sources`
    -   **Input**: `source_type` ("pdf", "url", "text") and corresponding data.
    -   **Process**:
        -   **PDF**: Uploads file to `/tmp/uploads`.
        -   **URL**: Stores URL.
        -   **Text**: Saves content to a `.txt` file in `/tmp/uploads`.
        -   **Queue**: Enqueues the source ID for background processing.
    -   **Output**: Returns `id` of the new source.
3.  **Check Status**:
    -   **Endpoint**: `GET /api/v1/sources/:id`
    -   **Process**: Returns current status (`pending`, `processing`, `completed`, `failed`).
    -   **Frontend Behavior**: Polls this endpoint until status is terminal.
4.  **Delete Source**:
    -   **Endpoint**: `DELETE /api/v1/sources/:id`
    -   **Process**: Deletes source record from DB and attempts to delete associated vectors from Qdrant.

### 2.4. Chat Interaction (RAG Pipeline)
**Flows:**
1.  **Send Message**:
    -   **Endpoint**: `POST /api/v1/chatbots/:id/chat`
    -   **Input**: `message`, `session_id`.
    -   **Process**:
        1.  Validates ownership/existence.
        2.  Gets/Creates `Conversation` based on `session_id`.
        3.  Saves User Message to DB.
        4.  **Retrieval**: Generates embedding for user message -> Searches Qdrant for relevant chunks.
        5.  **Generation**: Constructs prompt with context -> Calls OpenAI Chat Completion.
        6.  Saves Assistant Message to DB.
    -   **Output**: Returns `response`, `tokens_used`, and `sources_used` (citations).

---

## 3. Data Models (Key Entities)

### Users
- `id`, `email`, `password_hash`, `full_name`
- **Auth**: Uses JWT (Access + Refresh tokens).

### Chatbots
- `id`, `user_id`
- **Settings**: `name`, `system_prompt`, `model` (default: gpt-3.5-turbo), `temperature`, `max_tokens`.
- **Styling**: `theme_color`, `welcome_message`.

### Data Sources
- `id`, `chatbot_id`, `source_type`
- **Status**: `pending` -> `processing` -> `completed` / `failed`.
- **Storage**: Files stored temporarily in `/tmp/uploads` for processing.

### Conversations & Messages
- **Conversation**: Tracks a session (`session_id`) for a visitor.
- **Message**: Individual messages (`role`: user/assistant) with token usage tracking.

---

## 4. Error Handling & Edge Cases

### Backend
- **Authentication**:
    -   Invalid credentials -> `401 Unauthorized`.
    -   Expired access token -> `401 Unauthorized` (Frontend handles refresh).
    -   Invalid/Revoked refresh token -> `401 Unauthorized`.
- **Input Validation**:
    -   Missing required fields -> `400 Bad Request`.
    -   Invalid file types (non-PDF for PDF source) -> `400 Bad Request`.
    -   File size limits (50MB) -> `413 Request Entity Too Large` (handled manually in code).
- **Resource Access**:
    -   Accessing another user's chatbot -> `403 Forbidden`.
    -   Resource not found -> `404 Not Found`.
- **Processing Failures**:
    -   If background processing fails, source status is set to `failed` with an error message.
- **External Services**:
    -   OpenAI/Qdrant failures -> `500 Internal Server Error` (graceful degradation in some cases, e.g., chat without context if Qdrant fails).

### Frontend
- **Network Errors**: Generic error handling (often silent or console logs).
- **Auth Expiry**: Auto-redirect to `/login` if refresh fails.
- **Polling**: Retries status check 50 times with 400ms delay.

---

## 5. Current Limitations & Observations
1.  **File Storage**: Files are stored in `/tmp/uploads`. This is ephemeral and will be lost on container restart. Not suitable for production persistence if re-processing is needed.
2.  **Vector Deletion**: Deleting a source attempts to delete vectors, but if it fails, the DB record is still deleted ("Best-effort"). This could lead to orphaned vectors.
3.  **Hardcoded Paths**: `/tmp/uploads` is hardcoded.
4.  **CORS**: Configured to allow origins from env `CORS_ALLOWED_ORIGINS`.
5.  **Widget**: Widget code generation is simple string interpolation.
6.  **Analytics**: Basic analytics endpoint exists but implementation details were not deeply inspected (likely aggregates message counts).
