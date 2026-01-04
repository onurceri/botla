package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSourceRepo(t *testing.T) repository.SourceRepository {
	db := testdb.OpenParallelTestDB(t)
	return repository.NewPostgresSourceRepo(db)
}

// TestPostgresSourceRepo_GetByID_NotFound tests retrieving a non-existent source.
func TestPostgresSourceRepo_GetByID_NotFound(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	source, err := repo.GetByID(ctx, uuid.NewString())

	require.NoError(t, err)
	assert.Nil(t, source, "should return nil for non-existent source")
}

// TestPostgresSourceRepo_GetByID_Success tests retrieving an existing source.
func TestPostgresSourceRepo_GetByID_Success(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())
	source, err := repo.GetByID(ctx, sourceResult.Source.ID)

	require.NoError(t, err)
	require.NotNil(t, source, "source should be found")
	assert.Equal(t, sourceResult.Source.ID, source.ID)
	assert.Equal(t, sourceResult.Chatbot.ID, source.ChatbotID)
	assert.Equal(t, sourceResult.Source.SourceType, source.SourceType)
}

// TestPostgresSourceRepo_GetByChatbot_Empty tests listing sources for a chatbot with no sources.
func TestPostgresSourceRepo_GetByChatbot_Empty(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	chatbotResult := testdb.CreateChatbot(t, repo.(*repository.PostgresSourceRepo).Pool())

	sources, err := repo.GetByChatbot(ctx, chatbotResult.Chatbot.ID)

	require.NoError(t, err)
	assert.Empty(t, sources, "should return empty list for chatbot with no sources")
}

// TestPostgresSourceRepo_GetByChatbot_Success tests listing sources for a chatbot.
func TestPostgresSourceRepo_GetByChatbot_Success(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())

	sources, err := repo.GetByChatbot(ctx, sourceResult.Chatbot.ID)

	require.NoError(t, err)
	require.Len(t, sources, 1, "should find 1 source")
	assert.Equal(t, sourceResult.Source.ID, sources[0].ID)
}

// TestPostgresSourceRepo_GetByChatbot_ExcludesDeleted tests that deleted sources are excluded.
func TestPostgresSourceRepo_GetByChatbot_ExcludesDeleted(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())
	sourceResult2 := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{ChatbotID: sourceResult.Chatbot.ID})

	// Soft delete the first source
	err := repo.SoftDelete(ctx, sourceResult.Source.ID)
	require.NoError(t, err)

	sources, err := repo.GetByChatbot(ctx, sourceResult.Chatbot.ID)

	require.NoError(t, err)
	assert.Len(t, sources, 1, "should only return non-deleted sources")
	assert.Equal(t, sourceResult2.Source.ID, sources[0].ID, "should return the non-deleted source (second one, ordered by created_at DESC)")
}

// TestPostgresSourceRepo_GetURLSources_Empty tests listing URL sources for a chatbot with none.
func TestPostgresSourceRepo_GetURLSources_Empty(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	chatbotResult := testdb.CreateChatbot(t, repo.(*repository.PostgresSourceRepo).Pool())
	_ = testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		ChatbotID:  chatbotResult.Chatbot.ID,
		SourceType: "text",
	})

	sources, err := repo.GetURLSources(ctx, chatbotResult.Chatbot.ID)

	require.NoError(t, err)
	assert.Empty(t, sources, "should return empty list when no URL sources exist")
}

// TestPostgresSourceRepo_GetURLSources_Success tests listing URL sources for a chatbot.
func TestPostgresSourceRepo_GetURLSources_Success(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		SourceType: "url",
		SourceURL:  stringPtr("https://example.com/page1"),
	})
	_ = testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		ChatbotID:  sourceResult.Chatbot.ID,
		SourceType: "url",
		SourceURL:  stringPtr("https://example.com/page2"),
	})

	sources, err := repo.GetURLSources(ctx, sourceResult.Chatbot.ID)

	require.NoError(t, err)
	require.Len(t, sources, 2, "should find 2 URL sources")
}

// TestPostgresSourceRepo_GetURLSources_ExcludesDeleted tests that deleted URL sources are excluded.
func TestPostgresSourceRepo_GetURLSources_ExcludesDeleted(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		SourceType: "url",
		SourceURL:  stringPtr("https://example.com/page1"),
	})

	// Soft delete the source
	err := repo.SoftDelete(ctx, sourceResult.Source.ID)
	require.NoError(t, err)

	sources, err := repo.GetURLSources(ctx, sourceResult.Chatbot.ID)

	require.NoError(t, err)
	assert.Empty(t, sources, "should exclude deleted URL sources")
}

