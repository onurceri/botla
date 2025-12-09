# Source Usage Tracking Implementation Plan

## Overview
Implement comprehensive source tracking to persist which specific data sources are used in RAG responses, enabling analytics on source effectiveness and usage patterns.

## Background
Currently, we:
- Return source citations in API responses (`ChatResult.Sources`)
- Do NOT persist source usage in the database
- Have `ChunkMetadata` with `SourceID`, `SourceType`, `ChunkIndex`, and `Score` from RAG search

## Goals
1. Persist source usage data for each assistant message
2. Enable analytics queries on:
   - Which sources are most frequently used
   - Which sources contribute to successful responses (based on feedback)
   - Source usage trends over time
   - Per-chatbot source effectiveness

---

## Phase 1: Database Schema

### Migration: Add Message Sources Table

Create a new junction table to store the many-to-many relationship between messages and sources.

```sql
-- Migration: XXXXXX_add_message_sources.up.sql
CREATE TABLE message_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    source_id UUID NOT NULL REFERENCES data_sources(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    relevance_score FLOAT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(message_id, source_id, chunk_index)
);

CREATE INDEX idx_message_sources_message_id ON message_sources(message_id);
CREATE INDEX idx_message_sources_source_id ON message_sources(source_id);
CREATE INDEX idx_message_sources_created_at ON message_sources(created_at);

-- Migration: XXXXXX_add_message_sources.down.sql
DROP TABLE IF EXISTS message_sources;
```

**Alternative Approach (JSONB):**
If preferred for simpler queries, add a JSONB column to `messages`:
```sql
ALTER TABLE messages ADD COLUMN sources_used JSONB;
CREATE INDEX idx_messages_sources_used ON messages USING gin(sources_used);
```

**Recommendation:** Use separate table for:
- Better normalization
- Easier analytics joins
- Ability to track additional metadata per source usage

---

## Phase 2: Backend Models

### Update `internal/models/message.go`

```go
type Message struct {
    ID             string    `json:"id"`
    ConversationID string    `json:"conversation_id"`
    Role           string    `json:"role"`
    Content        string    `json:"content"`
    TokensUsed     int       `json:"tokens_used"`
    ThumbsUp       *bool     `json:"thumbs_up,omitempty"`
    CreatedAt      time.Time `json:"created_at"`
    // NEW: For loading message with sources
    Sources        []MessageSource `json:"sources,omitempty"`
}

type MessageSource struct {
    ID             string    `json:"id"`
    MessageID      string    `json:"message_id"`
    SourceID       string    `json:"source_id"`
    ChunkIndex     int       `json:"chunk_index"`
    RelevanceScore float64   `json:"relevance_score"`
    CreatedAt      time.Time `json:"created_at"`
}
```

### Update `internal/models/chunk.go`

Already has `ChunkMetadata` with `SourceID` - no changes needed.

---

## Phase 3: Database Layer

### New file: `internal/db/message_sources.go`

```go
package db

import (
    "context"
    "database/sql"
    "github.com/onurceri/botla-co/internal/models"
)

// SaveMessageSources persists source usage for a message
func SaveMessageSources(ctx context.Context, pool *sql.DB, messageID string, sources []models.ChunkMetadata) error {
    if len(sources) == 0 {
        return nil
    }

    tx, err := pool.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer func() { _ = tx.Rollback() }()

    stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO message_sources (message_id, source_id, chunk_index, relevance_score)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (message_id, source_id, chunk_index) DO NOTHING
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, src := range sources {
        if src.SourceID == "" {
            continue // Skip if no source ID
        }
        _, err = stmt.ExecContext(ctx, messageID, src.SourceID, src.ChunkIndex, src.Score)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

// GetMessageSources retrieves sources used in a specific message
func GetMessageSources(ctx context.Context, pool *sql.DB, messageID string) ([]models.MessageSource, error) {
    query := `
        SELECT id, message_id, source_id, chunk_index, relevance_score, created_at
        FROM message_sources
        WHERE message_id = $1
        ORDER BY relevance_score DESC
    `
    rows, err := pool.QueryContext(ctx, query, messageID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var sources []models.MessageSource
    for rows.Next() {
        var s models.MessageSource
        if err := rows.Scan(&s.ID, &s.MessageID, &s.SourceID, &s.ChunkIndex, &s.RelevanceScore, &s.CreatedAt); err != nil {
            return nil, err
        }
        sources = append(sources, s)
    }
    return sources, rows.Err()
}
```

---

## Phase 4: Service Layer

### Update `internal/services/chat_service.go`

#### In `ProcessChat` method (around line 154-160):

```go
// Save assistant message
am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: ans, TokensUsed: tokens}
var amID string
if id, err := db.CreateMessage(ctx, s.DB, am); err == nil {
    amID = id
    _ = db.IncrementConversationMessageCount(ctx, s.DB, conv.ID)
    
    // NEW: Save source usage
    if len(sources) > 0 {
        if err := db.SaveMessageSources(ctx, s.DB, amID, sources); err != nil && s.Log != nil {
            s.Log.Warn("save_message_sources_error", map[string]any{"message_id": amID, "error": err.Error()})
        }
    }
}
```

