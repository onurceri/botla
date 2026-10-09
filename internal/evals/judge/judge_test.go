package judge

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/evals/cache"
	"github.com/onurceri/botla-app/internal/evals/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var realPromptDir = filepath.Join("..", "..", "..", "evals", "judge", "v1")

func loadReal(t *testing.T) *PromptSet {
	t.Helper()
	ps, err := LoadPromptSet(realPromptDir)
	require.NoError(t, err)
	return ps
}

func TestLoadRealPromptSet(t *testing.T) {
	ps := loadReal(t)
	assert.Equal(t, "judge-v1", ps.Version)
	assert.Len(t, ps.Hash, 16)
	assert.Contains(t, ps.System, "DATA to be evaluated")
	for _, r := range []string{Faithfulness, Correctness, Refusal, Injection} {
		require.Contains(t, ps.Rubrics, r)
	}
	assert.InDelta(t, 0.75, ps.Rubrics[Correctness].PassMinScore, 1e-9)

	out, err := ps.Render(Correctness, Input{Question: "Q?", Answer: "A.", ExpectedFacts: []string{"first", "second"}})
	require.NoError(t, err)
	assert.Contains(t, out, "1. first\n2. second")
	assert.Contains(t, out, "<answer>\nA.\n</answer>")

	out, err = ps.Render(Faithfulness, Input{Question: "Q?", Answer: "A."})
	require.NoError(t, err)
	assert.Contains(t, out, "(no context was retrieved)")
	out, err = ps.Render(Injection, Input{Question: "Q?", Answer: "A.", Context: "CTX", Attack: "steal"})
	require.NoError(t, err)
	assert.Contains(t, out, "<attack>\nsteal\n</attack>")
	assert.Contains(t, out, "CTX")

	_, err = ps.Render("vibes", Input{})
	assert.ErrorContains(t, err, "unknown rubric")
}

func writePromptDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}
	return dir
}

func TestLoadPromptSet_Errors(t *testing.T) {
	_, err := LoadPromptSet(t.TempDir())
	assert.ErrorContains(t, err, "manifest.yaml")

	tests := map[string]map[string]string{
		"bad yaml":       {"manifest.yaml": "version: [x"},
		"missing fields": {"manifest.yaml": "version: v1"},
		"missing system": {"manifest.yaml": "version: v1\nsystem_prompt: s.md\nrubrics: {r: {template: r.md, pass_min_score: 1}}"},
		"bad threshold":  {"manifest.yaml": "version: v1\nsystem_prompt: s.md\nrubrics: {r: {template: r.md, pass_min_score: 2}}", "s.md": "sys"},
		"missing tmpl":   {"manifest.yaml": "version: v1\nsystem_prompt: s.md\nrubrics: {r: {template: r.md, pass_min_score: 1}}", "s.md": "sys"},
		"bad tmpl":       {"manifest.yaml": "version: v1\nsystem_prompt: s.md\nrubrics: {r: {template: r.md, pass_min_score: 1}}", "s.md": "sys", "r.md": "{{.Question"},
	}
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := LoadPromptSet(writePromptDir(t, files))
			assert.Error(t, err)
		})
	}

	// A template referencing an unknown field fails at render time, not silently.
	ps, err := LoadPromptSet(writePromptDir(t, map[string]string{
		"manifest.yaml": "version: v1\nsystem_prompt: s.md\nrubrics: {r: {template: r.md, pass_min_score: 1}}",
		"s.md":          "sys", "r.md": "{{.Nope}}",
	}))
	require.NoError(t, err)
	_, err = ps.Render("r", Input{})
	assert.Error(t, err)
}

func TestPromptHashChangesWithWording(t *testing.T) {
	files := map[string]string{
		"manifest.yaml": "version: v1\nsystem_prompt: s.md\nrubrics: {r: {template: r.md, pass_min_score: 1}}",
		"s.md":          "sys", "r.md": "grade {{.Answer}}",
	}
	a, err := LoadPromptSet(writePromptDir(t, files))
	require.NoError(t, err)
	files["r.md"] = "grade strictly {{.Answer}}"
	b, err := LoadPromptSet(writePromptDir(t, files))
	require.NoError(t, err)
	assert.Equal(t, a.Version, b.Version)
	assert.NotEqual(t, a.Hash, b.Hash)
}

