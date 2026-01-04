package processing

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
)

type SourceProcessor interface {
	ProcessWithSteps(ctx context.Context, jobID string, source *models.DataSource, bot *models.Chatbot, langCode string, plan *models.Plan, lastStep *models.TrainingStep, onStep StepCallback) ProcessResult
}