#### In `ProcessChatWithTools` method (around line 330-336):

```go
// Save assistant message
am := &models.Message{ConversationID: conv.ID, Role: "assistant", Content: finalResponse, TokensUsed: totalTokens}
var amID string
if id, err := db.CreateMessage(ctx, s.DB, am); err == nil {
    amID = id
    _ = db.IncrementConversationMessageCount(ctx, s.DB, conv.ID)
    
    // NEW: Save source usage
    if len(sources) > 0 {
        if err := db.SaveMessageSources(ctx, s.DB, amID, sources); err != nil && s.Log != nil {
            s.Log.Warn("save_message_sources_error", map[string]any{"message_id": amID, "error": err.Error()})
        }
    }
}
```

---

## Phase 5: Analytics Layer

### New file: `internal/db/source_analytics.go`

```go
package db

import (
    "context"
    "database/sql"
)

// SourceUsageStats represents usage statistics for a data source
type SourceUsageStats struct {
    SourceID         string  `json:"source_id"`
    SourceName       string  `json:"source_name"`
    SourceType       string  `json:"source_type"`
    TimesUsed        int     `json:"times_used"`
    AvgRelevance     float64 `json:"avg_relevance"`
    PositiveFeedback int     `json:"positive_feedback"`
    NegativeFeedback int     `json:"negative_feedback"`
    LastUsed         string  `json:"last_used"`
}

// GetSourceUsageStats returns usage statistics for sources of a chatbot
func GetSourceUsageStats(ctx context.Context, pool *sql.DB, chatbotID string, days int) ([]SourceUsageStats, error) {
    query := `
        SELECT 
            ds.id as source_id,
            ds.name as source_name,
            ds.source_type,
            COUNT(DISTINCT ms.message_id) as times_used,
            AVG(ms.relevance_score) as avg_relevance,
            COUNT(CASE WHEN m.thumbs_up = true THEN 1 END) as positive_feedback,
            COUNT(CASE WHEN m.thumbs_up = false THEN 1 END) as negative_feedback,
            MAX(ms.created_at) as last_used
        FROM data_sources ds
        INNER JOIN message_sources ms ON ds.id = ms.source_id
        INNER JOIN messages m ON ms.message_id = m.id
        INNER JOIN conversations c ON m.conversation_id = c.id
        WHERE c.chatbot_id = $1
          AND ms.created_at >= CURRENT_DATE - ($2 || ' days')::interval
        GROUP BY ds.id, ds.name, ds.source_type
        ORDER BY times_used DESC
    `
    
    rows, err := pool.QueryContext(ctx, query, chatbotID, days)
    if err != nil {
        return nil, err
    }
    defer func() { _ = rows.Close() }()

    var stats []SourceUsageStats
    for rows.Next() {
        var s SourceUsageStats
        if err := rows.Scan(&s.SourceID, &s.SourceName, &s.SourceType, &s.TimesUsed, 
            &s.AvgRelevance, &s.PositiveFeedback, &s.NegativeFeedback, &s.LastUsed); err != nil {
            return nil, err
        }
        stats = append(stats, s)
    }
    return stats, rows.Err()
}
```

---

## Phase 6: API Layer

### New endpoint in `internal/api/handlers/analytics.go`

```go
// GetSourceUsage returns source usage analytics for a chatbot
func (h *AnalyticsHandlers) GetSourceUsage(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 7 {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    botID := parts[4]

    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }

    // Access check (same as existing analytics endpoints)
    bot, err := db.GetChatbotByID(r.Context(), h.DB, botID)
    if err != nil || bot == nil {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    // RBAC check (workspace/org membership)
    allowed := checkBotAccess(r.Context(), h.OrgService, bot, userID)
    if !allowed {
        w.WriteHeader(http.StatusForbidden)
        return
    }

    // Parse days param
    days := 30
    if d := r.URL.Query().Get("days"); d != "" {
        if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= 365 {
            days = n
        }
    }

    stats, err := db.GetSourceUsageStats(r.Context(), h.DB, botID, days)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}
```

### Update router in `cmd/server/main.go`

```go
// In chatbotsDispatchHandlerWithSourcesRL function
if strings.HasSuffix(r.URL.Path, "/analytics/sources") {
    anh.GetSourceUsage(w, r)
    return
}
```

---

## Phase 7: Frontend Implementation

### New Component: `frontend/src/features/analytics/SourceUsageStats.tsx`