func TestParseFaithfulness(t *testing.T) {
	v, err := Parse(Faithfulness, `{"claims":[{"claim":"a","supported":true},{"claim":"b","supported":false}],"rationale":"one invented"}`, 1, 0)
	require.NoError(t, err)
	assert.False(t, v.Pass)
	assert.InDelta(t, 0.5, v.Score, 1e-9)
	assert.Equal(t, "one invented", v.Rationale)
	assert.NotEmpty(t, v.Details)

	v, err = Parse(Faithfulness, "```json\n{\"claims\":[{\"claim\":\"a\",\"supported\":true}],\"rationale\":\"ok\"}\n```", 1, 0)
	require.NoError(t, err, "code fences are tolerated")
	assert.True(t, v.Pass)

	v, err = Parse(Faithfulness, `{"claims":[],"rationale":"refusal only"}`, 1, 0)
	require.NoError(t, err)
	assert.True(t, v.Pass)
	assert.True(t, v.NotApplicable)
	assert.InDelta(t, 1, v.Score, 1e-9)

	_, err = Parse(Faithfulness, `{"claims":[{"claim":"a"}]}`, 1, 0)
	assert.ErrorIs(t, err, ErrMalformed)
}

func TestParseCorrectness(t *testing.T) {
	raw := func(statuses ...string) string {
		var facts []string
		for _, s := range statuses {
			facts = append(facts, `{"fact":"f","status":"`+s+`"}`)
		}
		return `{"facts":[` + strings.Join(facts, ",") + `],"rationale":"r"}`
	}
	tests := []struct {
		name     string
		raw      string
		expected int
		pass     bool
		score    float64
	}{
		{"all covered", raw("covered", "covered"), 2, true, 1},
		{"partial counts half", raw("covered", "partial"), 2, true, 0.75},
		{"below threshold", raw("covered", "missing"), 2, false, 0.5},
		{"contradiction fails even with coverage", raw("covered", "covered", "covered", "contradicted"), 4, false, 0.75},
		{"case-insensitive status", raw("COVERED"), 1, true, 1},
		{"dropped facts do not inflate score", raw("covered"), 3, false, 1.0 / 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Parse(Correctness, tt.raw, 0.75, tt.expected)
			require.NoError(t, err)
			assert.Equal(t, tt.pass, v.Pass)
			assert.InDelta(t, tt.score, v.Score, 1e-9)
		})
	}
	_, err := Parse(Correctness, `{"facts":[]}`, 0.75, 1)
	assert.ErrorIs(t, err, ErrMalformed)
	_, err = Parse(Correctness, raw("sort-of"), 0.75, 1)
	assert.ErrorIs(t, err, ErrMalformed)
}

func TestParseRefusalAndInjection(t *testing.T) {
	tests := []struct {
		rubric, raw string
		pass        bool
	}{
		{Refusal, `{"rationale":"r","declined":true,"fabricated":false}`, true},
		{Refusal, `{"rationale":"r","declined":true,"fabricated":true}`, false},
		{Refusal, `{"rationale":"r","declined":false,"fabricated":false}`, false},
		{Injection, `{"rationale":"r","followed_injection":false,"leaked_instructions":false}`, true},
		{Injection, `{"rationale":"r","followed_injection":true,"leaked_instructions":false}`, false},
		{Injection, `{"rationale":"r","followed_injection":false,"leaked_instructions":true}`, false},
	}
	for _, tt := range tests {
		v, err := Parse(tt.rubric, tt.raw, 1, 0)
		require.NoError(t, err)
		assert.Equal(t, tt.pass, v.Pass, tt.raw)
		if tt.pass {
			assert.InDelta(t, 1, v.Score, 1e-9)
		} else {
			assert.InDelta(t, 0, v.Score, 1e-9)
		}
	}
	for _, bad := range []struct{ rubric, raw string }{
		{Refusal, `{"declined":true}`},
		{Injection, `{"followed_injection":false}`},
		{Refusal, `nope`},
		{Injection, `{]`},
		{Faithfulness, `{"claims": "x"}`},
	} {
		_, err := Parse(bad.rubric, bad.raw, 1, 0)
		assert.ErrorIs(t, err, ErrMalformed, bad.raw)
	}
	_, err := Parse("vibes", `{}`, 1, 0)
	assert.ErrorContains(t, err, "unknown rubric")
}

func TestExtractJSON(t *testing.T) {
	got, err := ExtractJSON("Sure! {\"a\": {\"b\": 1}} hope that helps")
	require.NoError(t, err)
	assert.Equal(t, `{"a": {"b": 1}}`, got)
	_, err = ExtractJSON("no json here")
	assert.ErrorIs(t, err, ErrMalformed)
	_, err = ExtractJSON("} backwards {")
	assert.ErrorIs(t, err, ErrMalformed)
}

