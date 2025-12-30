package tokenizer

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/onurceri/botla-co/pkg/langconfig"
	"github.com/onurceri/botla-co/pkg/storage"
)

// Loader handles loading and caching tokenizer training data from R2.
type Loader struct {
	storage storage.StorageService
	cache   map[string][]byte // langCode -> training data
	mu      sync.RWMutex
}

var (
	globalLoader *Loader
	initOnce     sync.Once
)

// Init initializes the global tokenizer loader and preloads all language data.
// Should be called once at application startup.
func Init(ctx context.Context, s storage.StorageService) error {
	var initErr error
	initOnce.Do(func() {
		globalLoader = &Loader{
			storage: s,
			cache:   make(map[string][]byte),
		}
		initErr = globalLoader.preload(ctx)
	})
	return initErr
}

// GetTrainingData returns cached training data for the given language code.
// Returns nil, false if not loaded or language not supported.
func GetTrainingData(langCode string) ([]byte, bool) {
	if globalLoader == nil {
		return nil, false
	}
	return globalLoader.get(langCode)
}

// preload downloads tokenizer data for all supported languages.
func (l *Loader) preload(ctx context.Context) error {
	for code := range langconfig.Configs {
		cfg := langconfig.Configs[code]
		if cfg.TokenizerData == "" {
			continue
		}

		key := storage.SystemKey("tokenizer", code+".json")
		data, err := l.download(ctx, key)
		if err != nil {
			// Log but don't fail - fallback to English tokenizer will be used
			continue
		}

		l.mu.Lock()
		l.cache[code] = data
		l.mu.Unlock()
	}
	return nil
}

// download fetches a file from R2 and returns its contents.
func (l *Loader) download(ctx context.Context, key string) ([]byte, error) {
	reader, err := l.storage.DownloadFile(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", key, err)
	}
	defer func() {
		_ = reader.Close()
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", key, err)
	}
	return data, nil
}

// get returns cached data for a language code.
func (l *Loader) get(langCode string) ([]byte, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	data, ok := l.cache[langCode]
	return data, ok
}
