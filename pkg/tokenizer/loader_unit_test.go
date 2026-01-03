package tokenizer

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/storage"
)

// mockStorageService is a mock implementation of storage.StorageService for testing.
type mockStorageService struct {
	downloadFunc  func(ctx context.Context, key string) (io.ReadCloser, error)
	uploadFunc    func(ctx context.Context, key string, body io.Reader) (string, error)
	deleteFunc    func(ctx context.Context, key string) error
	signedURLFunc func(ctx context.Context, key string, expires time.Duration) (string, error)
}

func (m *mockStorageService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, key)
	}
	return nil, errors.New("not implemented")
}

func (m *mockStorageService) UploadFile(ctx context.Context, key string, body io.Reader) (string, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(ctx, key, body)
	}
	return "", errors.New("not implemented")
}

func (m *mockStorageService) DeleteFile(ctx context.Context, key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return errors.New("not implemented")
}

func (m *mockStorageService) GetSignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if m.signedURLFunc != nil {
		return m.signedURLFunc(ctx, key, expires)
	}
	return "", errors.New("not implemented")
}

func TestLoader_Get_EmptyCache(t *testing.T) {
	t.Parallel()

	mockStorage := &mockStorageService{}
	loader := NewLoader(mockStorage)

	tests := []struct {
		name     string
		langCode string
		wantData []byte
		wantOk   bool
	}{
		{
			name:     "returns nil when lang not in cache",
			langCode: "en",
			wantData: nil,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, ok := loader.Get(tt.langCode)
			if ok != tt.wantOk {
				t.Errorf("Get(%q) ok = %v, want %v", tt.langCode, ok, tt.wantOk)
			}
			if !bytes.Equal(data, tt.wantData) {
				t.Errorf("Get(%q) data = %v, want %v", tt.langCode, data, tt.wantData)
			}
		})
	}
}

func TestLoader_Get_WithCache(t *testing.T) {
	t.Parallel()

	expectedData := []byte(`{"vocab": ["hello", "world"]}`)

	// Create a mock storage that returns data
	mockStorage := &mockStorageService{
		downloadFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(expectedData)), nil
		},
	}

	// Create a loader and manually add to cache
	loader := NewLoader(mockStorage)
	loader.cache["en"] = expectedData

	// Test retrieval
	data, ok := loader.Get("en")
	if !ok {
		t.Error("Get(\"en\") = false, want true")
	}
	if !bytes.Equal(data, expectedData) {
		t.Errorf("Get(\"en\") = %v, want %v", data, expectedData)
	}
}

func TestLoader_Get_MissingLang(t *testing.T) {
	t.Parallel()

	expectedData := []byte(`{"vocab": ["hello", "world"]}`)

	// Create a mock storage that returns data
	mockStorage := &mockStorageService{
		downloadFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(expectedData)), nil
		},
	}

	// Create a loader and manually add to cache
	loader := NewLoader(mockStorage)
	loader.cache["en"] = expectedData

	// Test retrieval of non-existent language
	data, ok := loader.Get("fr")
	if ok {
		t.Error("Get(\"fr\") = true, want false")
	}
	if data != nil {
		t.Errorf("Get(\"fr\") = %v, want nil", data)
	}
}

func TestLoader_download_Success(t *testing.T) {
	t.Parallel()

	expectedData := []byte(`{"vocab": ["test"]}`)
	mockStorage := &mockStorageService{
		downloadFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(expectedData)), nil
		},
	}

	loader := &Loader{
		storage: mockStorage,
		cache:   make(map[string][]byte),
	}

	data, err := loader.download(context.Background(), "system/tokenizer/en.json")
	if err != nil {
		t.Errorf("download() error = %v", err)
	}
	if !bytes.Equal(data, expectedData) {
		t.Errorf("download() = %v, want %v", data, expectedData)
	}
}

func TestLoader_download_StorageError(t *testing.T) {
	t.Parallel()

	mockStorage := &mockStorageService{
		downloadFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return nil, errors.New("storage error")
		},
	}

	loader := &Loader{
		storage: mockStorage,
		cache:   make(map[string][]byte),
	}

	_, err := loader.download(context.Background(), "system/tokenizer/en.json")
	if err == nil {
		t.Error("download() expected error, got nil")
	}
}