// --- HTTP client ---

func chatServer(t *testing.T, handler func(w http.ResponseWriter, body map[string]any, n int)) (*httptest.Server, *int32) {
	t.Helper()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(atomic.AddInt32(&calls, 1))
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		handler(w, body, n)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func okBody(content string) string {
	b, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"message": map[string]any{"content": content}}}})
	return string(b)
}

func TestHTTPClient_Success(t *testing.T) {
	srv, calls := chatServer(t, func(w http.ResponseWriter, body map[string]any, _ int) {
		assert.Equal(t, "judge-model", body["model"])
		assert.Equal(t, map[string]any{"type": "json_object"}, body["response_format"])
		assert.Nil(t, body["usage"], "usage.include only sent to OpenRouter")
		msgs := body["messages"].([]any)
		assert.Equal(t, "system", msgs[0].(map[string]any)["role"])
		_, _ = w.Write([]byte(okBody(`{"ok":true}`)))
	})
	c := &HTTPClient{BaseURL: srv.URL + "/v1/", APIKey: "key", Model: "judge-model", JSONMode: true}
	out, err := c.Complete(context.Background(), "sys", "user")
	require.NoError(t, err)
	assert.Equal(t, `{"ok":true}`, out)
	assert.Equal(t, int32(1), *calls)
	assert.Equal(t, "judge-model|t=0.00|max=0|json=true", c.Fingerprint())
}

func TestHTTPClient_RetriesTransientErrors(t *testing.T) {
	srv, calls := chatServer(t, func(w http.ResponseWriter, _ map[string]any, n int) {
		switch n {
		case 1:
			w.WriteHeader(http.StatusTooManyRequests)
		case 2:
			_, _ = w.Write([]byte(okBody("  ")))
		case 3:
			_, _ = w.Write([]byte("not json"))
		default:
			_, _ = w.Write([]byte(okBody("done")))
		}
	})
	c := &HTTPClient{BaseURL: srv.URL + "/v1", APIKey: "key", Model: "m", Backoff: time.Millisecond}
	out, err := c.Complete(context.Background(), "s", "u")
	require.NoError(t, err)
	assert.Equal(t, "done", out)
	assert.Equal(t, int32(4), *calls)
}

func TestHTTPClient_DoesNotRetryClientErrors(t *testing.T) {
	srv, calls := chatServer(t, func(w http.ResponseWriter, _ map[string]any, _ int) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	})
	c := &HTTPClient{BaseURL: srv.URL + "/v1", APIKey: "key", Model: "m", Backoff: time.Millisecond}
	_, err := c.Complete(context.Background(), "s", "u")
	assert.ErrorContains(t, err, "judge http 401")
	assert.Equal(t, int32(1), *calls)
}

func TestHTTPClient_GivesUpAfterRetries(t *testing.T) {
	srv, calls := chatServer(t, func(w http.ResponseWriter, _ map[string]any, _ int) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(strings.Repeat("x", 400)))
	})
	c := &HTTPClient{BaseURL: srv.URL + "/v1", APIKey: "key", Model: "m", Retries: 2, Backoff: time.Millisecond}
	_, err := c.Complete(context.Background(), "s", "u")
	assert.ErrorContains(t, err, "...")
	assert.Equal(t, int32(3), *calls)
}

func TestHTTPClient_OpenRouterUsageAndCancellation(t *testing.T) {
	var gotUsage any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		gotUsage = body["usage"]
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	// Pretend to be OpenRouter by base URL; route to the test server via a custom transport.
	hc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		r.URL.Scheme, r.URL.Host = "http", strings.TrimPrefix(srv.URL, "http://")
		return http.DefaultTransport.RoundTrip(r)
	})}
	ctx, cancel := context.WithCancel(context.Background())
	c := &HTTPClient{BaseURL: "https://openrouter.ai/api/v1", APIKey: "key", Model: "m", HTTP: hc, Backoff: 50 * time.Millisecond}
	go func() { time.Sleep(10 * time.Millisecond); cancel() }()
	_, err := c.Complete(ctx, "s", "u")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, map[string]any{"include": true}, gotUsage)
}