// TestPostgresSourceRepo_SoftDelete_Success tests soft deleting a source.
func TestPostgresSourceRepo_SoftDelete_Success(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())

	err := repo.SoftDelete(ctx, sourceResult.Source.ID)

	require.NoError(t, err)

	// Verify source is soft deleted (not returned by GetByChatbot)
	sources, err := repo.GetByChatbot(ctx, sourceResult.Chatbot.ID)
	require.NoError(t, err)
	assert.Empty(t, sources, "soft deleted source should not appear in GetByChatbot")

	// But it should still be retrievable by ID
	retrieved, err := repo.GetByID(ctx, sourceResult.Source.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.NotNil(t, retrieved.DeletedAt, "source should have deleted_at set")
}

// TestPostgresSourceRepo_Delete_Success tests permanently deleting a source.
func TestPostgresSourceRepo_Delete_Success(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())

	err := repo.Delete(ctx, sourceResult.Source.ID)

	require.NoError(t, err)

	// Verify source is deleted (not retrievable by ID)
	retrieved, err := repo.GetByID(ctx, sourceResult.Source.ID)
	require.NoError(t, err)
	assert.Nil(t, retrieved, "deleted source should not be retrievable")
}

// TestPostgresSourceRepo_Exists_True tests checking if a source exists.
func TestPostgresSourceRepo_Exists_True(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		SourceURL: stringPtr("https://existing-url.com"),
	})

	exists, err := repo.Exists(ctx, sourceResult.Chatbot.ID, "https://existing-url.com")

	require.NoError(t, err)
	assert.True(t, exists, "should return true for existing source")
}

// TestPostgresSourceRepo_Exists_False tests checking if a source does not exist.
func TestPostgresSourceRepo_Exists_False(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateChatbot(t, repo.(*repository.PostgresSourceRepo).Pool())

	exists, err := repo.Exists(ctx, sourceResult.Chatbot.ID, "https://non-existent.com")

	require.NoError(t, err)
	assert.False(t, exists, "should return false for non-existent source")
}

// TestPostgresSourceRepo_ExistsByHash tests ExistsByHash functionality.
func TestPostgresSourceRepo_ExistsByHash(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())

	// Non-existent hash should return false
	exists, err := repo.ExistsByHash(ctx, sourceResult.Chatbot.ID, "non-existent-hash")
	require.NoError(t, err)
	assert.False(t, exists, "should return false for non-existent hash")
}

// TestPostgresSourceRepo_GetByHash_NotFound tests retrieving a source by non-existent hash.
func TestPostgresSourceRepo_GetByHash_NotFound(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateChatbot(t, repo.(*repository.PostgresSourceRepo).Pool())

	source, err := repo.GetByHash(ctx, sourceResult.Chatbot.ID, "non-existent-hash")

	require.NoError(t, err)
	assert.Nil(t, source, "should return nil for non-existent hash")
}

// TestPostgresSourceRepo_GetByHash_ExcludesDeleted tests that deleted sources are excluded.
func TestPostgresSourceRepo_GetByHash_ExcludesDeleted(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())

	// Soft delete the source
	err := repo.SoftDelete(ctx, sourceResult.Source.ID)
	require.NoError(t, err)

	// GetByHash should not find it
	source, err := repo.GetByHash(ctx, sourceResult.Chatbot.ID, "any-hash")

	require.NoError(t, err)
	assert.Nil(t, source, "should not return soft-deleted source")
}

// TestPostgresSourceRepo_CountByType_Success tests counting sources by type.
func TestPostgresSourceRepo_CountByType_Success(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		SourceType: "url",
		Status:     "completed",
	})
	_ = testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		ChatbotID:  sourceResult.Chatbot.ID,
		SourceType: "url",
		Status:     "completed",
	})
	_ = testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		ChatbotID:  sourceResult.Chatbot.ID,
		SourceType: "text",
		Status:     "completed",
	})

	count, err := repo.CountByType(ctx, sourceResult.Chatbot.ID, "url")

	require.NoError(t, err)
	assert.Equal(t, 2, count, "should count 2 URL sources")
}

// TestPostgresSourceRepo_CountByType_ExcludesFailed tests that failed sources are excluded.
func TestPostgresSourceRepo_CountByType_ExcludesFailed(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		SourceType: "url",
		Status:     "completed",
	})
	_ = testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		ChatbotID:  sourceResult.Chatbot.ID,
		SourceType: "url",
		Status:     "failed",
	})

	count, err := repo.CountByType(ctx, sourceResult.Chatbot.ID, "url")

	require.NoError(t, err)
	assert.Equal(t, 1, count, "should not count failed sources")
}

