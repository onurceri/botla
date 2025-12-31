package processing

import (
	"context"
)

type QueueStatus struct {
	WorkerCount int
	QueueLength int
	IsRunning   bool
}

type ProcessingSubsystem interface {
	EnqueueSource(ctx context.Context, sourceID, chatbotID string) (jobID string, err error)
	EnqueueJob(jobID string)
	Status() QueueStatus
	Stop()
}

var _ ProcessingSubsystem = (*SourceQueue)(nil)

func (sq *SourceQueue) Status() QueueStatus {
	return QueueStatus{
		WorkerCount: sq.WorkerCount(),
		QueueLength: sq.QueueLength(),
		IsRunning:   sq.queue != nil,
	}
}

func (sq *SourceQueue) EnqueueJob(jobID string) {
	sq.Enqueue(jobID)
}