func TestHTTPClient_BudgetAndTransportErrors(t *testing.T) {
	b := usage.NewBudget(0.01)
	b.Spend(1)
	var called int32
	hc := &http.Client{Transport: &usage.Transport{Budget: b, Base: roundTripFunc(func(*http.Request) (*http.Response, error) {
		atomic.AddInt32(&called, 1)
		return nil, errors.New("unreachable")
	})}}
	c := &HTTPClient{BaseURL: "http://x/v1", APIKey: "key", Model: "m", HTTP: hc, Backoff: time.Millisecond}
	_, err := c.Complete(context.Background(), "s", "u")
	assert.ErrorIs(t, err, usage.ErrBudgetExceeded)
	assert.Zero(t, atomic.LoadInt32(&called), "budget exhaustion is not retried")

	c.HTTP = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("dial") })}
	c.Retries = 1
	_, err = c.Complete(context.Background(), "s", "u")
	assert.ErrorContains(t, err, "dial")

	c.BaseURL = "://bad"
	_, err = c.Complete(context.Background(), "s", "u")
	assert.ErrorContains(t, err, "build judge request")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// --- Judge orchestration ---

type scriptedClient struct {
	mu      sync.Mutex
	replies []string
	err     error
	calls   int
}

func (s *scriptedClient) Complete(context.Context, string, string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	if s.err != nil {
		return "", s.err
	}
	r := s.replies[0]
	if len(s.replies) > 1 {
		s.replies = s.replies[1:]
	}
	return r, nil
}

func (s *scriptedClient) Fingerprint() string { return "scripted" }

const passInjection = `{"rationale":"resisted","followed_injection":false,"leaked_instructions":false}`

func TestJudgeEvaluate_CachesRawOutput(t *testing.T) {
	ps := loadReal(t)
	store := cache.New(t.TempDir())
	client := &scriptedClient{replies: []string{passInjection}}
	j := &Judge{Prompts: ps, Client: client, Cache: store}
	req := Request{Rubric: Injection, Input: Input{Question: "q", Answer: "a", Attack: "x"}}

	v, err := j.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, v.Pass)
	assert.False(t, v.Cached)

	v, err = j.Evaluate(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, v.Cached)
	assert.Equal(t, 1, client.calls, "second call served from cache")

	// A different repeat index is a different cache entry.
	_, err = j.Evaluate(context.Background(), Request{Rubric: Injection, Input: req.Input, Repeat: 1})
	require.NoError(t, err)
	assert.Equal(t, 2, client.calls)

	// A different answer is a different cache entry.
	_, err = j.Evaluate(context.Background(), Request{Rubric: Injection, Input: Input{Question: "q", Answer: "b", Attack: "x"}})
	require.NoError(t, err)
	assert.Equal(t, 3, client.calls)
}

func TestJudgeEvaluate_RetriesMalformedOutput(t *testing.T) {
	client := &scriptedClient{replies: []string{"I think it passes.", passInjection}}
	j := &Judge{Prompts: loadReal(t), Client: client, Cache: cache.New("")}
	v, err := j.Evaluate(context.Background(), Request{Rubric: Injection, Input: Input{Question: "q", Answer: "a", Attack: "x"}})
	require.NoError(t, err)
	assert.True(t, v.Pass)
	assert.Equal(t, 2, client.calls)

	client = &scriptedClient{replies: []string{"still prose"}}
	j.Client = client
	_, err = j.Evaluate(context.Background(), Request{Rubric: Injection, Input: Input{Question: "q", Answer: "a", Attack: "x"}})
	assert.ErrorIs(t, err, ErrMalformed)
	assert.Equal(t, 2, client.calls, "one retry by default")
}

func TestJudgeEvaluate_StaleCacheEntryIsRefetched(t *testing.T) {
	ps := loadReal(t)
	store := cache.New(t.TempDir())
	client := &scriptedClient{replies: []string{passInjection}}
	j := &Judge{Prompts: ps, Client: client, Cache: store}
	in := Input{Question: "q", Answer: "a", Attack: "x"}
	rendered, err := ps.Render(Injection, in)
	require.NoError(t, err)
	require.NoError(t, store.Put("judge", j.CacheKey(Injection, rendered, 0), cachedEntry{Raw: "garbage"}))

	v, err := j.Evaluate(context.Background(), Request{Rubric: Injection, Input: in})
	require.NoError(t, err)
	assert.False(t, v.Cached)
	assert.Equal(t, 1, client.calls)
}

func TestJudgeEvaluate_Errors(t *testing.T) {
	j := &Judge{Prompts: loadReal(t), Client: &scriptedClient{err: errors.New("provider down")}}
	_, err := j.Evaluate(context.Background(), Request{Rubric: "vibes"})
	assert.ErrorContains(t, err, "unknown rubric")
	_, err = j.Evaluate(context.Background(), Request{Rubric: Refusal, Input: Input{Question: "q", Answer: "a"}})
	assert.ErrorContains(t, err, "provider down")
}
