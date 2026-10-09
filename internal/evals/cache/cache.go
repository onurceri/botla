// Package cache is a small content-addressed JSON file cache. The eval uses it
// so that unchanged inputs (embeddings, capability summaries and, most
// importantly, judge verdicts) are never paid for twice.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
)

var tmpSeq atomic.Uint64

// Cache stores JSON values under hashed keys. A Cache with an empty Dir is a
// valid no-op cache (every Get misses, every Put is dropped).
type Cache struct {
	Dir string
}

// New returns a cache rooted at dir ("" disables caching).
func New(dir string) *Cache { return &Cache{Dir: dir} }

// Enabled reports whether the cache persists anything.
func (c *Cache) Enabled() bool { return c != nil && c.Dir != "" }

// Key hashes the parts into a cache key. Each part is length-prefixed so that
// ("ab","c") and ("a","bc") cannot collide.
func Key(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(strconv.Itoa(len(p))))
		_, _ = h.Write([]byte{':'})
		_, _ = h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Cache) path(ns, key string) string {
	shard := key
	if len(shard) > 2 {
		shard = shard[:2]
	}
	return filepath.Join(c.Dir, ns, shard, key+".json")
}

// Get loads the value stored under (ns, key) into v. It returns false on a miss.
func (c *Cache) Get(ns, key string, v any) (bool, error) {
	if !c.Enabled() {
		return false, nil
	}
	raw, err := os.ReadFile(c.path(ns, key))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read cache entry: %w", err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		// A corrupt entry is treated as a miss; it will be overwritten.
		return false, nil
	}
	return true, nil
}

// Put stores v under (ns, key). The write is atomic (temp file + rename) so
// concurrent workers never observe a partial entry.
func (c *Cache) Put(ns, key string, v any) error {
	if !c.Enabled() {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode cache entry: %w", err)
	}
	p := c.path(ns, key)
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	// Unique temp name per process and write, so concurrent writers never share it.
	tmp := fmt.Sprintf("%s.tmp-%d-%d", p, os.Getpid(), tmpSeq.Add(1))
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return fmt.Errorf("write cache entry: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("commit cache entry: %w", err)
	}
	return nil
}
