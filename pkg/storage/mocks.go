package storage

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

// MockStorageService is a mock implementation of StorageService
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) UploadFile(ctx context.Context, key string, body io.Reader) (string, error) {
	args := m.Called(ctx, key, body)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockStorageService) DeleteFile(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}
