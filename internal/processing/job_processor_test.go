package processing

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/logger"
)

func TestNewJobProcessor_WithProcessorMap(t *testing.T) {
	t.Parallel()
	tdb := testdb.OpenParallelTestDB(t)
	log := logger.New("TEST")

	cfg := JobProcessorConfig{
		DB:  tdb,
		Log: log,
	}

	jp := NewJobProcessor(cfg)

	if jp == nil {
		t.Fatal("expected non-nil JobProcessor")
	}

	if jp.processors == nil {
		t.Error("expected processors map to be initialized")
	}

	t.Run("URL processor is registered", func(t *testing.T) {
		proc, ok := jp.processors["url"]
		if !ok {
			t.Error("url processor not found in map")
		}
		if proc == nil {
			t.Error("url processor should not be nil")
		}
	})

	t.Run("PDF processor is registered", func(t *testing.T) {
		proc, ok := jp.processors["pdf"]
		if !ok {
			t.Error("pdf processor not found in map")
		}
		if proc == nil {
			t.Error("pdf processor should not be nil")
		}
	})

	t.Run("Text processor is registered", func(t *testing.T) {
		proc, ok := jp.processors["text"]
		if !ok {
			t.Error("text processor not found in map")
		}
		if proc == nil {
			t.Error("text processor should not be nil")
		}
	})
}

func TestProcessWithResume_URLType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)

	log := logger.New("TEST")

	called := false
	mockProcessor := &MockSourceProcessor{
		ProcessFunc: func(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
			called = true
			return ProcessResult{ChunkCount: 10}
		},
	}

	jp := &JobProcessor{
		db:  tdb,
		log: log,
		processors: map[string]SourceProcessor{
			"url": mockProcessor,
		},
	}

	sourceURL := "https://example.com"
	source := &models.DataSource{
		ID:         "test-source-1",
		ChatbotID:  "test-bot-1",
		SourceType: "url",
		SourceURL:  &sourceURL,
	}
	bot := &models.Chatbot{
		ID:     "test-bot-1",
		UserID: "test-user-1",
	}
	plan := &models.Plan{
		Limits: &models.PlanLimits{},
	}

	result := jp.processWithResume(ctx, "test-job-1", source, bot, "en", plan, nil)

	if !called {
		t.Error("mock processor was not called")
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if result.ChunkCount != 10 {
		t.Errorf("expected chunk count 10, got %d", result.ChunkCount)
	}
}

func TestProcessWithResume_PDFType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)

	log := logger.New("TEST")

	called := false
	mockProcessor := &MockSourceProcessor{
		ProcessFunc: func(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
			called = true
			return ProcessResult{ChunkCount: 5}
		},
	}

	jp := &JobProcessor{
		db:  tdb,
		log: log,
		processors: map[string]SourceProcessor{
			"pdf": mockProcessor,
		},
	}

	filePath := "test.pdf"
	source := &models.DataSource{
		ID:         "test-source-2",
		ChatbotID:  "test-bot-2",
		SourceType: "pdf",
		FilePath:   &filePath,
	}
	bot := &models.Chatbot{
		ID:     "test-bot-2",
		UserID: "test-user-2",
	}
	plan := &models.Plan{
		Limits: &models.PlanLimits{},
	}

	result := jp.processWithResume(ctx, "test-job-2", source, bot, "en", plan, nil)

	if !called {
		t.Error("mock processor was not called")
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if result.ChunkCount != 5 {
		t.Errorf("expected chunk count 5, got %d", result.ChunkCount)
	}
}

func TestProcessWithResume_TextType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)

	log := logger.New("TEST")

	called := false
	mockProcessor := &MockSourceProcessor{
		ProcessFunc: func(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
			called = true
			return ProcessResult{ChunkCount: 3}
		},
	}

	jp := &JobProcessor{
		db:  tdb,
		log: log,
		processors: map[string]SourceProcessor{
			"text": mockProcessor,
		},
	}

	filePath := "test.txt"
	source := &models.DataSource{
		ID:         "test-source-3",
		ChatbotID:  "test-bot-3",
		SourceType: "text",
		FilePath:   &filePath,
	}
	bot := &models.Chatbot{
		ID:     "test-bot-3",
		UserID: "test-user-3",
	}
	plan := &models.Plan{
		Limits: &models.PlanLimits{},
	}

	result := jp.processWithResume(ctx, "test-job-3", source, bot, "en", plan, nil)

	if !called {
		t.Error("mock processor was not called")
	}
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if result.ChunkCount != 3 {
		t.Errorf("expected chunk count 3, got %d", result.ChunkCount)
	}
}