```tsx
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'

interface SourceStat {
  source_id: string
  source_name: string
  source_type: string
  times_used: number
  avg_relevance: number
  positive_feedback: number
  negative_feedback: number
  last_used: string
}

export function SourceUsageStats({ chatbotId }: { chatbotId: string }) {
  const [stats, setStats] = useState<SourceStat[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchSourceStats()
  }, [chatbotId])

  const fetchSourceStats = async () => {
    const res = await fetch(`/api/v1/chatbots/${chatbotId}/analytics/sources?days=30`)
    const data = await res.json()
    setStats(data)
    setLoading(false)
  }

  const getFeedbackRate = (pos: number, neg: number) => {
    const total = pos + neg
    return total > 0 ? ((pos / total) * 100).toFixed(0) : '-'
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Kaynak Kullanım İstatistikleri</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Kaynak</TableHead>
              <TableHead>Tip</TableHead>
              <TableHead className="text-right">Kullanım</TableHead>
              <TableHead className="text-right">Ortalama İlgi</TableHead>
              <TableHead className="text-right">Memnuniyet</TableHead>
              <TableHead>Son Kullanım</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {stats.map((stat) => (
              <TableRow key={stat.source_id}>
                <TableCell className="font-medium">{stat.source_name}</TableCell>
                <TableCell>
                  <Badge variant="outline">{stat.source_type}</Badge>
                </TableCell>
                <TableCell className="text-right">{stat.times_used}</TableCell>
                <TableCell className="text-right">
                  {(stat.avg_relevance * 100).toFixed(1)}%
                </TableCell>
                <TableCell className="text-right">
                  {getFeedbackRate(stat.positive_feedback, stat.negative_feedback)}%
                </TableCell>
                <TableCell>{new Date(stat.last_used).toLocaleDateString('tr-TR')}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
```

### Integrate into `ChatbotAnalytics.tsx`

Add the `SourceUsageStats` component to the analytics tab:

```tsx
<SourceUsageStats chatbotId={chatbotId} />
```

---

## Verification Plan

### Unit Tests

1. **`internal/db/message_sources_test.go`**:
   - Test `SaveMessageSources` with valid data
   - Test deduplication (ON CONFLICT)
   - Test `GetMessageSources` returns correct data

2. **`internal/db/source_analytics_test.go`**:
   - Test `GetSourceUsageStats` aggregates correctly
   - Test date filtering
   - Test feedback correlation

### Integration Tests

1. **Chat Flow Test**:
   - Send message → Verify sources saved to `message_sources`
   - Check `source_id`, `chunk_index`, `relevance_score` are populated

2. **Analytics Test**:
   - Create test data with known source usage
   - Call analytics endpoint
   - Verify stats match expectations

### Manual Verification

1. **Database Check**:
   ```sql
   SELECT * FROM message_sources LIMIT 10;
   SELECT COUNT(*) FROM message_sources;
   ```

2. **API Test**:
   ```bash
   curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v1/chatbots/{id}/analytics/sources
   ```

3. **UI Verification**:
   - Navigate to chatbot analytics
   - Verify source stats table renders
   - Check feedback correlation is accurate

---

## Migration Strategy

### Phase 1: Deploy Schema (No Breaking Changes)
1. Run migration to create `message_sources` table
2. Deploy backend with source tracking code
3. **New messages** will track sources; **old messages** won't

### Phase 2: Backfill (Optional)
If desired, backfill historical data:
- Parse existing message contexts (if stored)
- Or accept that historical data doesn't have source tracking

### Phase 3: Enable Analytics
1. Verify data is being collected
2. Enable frontend components
3. Monitor for performance issues

---

## Performance Considerations

1. **Write Performance**:
   - Bulk insert in single transaction
   - Use prepared statement
   - Non-blocking (don't fail message creation if source save fails)

2. **Read Performance**:
   - Indexes on `message_id`, `source_id`, `created_at`
   - Consider materialized view for heavy analytics queries
   - Cache analytics results (1 hour TTL)

3. **Storage**:
   - Estimate: ~5 sources per message
   - 1M messages = 5M `message_sources` rows
   - With indexes: ~200MB for 1M messages

---

## Rollback Plan

If issues arise:
1. **Disable source saving**: Comment out `SaveMessageSources` calls
2. **Hide UI**: Remove `SourceUsageStats` component
3. **Drop table** (if needed): Run down migration

No data loss for core functionality as source tracking is additive.

---

## Timeline Estimate

- **Phase 1 (Schema)**: 1 hour
- **Phase 2 (Models)**: 30 min
- **Phase 3 (DB Layer)**: 2 hours
- **Phase 4 (Service)**: 1 hour
- **Phase 5 (Analytics)**: 3 hours
- **Phase 6 (API)**: 1 hour
- **Phase 7 (Frontend)**: 3 hours
- **Testing & Verification**: 2-3 hours

**Total**: ~14-15 hours of development + testing

---

## Future Enhancements

1. **Source Recommendations**:
   - Identify underutilized sources
   - Suggest sources to add based on unanswered queries

2. **Source Quality Metrics**:
   - Track which sources lead to positive feedback
   - Automatic source pruning based on quality

3. **Cost Attribution**:
   - Track token usage per source
   - ROI analysis for source ingestion costs
