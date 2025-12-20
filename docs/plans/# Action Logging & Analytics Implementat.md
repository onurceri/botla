# Action Logging & Analytics Implementation Plan

**Status: Implemented**

## Overview
Currently, the system executes Chatbot Actions (HTTP, Zapier, etc.) but does not persist any record of these executions. This makes it impossible for users to:
1. Debug failed action calls.
2. View the history of what data was sent/received.
3. Analyze usage patterns (which actions are used most, failure rates, etc.).

This plan details the implementation of a comprehensive logging system for Action executions, including storing request/response payloads for debugging purposes.

## 1. Database Schema Changes

We will introduce a new table `action_execution_logs` to store the details of every action attempt.

### New Table: `action_execution_logs`

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID (PK) | Unique identifier for the log entry |
| `chatbot_id` | UUID (FK) | Reference to the chatbot |
| `action_id` | UUID (FK) | Reference to the specific action definition |
| `conversation_id` | UUID (FK) | Reference to the conversation where it occurred |
| `message_id` | UUID (FK) | Reference to the assistant message (optional/nullable) |
| `status` | VARCHAR | `success` or `failure` |
| `request_payload` | JSONB | The arguments passed to the tool/action |
| `response_payload` | JSONB | The result returned from the tool or error details |
| `error_message` | TEXT | Human-readable error message (if any) |
| `duration_ms` | INT | Execution time in milliseconds |
| `created_at` | TIMESTAMPTZ | When the execution finished |

**Indexes:**
- `idx_action_logs_chatbot_created` (chatbot_id, created_at DESC) for efficient history retrieval.
- `idx_action_logs_action_created` (action_id, created_at DESC) for filtering by specific action.

## 2. Backend Implementation

### 2.1 Models (`internal/models/action.go`)
Define the `ActionExecutionLog` struct matching the DB schema.

```go
type ActionExecutionLog struct {
    ID              string          `json:"id"`
    ChatbotID       string          `json:"chatbot_id"`
    ActionID        string          `json:"action_id"`
    ConversationID  string          `json:"conversation_id"`
    Status          string          `json:"status"` // "success", "failure"
    RequestPayload  json.RawMessage `json:"request_payload"`
    ResponsePayload json.RawMessage `json:"response_payload"`
    ErrorMessage    *string         `json:"error_message,omitempty"`
    DurationMs      int             `json:"duration_ms"`
    CreatedAt       time.Time       `json:"created_at"`
}
```

### 2.2 Database Layer (`internal/db/action_logs.go`)
Implement functions to insert and query logs.
- `CreateActionLog(ctx, db, log)`
- `GetActionLogs(ctx, db, chatbotID, limit, offset)`
- `GetActionLogStats(ctx, db, chatbotID, days)` (Aggregation for analytics)

### 2.3 Tool Executor Update (`internal/rag/tool_executor.go`)
Modify the `Execute` method in `ToolExecutor` to measure time and capture results.

**Logic Flow:**
1. Start timer (`start := time.Now()`).
2. Execute the tool (HTTP, Zapier, etc.).
3. Capture error or success result.
4. Stop timer.
5. Asynchronously (or synchronously, depending on criticality) write to `action_execution_logs`.
   - *Note:* We must ensure sensitive data is handled according to policy, though for now, we will store full payloads for debugging.

### 2.4 API Endpoints (`internal/api/handlers/action.go` or `analytics.go`)

**GET /api/v1/chatbots/{id}/actions/logs**
- Returns a paginated list of execution logs.
- Query params: `page`, `limit`, `status` (success/failure), `action_id`.

**GET /api/v1/chatbots/{id}/analytics/actions**
- Returns aggregated stats:
  - Total executions
  - Success rate
  - Average duration
  - Breakdown by Action ID

## 3. Frontend Implementation

### 3.1 Action List Update
- Add a "History" or "Logs" tab in the `ActionsTab` component.
- Alternatively, add a global "Action Logs" section in the Analytics dashboard, but keeping it near the Actions configuration is better for debugging.

### 3.2 Logs UI
- **Table View:**
  - Date/Time
  - Action Name
  - Status (Green check / Red X)
  - Duration
  - Triggered by (Session ID)
- **Detail Modal:**
  - Clicking a row opens a modal/drawer.
  - Shows `Request JSON` and `Response JSON` with syntax highlighting.
  - Shows full error message if failed.

## 4. Considerations & Limits

### 4.1 Storage Limits
- JSON payloads can be large. We should enforce a soft limit (e.g., truncate if > 100KB) to prevent DB bloat.
- Retention Policy: We may want to auto-delete logs older than 30 days or 90 days depending on the plan level (to be decided later, but schema should allow for easy cleanup).

### 4.2 Failure Handling
- If the *logging* itself fails (DB error), it should not fail the *action execution* seen by the user/bot. Logging should be best-effort.

## 5. Execution Steps

1.  **Migration:** Create the SQL migration file for `action_execution_logs`.
2.  **Backend:** Implement DB functions and update `ToolExecutor`.
3.  **API:** Expose the logs via REST API.
4.  **Frontend:** Build the UI to view these logs.