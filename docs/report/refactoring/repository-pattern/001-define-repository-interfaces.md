# Task 001: Define Repository Interfaces

## Background

The first step in adopting the Repository Pattern is to define interfaces that abstract data access. These interfaces will be implemented by concrete repositories and can be mocked for testing.

**Target Location:** `internal/repository/interfaces.go`

## Implementation Plan

1. **Create repository package**
   - Create `internal/repository/` directory
   - Create `interfaces.go` with repository interfaces

2. **Define ChatbotRepository interface**
   - `GetByID(ctx, id) (*models.Chatbot, error)`
   - `GetByUserID(ctx, userID) ([]models.Chatbot, error)`
   - `GetByWorkspace(ctx, workspaceID) ([]models.Chatbot, error)`
   - `Create(ctx, bot) (string, error)`
   - `Update(ctx, bot) error`
   - `SoftDelete(ctx, id, userID) ([]string, error)`
   - `CountByUserID(ctx, userID) (int, error)`
   - `CountByWorkspace(ctx, workspaceID) (int, error)`

3. **Define ActionRepository interface**
   - `List(ctx, chatbotID) ([]*models.ChatbotAction, error)`
   - `GetByID(ctx, id) (*models.ChatbotAction, error)`
   - `Create(ctx, action) error`
   - `Update(ctx, action) error`
   - `Delete(ctx, id) error`
   - `GetLogs(ctx, chatbotID, limit, offset) ([]*models.ActionExecutionLog, error)`

4. **Define SourceRepository interface**
   - `GetByID(ctx, id) (*models.DataSource, error)`
   - `GetByChatbot(ctx, chatbotID) ([]models.DataSource, error)`

## Checklist

- [x] Create `internal/repository/` directory
- [x] Create `internal/repository/interfaces.go`
- [x] Define `ChatbotRepository` interface
- [x] Define `ActionRepository` interface
- [x] Define `SourceRepository` interface
- [x] Run `go build ./...` to verify compilation