// TestPostgresSourceRepo_CountByType_ExcludesDeleted tests that deleted sources are excluded.
func TestPostgresSourceRepo_CountByType_ExcludesDeleted(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		SourceType: "url",
		Status:     "completed",
	})
	_ = testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		ChatbotID:  sourceResult.Chatbot.ID,
		SourceType: "url",
		Status:     "completed",
	})

	// Soft delete one source
	err := repo.SoftDelete(ctx, sourceResult.Source.ID)
	require.NoError(t, err)

	count, err := repo.CountByType(ctx, sourceResult.Chatbot.ID, "url")

	require.NoError(t, err)
	assert.Equal(t, 1, count, "should not count deleted sources")
}

// TestPostgresSourceRepo_CountByType_Empty tests counting sources when none exist.
func TestPostgresSourceRepo_CountByType_Empty(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	chatbotResult := testdb.CreateChatbot(t, repo.(*repository.PostgresSourceRepo).Pool())

	count, err := repo.CountByType(ctx, chatbotResult.Chatbot.ID, "url")

	require.NoError(t, err)
	assert.Equal(t, 0, count, "should return 0 for chatbot with no sources")
}

// TestPostgresSourceRepo_GetByChatbot_MultipleChatbots tests that sources from other chatbots are not returned.
func TestPostgresSourceRepo_GetByChatbot_MultipleChatbots(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult1 := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())
	sourceResult2 := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		ChatbotID: sourceResult1.Chatbot.ID,
	})

	sources, err := repo.GetByChatbot(ctx, sourceResult1.Chatbot.ID)

	require.NoError(t, err)
	require.Len(t, sources, 2, "should find 2 sources for chatbot 1")
	_ = sourceResult2
}

// TestPostgresSourceRepo_GetByChatbot_Ordering tests that sources are ordered by created_at DESC.
func TestPostgresSourceRepo_GetByChatbot_Ordering(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	sourceResult := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool())

	// Wait a moment to ensure time difference
	time.Sleep(10 * time.Millisecond)

	sourceResult2 := testdb.CreateSource(t, repo.(*repository.PostgresSourceRepo).Pool(), testdb.SourceFixture{
		ChatbotID: sourceResult.Chatbot.ID,
	})

	sources, err := repo.GetByChatbot(ctx, sourceResult.Chatbot.ID)

	require.NoError(t, err)
	require.Len(t, sources, 2)
	assert.Equal(t, sourceResult2.Source.ID, sources[0].ID, "newer source should be first")
	assert.Equal(t, sourceResult.Source.ID, sources[1].ID, "older source should be second")
}

// TestPostgresSourceRepo_Create_WithAllFields tests creating a source with all fields populated.
func TestPostgresSourceRepo_Create_WithAllFields(t *testing.T) {
	repo := newSourceRepo(t)
	ctx := context.Background()

	chatbotResult := testdb.CreateChatbot(t, repo.(*repository.PostgresSourceRepo).Pool())
	sourceURL := stringPtr("https://full-test.com/file.pdf")
	filePath := stringPtr("/uploads/file.pdf")
	originalFilename := stringPtr("My Document.pdf")
	hash := stringPtr("sha256-abc123")
	capabilitySummary := stringPtr("Contains information about X, Y, Z")

	source := &models.DataSource{
		ChatbotID:         chatbotResult.Chatbot.ID,
		SourceType:        "file",
		SourceURL:         sourceURL,
		FilePath:          filePath,
		OriginalFilename:  originalFilename,
		Status:            "pending",
		ErrorMessage:      nil,
		ChunkCount:        0,
		ProcessedAt:       nil,
		Hash:              hash,
		DeletedAt:         nil,
		SizeBytes:         2048,
		LastRefreshedAt:   nil,
		IsDiscovered:      false,
		CapabilitySummary: capabilitySummary,
	}

	id, err := repo.Create(ctx, source)

	require.NoError(t, err)
	require.NotEmpty(t, id)

	retrieved, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, id, retrieved.ID)
	assert.Equal(t, source.ChatbotID, retrieved.ChatbotID)
	assert.Equal(t, source.SourceType, retrieved.SourceType)
	assert.Equal(t, *source.SourceURL, *retrieved.SourceURL)
	assert.Equal(t, *source.FilePath, *retrieved.FilePath)
	assert.Equal(t, *source.OriginalFilename, *retrieved.OriginalFilename)
	assert.Equal(t, source.Status, retrieved.Status)
	assert.Equal(t, source.SizeBytes, retrieved.SizeBytes)
	assert.Equal(t, *source.Hash, *retrieved.Hash)
	assert.Equal(t, *source.CapabilitySummary, *retrieved.CapabilitySummary)
}

func stringPtr(s string) *string {
	return &s
}
