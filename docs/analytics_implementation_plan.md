# Analytics Implementation Plan

## 1. Current State Analysis
- **Database**: An `analytics` table exists with columns for daily stats (`total_conversations`, `total_messages`, `thumbs_up_count`, etc.).
- **API**: An endpoint `GET /api/v1/analytics` exists that queries this table.
- **Problem**: **The table is never populated.** The current chat flow updates `conversations` and `messages` tables but does not update the `analytics` table. The analytics dashboard is currently showing empty or zero data.

## 2. Implementation Strategy
We will implement a **Real-time Aggregation (Upsert)** strategy. This is suitable for the current scale and provides immediate feedback to users on the dashboard.

### 2.1. Database Changes
No schema changes are strictly necessary, but we need to ensure the `UNIQUE(chatbot_id, analytics_date)` constraint exists to support upserts.

### 2.2. Backend Logic Updates
We need to modify the `Chat` handler (or a service layer called by it) to update the `analytics` table whenever a message is sent.

#### New Function: `IncrementAnalytics`
Located in `internal/db/analytics.go` (to be created).

```go
func IncrementAnalytics(ctx context.Context, pool *sql.DB, chatbotID string, date time.Time, isNewConversation bool, tokens int) error {
    // Logic:
    // INSERT INTO analytics (chatbot_id, analytics_date, total_messages, total_conversations, tokens_used)
    // VALUES ($1, $2, 1, $3, $4)
    // ON CONFLICT (chatbot_id, analytics_date)
    // DO UPDATE SET
    //    total_messages = analytics.total_messages + 1,
    //    total_conversations = analytics.total_conversations + EXCLUDED.total_conversations,
    //    tokens_used = analytics.tokens_used + EXCLUDED.tokens_used;
}
```

#### Update `Chat` Handler
In `internal/api/handlers/chat.go`:
1.  After successfully saving a user message and assistant response.
2.  Call `IncrementAnalytics`.
    -   If it was a new conversation (message count was 0), increment `total_conversations`.
    -   Always increment `total_messages` (by 2: user + assistant, or 1 per message).
    -   Update `tokens_used`.

#### Feedback Handling
We also need to update analytics when a user gives a thumbs up/down.
-   **Endpoint**: `POST /api/v1/messages/:id/feedback` (Needs to be implemented).
-   **Logic**: Update `messages` table AND increment `thumbs_up_count` / `thumbs_down_count` in `analytics` table.

## 3. Detailed Tasks

### Phase 1: Core Analytics (Message & Conversation Counts)
1.  [ ] Create `internal/db/analytics.go`.
2.  [ ] Implement `UpsertAnalytics` function using PostgreSQL `ON CONFLICT`.
3.  [ ] Modify `ChatHandlers.Chat` to call `UpsertAnalytics` after a successful chat exchange.
    -   Need to detect if this is a new conversation session to increment `total_conversations` correctly.

### Phase 2: Token Usage Tracking
1.  [ ] Ensure `tokens_used` column exists in `analytics` table (it might be missing or named differently, check schema).
    -   *Correction*: Schema has `average_tokens_per_message`. We should probably track `total_tokens` and calculate average on read, or update the running average. **Better approach**: Add `total_tokens_used` column to `analytics` table.
2.  [ ] Update `UpsertAnalytics` to include token counts.

### Phase 3: Feedback System
1.  [ ] Implement `FeedbackHandler` in backend.
2.  [ ] Add route `POST /api/v1/messages/{id}/feedback`.
3.  [ ] Update `analytics` table on feedback.

## 4. Future Considerations (Scale)
If write load becomes too high for real-time upserts:
1.  **Buffer**: Write events to Redis or a Go channel.
2.  **Batch**: A background worker flushes aggregated stats to Postgres every 1-5 minutes.