func TestLoader_Get_WithLock(t *testing.T) {
	t.Parallel()

	expectedData := []byte(`{"vocab": ["test"]}`)
	loader := &Loader{
		cache: map[string][]byte{
			"en": expectedData,
		},
		mu: sync.RWMutex{},
	}

	data, ok := loader.Get("en")
	if !ok {
		t.Error("Get(\"en\") ok = false, want true")
	}
	if !bytes.Equal(data, expectedData) {
		t.Errorf("Get(\"en\") = %v, want %v", data, expectedData)
	}
}

func TestLoader_Get_MissingKey(t *testing.T) {
	t.Parallel()

	loader := &Loader{
		cache: make(map[string][]byte),
		mu:    sync.RWMutex{},
	}

	data, ok := loader.Get("missing")
	if ok {
		t.Error("Get(\"missing\") ok = true, want false")
	}
	if data != nil {
		t.Errorf("Get(\"missing\") = %v, want nil", data)
	}
}

func TestLoader_Preload_Success(t *testing.T) {

	// Save original configs and restore after
	originalConfigs := langconfig.Configs
	defer func() { langconfig.Configs = originalConfigs }()

	// Set up test config with one language
	langconfig.Configs = map[string]langconfig.LanguageConfig{
		"en": {
			TokenizerData: "system/tokenizer/en.json",
		},
	}

	expectedData := []byte(`{"vocab": ["test"]}`)
	mockStorage := &mockStorageService{
		downloadFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(expectedData)), nil
		},
	}

	loader := NewLoader(mockStorage)

	err := loader.Preload(context.Background())
	if err != nil {
		t.Errorf("Preload() error = %v", err)
	}

	// Verify cache was populated
	data, ok := loader.Get("en")
	if !ok {
		t.Error("Preload() did not populate cache for 'en'")
	}
	if !bytes.Equal(data, expectedData) {
		t.Errorf("Preload() cached data = %v, want %v", data, expectedData)
	}
}

func TestLoader_Preload_SkipsEmptyTokenizerData(t *testing.T) {

	// Save original configs and restore after
	originalConfigs := langconfig.Configs
	defer func() { langconfig.Configs = originalConfigs }()

	// Set up test config with empty tokenizer data
	langconfig.Configs = map[string]langconfig.LanguageConfig{
		"en": {
			TokenizerData: "", // Empty - should be skipped
		},
	}

	mockStorage := &mockStorageService{
		downloadFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			t.Error("download should not be called for empty TokenizerData")
			return nil, errors.New("should not be called")
		},
	}

	loader := NewLoader(mockStorage)

	err := loader.Preload(context.Background())
	if err != nil {
		t.Errorf("Preload() error = %v", err)
	}

	// Verify cache is empty
	loader.mu.RLock()
	defer loader.mu.RUnlock()

	if len(loader.cache) != 0 {
		t.Errorf("Preload() cache should be empty, got %d items", len(loader.cache))
	}
}

func TestLoader_Preload_ContinuesOnError(t *testing.T) {

	// Save original configs and restore after
	originalConfigs := langconfig.Configs
	defer func() { langconfig.Configs = originalConfigs }()

	// Set up test config with multiple languages
	langconfig.Configs = map[string]langconfig.LanguageConfig{
		"en": {
			TokenizerData: "system/tokenizer/en.json",
		},
		"fr": {
			TokenizerData: "system/tokenizer/fr.json",
		},
	}

	mockStorage := &mockStorageService{
		downloadFunc: func(ctx context.Context, key string) (io.ReadCloser, error) {
			if key == "system/tokenizer/en.json" {
				return nil, errors.New("download error")
			}
			return io.NopCloser(bytes.NewReader([]byte(`{}`))), nil
		},
	}

	loader := NewLoader(mockStorage)

	err := loader.Preload(context.Background())
	// Should not error even if one download fails
	if err != nil {
		t.Errorf("Preload() should not error on individual failures, got: %v", err)
	}

	// Verify 'fr' was still cached despite 'en' failure
	_, okEn := loader.Get("en")
	_, okFr := loader.Get("fr")

	if okEn {
		t.Error("Preload() should not cache failed download for 'en'")
	}
	if !okFr {
		t.Error("Preload() should cache successful download for 'fr'")
	}
}

// Compile-time check: mockStorage implements storage.StorageService
var _ storage.StorageService = (*mockStorageService)(nil)
