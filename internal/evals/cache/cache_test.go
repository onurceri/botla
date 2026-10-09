package cache

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKey(t *testing.T) {
	assert.Len(t, Key("a"), 64)
	assert.Equal(t, Key("a", "b"), Key("a", "b"))
	assert.NotEqual(t, Key("ab", "c"), Key("a", "bc"), "length prefix prevents boundary collisions")
	assert.NotEqual(t, Key("a"), Key("a", ""))
}

func TestRoundTrip(t *testing.T) {
	c := New(t.TempDir())
	require.True(t, c.Enabled())
	k := Key("x")

	var got []float32
	ok, err := c.Get("emb", k, &got)
	require.NoError(t, err)
	assert.False(t, ok)

	require.NoError(t, c.Put("emb", k, []float32{1, 2.5}))
	ok, err = c.Get("emb", k, &got)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []float32{1, 2.5}, got)

	// Namespaces are isolated.
	ok, err = c.Get("other", k, &got)
	require.NoError(t, err)
	assert.False(t, ok)

	// Short keys still work (no sharding beyond the key itself).
	require.NoError(t, c.Put("emb", "k", 1))
	var one int
	ok, err = c.Get("emb", "k", &one)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCorruptEntryIsAMiss(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	k := Key("y")
	require.NoError(t, c.Put("ns", k, "ok"))
	require.NoError(t, os.WriteFile(c.path("ns", k), []byte("{not json"), 0o600))
	var s string
	ok, err := c.Get("ns", k, &s)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestReadErrorIsReported(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	k := Key("dir")
	// Make the entry path a directory so ReadFile fails with a non-NotExist error.
	require.NoError(t, os.MkdirAll(c.path("ns", k), 0o750))
	var s string
	_, err := c.Get("ns", k, &s)
	assert.Error(t, err)
}

func TestDisabledCache(t *testing.T) {
	for _, c := range []*Cache{nil, New("")} {
		assert.False(t, c.Enabled())
		require.NoError(t, c.Put("ns", "k", 1))
		var v int
		ok, err := c.Get("ns", "k", &v)
		require.NoError(t, err)
		assert.False(t, ok)
	}
}

func TestPutErrors(t *testing.T) {
	c := New(t.TempDir())
	assert.Error(t, c.Put("ns", "k", make(chan int)), "unencodable value")

	file := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(file, nil, 0o600))
	assert.Error(t, New(file).Put("ns", "k", 1), "cache dir is a file")

	// Target path occupied by a non-empty directory: rename fails.
	c2 := New(t.TempDir())
	k := Key("z")
	require.NoError(t, os.MkdirAll(filepath.Join(c2.path("ns", k), "child"), 0o750))
	assert.Error(t, c2.Put("ns", k, 1))
}

func TestConcurrentPuts(t *testing.T) {
	c := New(t.TempDir())
	k := Key("shared")
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = c.Put("ns", k, i)
		}(i)
	}
	wg.Wait()
	var v int
	ok, err := c.Get("ns", k, &v)
	require.NoError(t, err)
	assert.True(t, ok, "an entry is always complete, whichever writer won")
}

func TestPutReadOnlyDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	c := New(t.TempDir())
	k := Key("ro")
	shard := filepath.Dir(c.path("ns", k))
	require.NoError(t, os.MkdirAll(shard, 0o750))
	require.NoError(t, os.Chmod(shard, 0o500))       // #nosec G302 -- test needs a read-only dir
	t.Cleanup(func() { _ = os.Chmod(shard, 0o750) }) // #nosec G302
	assert.Error(t, c.Put("ns", k, 1))
}
