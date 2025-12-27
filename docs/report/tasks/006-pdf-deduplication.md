# Task 006: PDF Hash-Based Deduplication

**Priority:** 🟡 High (Cost Optimization)  
**Phase:** 3 - Idempotency & Deduplication  
**Estimated Time:** 2-3 hours  
**Dependencies:** None  

---

## Problem Statement

Currently, users can upload the same PDF file multiple times to the same chatbot:
- Wastes embedding API tokens (costs money)
- Wastes storage space
- Creates duplicate vectors in Qdrant
- User confusion over having duplicate sources

**Evidence:**
- `source_create.go` computes hash but doesn't check for duplicates
- URL sources have duplicate check (`SourceExists`), PDFs don't

---

## Objective

Implement hash-based deduplication for PDF and text sources:
1. Check if content hash already exists for the chatbot
2. Reject duplicates with clear error message
3. Allow same content across different chatbots (that's valid use case)

---

## Implementation Details

### Step 1: Add Database Function

**File:** `internal/db/source.go` (MODIFY)

```go
// SourceExistsByHash checks if a source with the same hash exists for a chatbot
func SourceExistsByHash(ctx context.Context, db *sql.DB, chatbotID, hash string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM data_sources 
			WHERE chatbot_id = $1 
			  AND hash = $2 
			  AND deleted_at IS NULL
		)
	`, chatbotID, hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check source hash: %w", err)
	}
	return exists, nil
}

// GetSourceByHash returns the existing source with the same hash
func GetSourceByHash(ctx context.Context, db *sql.DB, chatbotID, hash string) (*models.DataSource, error) {
	var s models.DataSource
	err := db.QueryRowContext(ctx, `
		SELECT id, chatbot_id, source_type, status, created_at
		FROM data_sources 
		WHERE chatbot_id = $1 AND hash = $2 AND deleted_at IS NULL
		LIMIT 1
	`, chatbotID, hash).Scan(&s.ID, &s.ChatbotID, &s.SourceType, &s.Status, &s.CreatedAt)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get source by hash: %w", err)
	}
	return &s, nil
}
```

### Step 2: Add Error Code

**File:** `internal/api/error_codes.go` (MODIFY)

```go
const (
	// ... existing codes
	ErrDuplicateContent = "ERR_DUPLICATE_CONTENT" // New
)
```

### Step 3: Update PDF Handler

**File:** `internal/api/handlers/source_create.go` (MODIFY)

In `handlePDFUpload`, add duplicate check after computing hash:

```go
func (h *SourcesHandlers) handlePDFUpload(w http.ResponseWriter, r *http.Request, bot *models.Chatbot, plan *models.Plan, cfg langconfig.LanguageConfig) {
	// ... existing validation code ...

	// Read file into memory to compute hash
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hval := computeHash(buf.Bytes())

	// NEW: Check for duplicate content
	exists, err := db.SourceExistsByHash(r.Context(), h.DB, bot.ID, hval)
	if err != nil {
		h.logError("hash_check_failed", map[string]any{"error": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		api.WriteLocalizedError(w, http.StatusConflict, api.ErrDuplicateContent, cfg)
		return
	}

	// ... rest of upload code ...
}
```

### Step 4: Update Text Handler

**File:** `internal/api/handlers/source_create.go` (MODIFY)

In `handleTextSource`, add same duplicate check:

```go
func (h *SourcesHandlers) handleTextSource(w http.ResponseWriter, r *http.Request, bot *models.Chatbot, userID string, plan *models.Plan, cfg langconfig.LanguageConfig) {
	// ... existing validation ...

	hval := computeHash([]byte(text))

	// NEW: Check for duplicate content
	exists, err := db.SourceExistsByHash(r.Context(), h.DB, bot.ID, hval)
	if err != nil {
		h.logError("hash_check_failed", map[string]any{"error": err.Error()})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		api.WriteLocalizedError(w, http.StatusConflict, api.ErrDuplicateContent, cfg)
		return
	}

	// ... rest of code ...
}
```

### Step 5: Add Localized Error Messages

**File:** `internal/api/errors_localized.go` (MODIFY)

```go
var errorMessages = map[string]map[string]string{
	"ERR_DUPLICATE_CONTENT": {
		"en": "This content has already been added to this chatbot",
		"tr": "Bu içerik zaten bu chatbota eklenmiş",
	},
	// ... existing messages
}
```

### Step 6: Add Frontend Error Message

**File:** `frontend/src/i18n/errors.ts` (MODIFY)

```typescript
// Add to errorMessages
ERR_DUPLICATE_CONTENT: 'This content has already been added',

// Turkish
ERR_DUPLICATE_CONTENT: 'Bu içerik zaten eklenmiş',
```

---

## Tests to Write

### Unit Tests

**File:** `internal/db/source_dedup_test.go` (NEW)

```go
package db

import (
	"context"
	"testing"
)

func TestSourceExistsByHash(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	chatbotID := createTestChatbot(t, db)
	
	// Create a source with hash
	hash := "abc123hash"
	createSourceWithHash(t, db, chatbotID, hash)

	// Test: Same chatbot, same hash should exist
	exists, err := SourceExistsByHash(context.Background(), db, chatbotID, hash)
	if err != nil {
		t.Fatalf("SourceExistsByHash failed: %v", err)
	}
	if !exists {
		t.Error("expected source to exist")
	}

	// Test: Same chatbot, different hash should not exist
	exists, err = SourceExistsByHash(context.Background(), db, chatbotID, "different-hash")
	if err != nil {
		t.Fatalf("SourceExistsByHash failed: %v", err)
	}
	if exists {
		t.Error("expected source to not exist")
	}

	// Test: Different chatbot, same hash should not exist (ok to have same content in different bots)
	otherChatbotID := createTestChatbot(t, db)
	exists, err = SourceExistsByHash(context.Background(), db, otherChatbotID, hash)
	if err != nil {
		t.Fatalf("SourceExistsByHash failed: %v", err)
	}
	if exists {
		t.Error("expected source to not exist in different chatbot")
	}
}

func TestSourceExistsByHash_DeletedSource(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	chatbotID := createTestChatbot(t, db)
	hash := "deleted-hash"
	
	// Create and delete a source
	sourceID := createSourceWithHash(t, db, chatbotID, hash)
	softDeleteSource(t, db, sourceID)

	// Should NOT find deleted source
	exists, err := SourceExistsByHash(context.Background(), db, chatbotID, hash)
	if err != nil {
		t.Fatalf("SourceExistsByHash failed: %v", err)
	}
	if exists {
		t.Error("should not find deleted source")
	}
}
```

### Integration Test

**File:** `internal/integration/dedup_test.go` (NEW)

```go
package integration

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
)

func TestPDFDeduplication_Integration(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	token := authToken(t, te.Server.URL, "dedup@example.com")
	chatbotID := createChatbot(t, te.Server.URL, token, "Dedup Bot")

	// Create a simple PDF
	pdfContent := []byte("%PDF-1.4 dummy content for test")

	// Upload first time - should succeed
	resp1 := uploadPDF(t, te.Server.URL, token, chatbotID, pdfContent, "test.pdf")
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first upload failed: %d", resp1.StatusCode)
	}

	// Upload same content again - should fail with 409
	resp2 := uploadPDF(t, te.Server.URL, token, chatbotID, pdfContent, "test2.pdf")
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("second upload should be 409 Conflict, got %d", resp2.StatusCode)
	}

	// Verify error code
	body, _ := io.ReadAll(resp2.Body)
	if !bytes.Contains(body, []byte("ERR_DUPLICATE_CONTENT")) {
		t.Errorf("expected ERR_DUPLICATE_CONTENT error, got %s", string(body))
	}
}

