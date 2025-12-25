package storage

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"
)

type MemoryStorage struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{data: make(map[string][]byte)}
}

func (m *MemoryStorage) UploadFile(ctx context.Context, key string, body io.Reader) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, _ := io.ReadAll(body)
	m.data[key] = b
	return key, nil
}

func (m *MemoryStorage) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.data[key]
	if !ok {
		return io.NopCloser(bytes.NewReader(nil)), nil
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (m *MemoryStorage) DeleteFile(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *MemoryStorage) GetSignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return "https://memory-storage.local/" + key, nil
}
