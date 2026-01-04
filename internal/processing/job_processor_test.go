package processing

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/logger"
)

// MockSourceProcessor is a test double for SourceProcessor
type MockSourceProcessor struct {
	ProcessFunc func(context.Context, string, *models.DataSource, *models.Chatbot, string, *models.Plan, *models.TrainingStep, StepCallback) ProcessResult
}

func (m *MockSourceProcessor) ProcessWithSteps(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
	if m.ProcessFunc != nil {
		return m.ProcessFunc(ctx, jobID, source, bot, langCode, plan, lastStep, onStep)
	}
	return ProcessResult{}
}

// MockTrainingJobRepo is a minimal mock for TrainingJobRepository
type MockTrainingJobRepo struct{}

func (m *MockTrainingJobRepo) GetByID(ctx context.Context, id string) (*models.TrainingJob, error) {
	return nil, nil
}

func (m *MockTrainingJobRepo) GetBySourceID(ctx context.Context, sourceID string) (*models.TrainingJob, error) {
	return nil, nil
}

func (m *MockTrainingJobRepo) GetByChatbotID(ctx context.Context, chatbotID string, limit int) ([]*models.TrainingJob, error) {
	return nil, nil
}

func (m *MockTrainingJobRepo) Create(ctx context.Context, sourceID, chatbotID string) (*models.TrainingJob, error) {
	return &models.TrainingJob{ID: "mock-job-id"}, nil
}

func (m *MockTrainingJobRepo) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, step *models.TrainingStep) error {
	return nil
}

func (m *MockTrainingJobRepo) ResetForRetry(ctx context.Context, id string) error {
	return nil
}

func (m *MockTrainingJobRepo) IncrementRetryCount(ctx context.Context, id string) (int, error) {
	return 0, nil
}

func (m *MockTrainingJobRepo) GetPendingJobs(ctx context.Context, limit int) ([]*models.TrainingJob, error) {
	return nil, nil
}

func (m *MockTrainingJobRepo) MarkStepCompleted(ctx context.Context, jobID string, step models.TrainingStep, outputHash string) error {
	return nil
}

func (m *MockTrainingJobRepo) GetLastCompletedStep(ctx context.Context, jobID string) (*models.TrainingStep, error) {
	return nil, nil
}

func (m *MockTrainingJobRepo) Fail(ctx context.Context, id string, step models.TrainingStep, errCode, errMsg string) error {
	return nil
}

func (m *MockTrainingJobRepo) Complete(ctx context.Context, id string) error {
	return nil
}

func (m *MockTrainingJobRepo) Cancel(ctx context.Context, id string) error {
	return nil
}

func (m *MockTrainingJobRepo) GetRetryableJobs(ctx context.Context, maxRetries, limit int) ([]*models.TrainingJob, error) {
	return nil, nil
}

func (m *MockTrainingJobRepo) GetRunningJobs(ctx context.Context, limit int) ([]*models.TrainingJob, error) {
	return nil, nil
}

func TestNewJobProcessor_WithProcessorMap(t *testing.T) {
	t.Parallel()
	log := logger.New("TEST")

	sourceRepo := repository.NewMockSourceRepo()
	trainingJobRepo := &MockTrainingJobRepo{}
	chatbotRepo := repository.NewMockChatbotRepo()
	planRepo := repository.NewMockPlanRepo()
	usageRepo := &MockUsageRepo{}

	cfg := JobProcessorConfig{
		TrainingJobRepo: trainingJobRepo,
		SourceRepo:      sourceRepo,
		ChatbotRepo:     chatbotRepo,
		PlanRepo:        planRepo,
		UsageRepo:       usageRepo,
		Log:             log,
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

	log := logger.New("TEST")

	called := false
	mockProcessor := &MockSourceProcessor{
		ProcessFunc: func(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
			called = true
			return ProcessResult{ChunkCount: 10}
		},
	}

	sourceRepo := repository.NewMockSourceRepo()
	trainingJobRepo := &MockTrainingJobRepo{}
	chatbotRepo := repository.NewMockChatbotRepo()
	planRepo := repository.NewMockPlanRepo()

	jp := &JobProcessor{
		trainingJobRepo: trainingJobRepo,
		sourceRepo:      sourceRepo,
		chatbotRepo:     chatbotRepo,
		planRepo:        planRepo,
		log:             log,
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

	log := logger.New("TEST")

	called := false
	mockProcessor := &MockSourceProcessor{
		ProcessFunc: func(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
			called = true
			return ProcessResult{ChunkCount: 5}
		},
	}

	sourceRepo := repository.NewMockSourceRepo()
	trainingJobRepo := &MockTrainingJobRepo{}
	chatbotRepo := repository.NewMockChatbotRepo()
	planRepo := repository.NewMockPlanRepo()

	jp := &JobProcessor{
		trainingJobRepo: trainingJobRepo,
		sourceRepo:      sourceRepo,
		chatbotRepo:     chatbotRepo,
		planRepo:        planRepo,
		log:             log,
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

	log := logger.New("TEST")

	called := false
	mockProcessor := &MockSourceProcessor{
		ProcessFunc: func(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult {
			called = true
			return ProcessResult{ChunkCount: 3}
		},
	}

	sourceRepo := repository.NewMockSourceRepo()
	trainingJobRepo := &MockTrainingJobRepo{}
	chatbotRepo := repository.NewMockChatbotRepo()
	planRepo := repository.NewMockPlanRepo()

	jp := &JobProcessor{
		trainingJobRepo: trainingJobRepo,
		sourceRepo:      sourceRepo,
		chatbotRepo:     chatbotRepo,
		planRepo:        planRepo,
		log:             log,
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