func TestTextDeduplication_Integration(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	token := authToken(t, te.Server.URL, "textdedup@example.com")
	chatbotID := createChatbot(t, te.Server.URL, token, "Text Dedup Bot")

	textContent := "This is some unique text content for testing"

	// Upload first time
	resp1 := uploadText(t, te.Server.URL, token, chatbotID, textContent)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("first upload failed: %d", resp1.StatusCode)
	}

	// Upload same text again
	resp2 := uploadText(t, te.Server.URL, token, chatbotID, textContent)
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("second upload should be 409, got %d", resp2.StatusCode)
	}
}

func TestDeduplication_DifferentChatbots_Allowed(t *testing.T) {
	te := SetupTestEnv(t)
	defer te.Teardown()

	token := authToken(t, te.Server.URL, "crossbot@example.com")
	chatbot1 := createChatbot(t, te.Server.URL, token, "Bot 1")
	chatbot2 := createChatbot(t, te.Server.URL, token, "Bot 2")

	textContent := "Same content for different bots"

	// Upload to first chatbot
	resp1 := uploadText(t, te.Server.URL, token, chatbot1, textContent)
	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("upload to bot1 failed: %d", resp1.StatusCode)
	}

	// Upload same content to second chatbot - should SUCCEED
	resp2 := uploadText(t, te.Server.URL, token, chatbot2, textContent)
	if resp2.StatusCode != http.StatusCreated {
		t.Errorf("upload to bot2 should succeed, got %d", resp2.StatusCode)
	}
}

func uploadPDF(t *testing.T, baseURL, token, chatbotID string, content []byte, filename string) *http.Response {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(content)
	writer.WriteField("source_type", "pdf")
	writer.Close()

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots/"+chatbotID+"/sources", &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func uploadText(t *testing.T, baseURL, token, chatbotID, text string) *http.Response {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("source_type", "text")
	writer.WriteField("text", text)
	writer.Close()

	req, _ := http.NewRequest("POST", baseURL+"/api/v1/chatbots/"+chatbotID+"/sources", &buf)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
```

---

## Verification Steps

1. **Run unit tests:**
   ```bash
   go test ./internal/db/... -v -run TestSourceExistsByHash
   ```

2. **Run integration tests:**
   ```bash
   go test ./internal/integration/... -v -run TestDedup
   ```

3. **Manual verification:**
   ```bash
   # Upload a PDF
   curl -X POST http://localhost:8080/api/v1/chatbots/{id}/sources \
     -H "Authorization: Bearer $TOKEN" \
     -F "source_type=pdf" \
     -F "file=@test.pdf"
   # Should return 201

   # Upload same PDF again
   curl -X POST http://localhost:8080/api/v1/chatbots/{id}/sources \
     -H "Authorization: Bearer $TOKEN" \
     -F "source_type=pdf" \
     -F "file=@test.pdf"
   # Should return 409 with ERR_DUPLICATE_CONTENT
   ```

---

## Acceptance Criteria

- [x] Duplicate PDF upload returns 409 Conflict
- [x] Duplicate text upload returns 409 Conflict
- [x] Same content in different chatbots is allowed
- [x] Deleted sources don't block re-upload
- [x] Error message is user-friendly
- [x] All tests pass

---

## Files Changed

| File | Action |
|------|--------|
| `internal/db/source.go` | MODIFY |
| `internal/api/error_codes.go` | MODIFY |
| `internal/api/handlers/source_create.go` | MODIFY |
| `internal/api/errors_localized.go` | MODIFY |
| `frontend/src/i18n/errors.ts` | MODIFY |
| `internal/db/source_dedup_test.go` | CREATE |
| `internal/integration/dedup_test.go` | CREATE |