func TestProcessWithResume_UnknownType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)

	cfg := JobProcessorConfig{
		DB: tdb,
	}

	jp := NewJobProcessor(cfg)

	source := &models.DataSource{
		ID:         "test-source-unknown",
		ChatbotID:  "test-bot-unknown",
		SourceType: "unknown_type",
	}
	bot := &models.Chatbot{
		ID: "test-bot-unknown",
	}
	plan := &models.Plan{
		Limits: &models.PlanLimits{},
	}

	result := jp.processWithResume(ctx, "test-job-unknown", source, bot, "en", plan, nil)

	if result.Error == nil {
		t.Error("expected error for unknown source type")
	}

	if result.Error.Error() != ErrCodeUnknownSourceType {
		t.Errorf("expected error %s, got %s", ErrCodeUnknownSourceType, result.Error.Error())
	}

	if result.FailedStep != models.StepFetchSource {
		t.Errorf("expected failed step %s, got %s", models.StepFetchSource, result.FailedStep)
	}
}

func TestProcessWithResume_NilProcessor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)
	log := logger.New("TEST")

	jp := &JobProcessor{
		db:  tdb,
		log: log,
		processors: map[string]SourceProcessor{
			"test_type": nil,
		},
	}

	source := &models.DataSource{
		ID:         "test-source-nil",
		ChatbotID:  "test-bot-nil",
		SourceType: "test_type",
	}
	bot := &models.Chatbot{
		ID: "test-bot-nil",
	}
	plan := &models.Plan{
		Limits: &models.PlanLimits{},
	}

	result := jp.processWithResume(ctx, "test-job-nil", source, bot, "en", plan, nil)

	if result.Error == nil {
		t.Error("expected error for nil processor")
	}

	if result.Error.Error() != ErrCodeUnknownSourceType {
		t.Errorf("expected error %s, got %s", ErrCodeUnknownSourceType, result.Error.Error())
	}

	if result.FailedStep != models.StepFetchSource {
		t.Errorf("expected failed step %s, got %s", models.StepFetchSource, result.FailedStep)
	}
}

func TestProcessWithResume_EmptyMap(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tdb := testdb.OpenParallelTestDB(t)
	log := logger.New("TEST")

	jp := &JobProcessor{
		db:         tdb,
		log:        log,
		processors: make(map[string]SourceProcessor),
	}

	source := &models.DataSource{
		ID:         "test-source-empty",
		ChatbotID:  "test-bot-empty",
		SourceType: "url",
	}
	bot := &models.Chatbot{
		ID: "test-bot-empty",
	}
	plan := &models.Plan{
		Limits: &models.PlanLimits{},
	}

	result := jp.processWithResume(ctx, "test-job-empty", source, bot, "en", plan, nil)

	if result.Error == nil {
		t.Error("expected error for empty processor map")
	}

	if result.Error.Error() != ErrCodeUnknownSourceType {
		t.Errorf("expected error %s, got %s", ErrCodeUnknownSourceType, result.Error.Error())
	}

	if result.FailedStep != models.StepFetchSource {
		t.Errorf("expected failed step %s, got %s", models.StepFetchSource, result.FailedStep)
	}
}

type MockSourceProcessor struct {
	ProcessFunc func(context.Context, string, *models.DataSource, *models.Chatbot, string, *models.Plan, *models.TrainingStep, StepCallback) ProcessResult
}

func (m *MockSourceProcessor) ProcessWithSteps(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
	if m.ProcessFunc != nil {
		return m.ProcessFunc(ctx, jobID, source, bot, langCode, plan, lastStep, onStep)
	}
	return ProcessResult{}
}
