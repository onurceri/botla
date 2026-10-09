package usage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageAddAndMeter(t *testing.T) {
	u := Usage{Calls: 1, PromptTokens: 10, CompletionTokens: 5, CostUSD: 0.5, UnpricedCalls: 1}
	sum := u.Add(u)
	assert.Equal(t, Usage{Calls: 2, PromptTokens: 20, CompletionTokens: 10, CostUSD: 1, UnpricedCalls: 2}, sum)
	assert.Equal(t, 30, sum.TotalTokens())

	m := NewMeter()
	m.Record("a", u)
	m.Record("b", u)
	m.Record("a", u)
	assert.Equal(t, 3, m.Total().Calls)
	by := m.ByModel()
	assert.Equal(t, 2, by["a"].Calls)
	by["a"] = Usage{} // copy, not a view
	assert.Equal(t, 2, m.ByModel()["a"].Calls)
}

func TestMetersOnContext(t *testing.T) {
	ctx := context.Background()
	assert.Empty(t, Meters(ctx))
	a, b := NewMeter(), NewMeter()
	ctxA := WithMeter(ctx, a)
	ctxAB := WithMeter(ctxA, b)
	assert.Equal(t, []*Meter{a}, Meters(ctxA), "parent context is not modified")
	assert.Equal(t, []*Meter{a, b}, Meters(ctxAB))
}

func TestPricing(t *testing.T) {
	p := filepath.Join(t.TempDir(), "pricing.yaml")
	require.NoError(t, os.WriteFile(p, []byte("models:\n  gpt-4o: {input: 2.5, output: 10}\n  gpt-4o-mini: {input: 0.15, output: 0.6}\n"), 0o600))
	pr, err := LoadPricing(p)
	require.NoError(t, err)

	price, ok := pr.Lookup("gpt-4o-mini-2024-07-18")
	require.True(t, ok)
	assert.InDelta(t, 0.15, price.Input, 1e-12, "longest prefix wins")
	price, ok = pr.Lookup("openai/gpt-4o-2024-08-06")
	require.True(t, ok)
	assert.InDelta(t, 2.5, price.Input, 1e-12, "provider prefix is stripped")
	_, ok = pr.Lookup("claude")
	assert.False(t, ok)
	_, ok = pr.Lookup("")
	assert.False(t, ok)
	var nilPricing *Pricing
	_, ok = nilPricing.Lookup("gpt-4o")
	assert.False(t, ok)

	assert.InDelta(t, (1000*0.15+500*0.6)/1e6, Price{Input: 0.15, Output: 0.6}.Cost(1000, 500), 1e-15)

	_, err = LoadPricing(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.Error(t, err)
	bad := filepath.Join(t.TempDir(), "bad.yaml")
	require.NoError(t, os.WriteFile(bad, []byte("models: [not a map"), 0o600))
	_, err = LoadPricing(bad)
	assert.Error(t, err)
}

func TestBudget(t *testing.T) {
	var nilBudget *Budget
	nilBudget.Spend(1)
	assert.False(t, nilBudget.Exceeded())
	assert.Zero(t, nilBudget.Spent())

	b := NewBudget(1)
	b.Spend(0.6)
	assert.False(t, b.Exceeded())
	b.Spend(0.4)
	assert.True(t, b.Exceeded())
	assert.InDelta(t, 1.0, b.Spent(), 1e-12)
	assert.False(t, NewBudget(0).Exceeded(), "zero limit disables the budget")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func respond(status int, body string) roundTripFunc {
	return func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	}
}

func post(t *testing.T, tr http.RoundTripper, ctx context.Context, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://llm.test/v1/chat/completions", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	res, err := tr.RoundTrip(req)
	require.NoError(t, err)
	return res
}

func TestTransport_RecordsUsageWithPricing(t *testing.T) {
	pr := &Pricing{Models: map[string]Price{"gpt-4o-mini": {Input: 1, Output: 2}}}
	budget := NewBudget(10)
	tr := &Transport{Base: respond(200, `{"model":"gpt-4o-mini-2024-07-18","usage":{"prompt_tokens":1000000,"completion_tokens":500000,"total_tokens":1500000}}`), Pricing: pr, Budget: budget}
	caseMeter, runMeter := NewMeter(), NewMeter()
	ctx := WithMeter(WithMeter(context.Background(), runMeter), caseMeter)

	res := post(t, tr, ctx, `{"model":"gpt-4o-mini"}`)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "usage", "body is still readable by the caller")

	for _, m := range []*Meter{caseMeter, runMeter} {
		u := m.ByModel()["gpt-4o-mini-2024-07-18"]
		assert.Equal(t, 1, u.Calls)
		assert.Equal(t, 1000000, u.PromptTokens)
		assert.InDelta(t, 2.0, u.CostUSD, 1e-9)
	}
	assert.InDelta(t, 2.0, budget.Spent(), 1e-9)
}

func TestTransport_ProviderCostAndEmbeddings(t *testing.T) {
	m := NewMeter()
	ctx := WithMeter(context.Background(), m)

	// OpenRouter reports the billed cost; it wins over the pricing table.
	tr := &Transport{Base: respond(200, `{"model":"x/y","usage":{"prompt_tokens":10,"completion_tokens":5,"cost":0.0042}}`)}
	post(t, tr, ctx, `{}`)
	assert.InDelta(t, 0.0042, m.ByModel()["x/y"].CostUSD, 1e-12)

	// Embeddings report total_tokens only; model taken from the request when absent in response.
	pr := &Pricing{Models: map[string]Price{"text-embedding-3-small": {Input: 0.02}}}
	tr = &Transport{Base: respond(200, `{"usage":{"total_tokens":1000}}`), Pricing: pr}
	post(t, tr, ctx, `{"model":"text-embedding-3-small"}`)
	u := m.ByModel()["text-embedding-3-small"]
	assert.Equal(t, 1000, u.PromptTokens)
	assert.InDelta(t, 0.00002, u.CostUSD, 1e-12)

	// Unknown model: counted as an unpriced call.
	tr = &Transport{Base: respond(200, `{"model":"mystery","usage":{"prompt_tokens":1}}`), Pricing: pr}
	post(t, tr, ctx, `{"model":"mystery"}`)
	assert.Equal(t, 1, m.ByModel()["mystery"].UnpricedCalls)
}

func TestTransport_PassThroughCases(t *testing.T) {
	m := NewMeter()
	ctx := WithMeter(context.Background(), m)

	for _, tr := range []*Transport{
		{Base: respond(500, `{"usage":{"prompt_tokens":1}}`)}, // non-200 not metered
		{Base: respond(200, `not json`)},                      // unparsable
		{Base: respond(200, `{"data":[]}`)},                   // no usage block
	} {
		res := post(t, tr, ctx, `{"model":"m"}`)
		_ = res.Body.Close()
	}
	assert.Zero(t, m.Total().Calls)

	failing := &Transport{Base: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("boom") })}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://x", nil)
	_, err := failing.RoundTrip(req)
	assert.EqualError(t, err, "boom")
}

func TestTransport_BudgetBlocksRequests(t *testing.T) {
	b := NewBudget(0.01)
	b.Spend(0.02)
	called := false
	tr := &Transport{Budget: b, Base: roundTripFunc(func(*http.Request) (*http.Response, error) {
		called = true
		return nil, errors.New("should not be called")
	})}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://x", nil)
	_, err := tr.RoundTrip(req)
	assert.ErrorIs(t, err, ErrBudgetExceeded)
	assert.False(t, called)
}

func TestTransport_DefaultBaseWithRealServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"model":"m","usage":{"prompt_tokens":3,"completion_tokens":4}}`))
	}))
	defer srv.Close()
	m := NewMeter()
	client := &http.Client{Transport: &Transport{}}
	req, err := http.NewRequestWithContext(WithMeter(context.Background(), m), http.MethodPost, srv.URL, strings.NewReader(`{"model":"m"}`))
	require.NoError(t, err)
	res, err := client.Do(req)
	require.NoError(t, err)
	_ = res.Body.Close()
	assert.Equal(t, 7, m.Total().TotalTokens())
	assert.Equal(t, 1, m.Total().UnpricedCalls)
}

func TestPeekRequestModel(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://x", nil)
	assert.Empty(t, peekRequestModel(req), "no body")
	req, _ = http.NewRequest(http.MethodPost, "http://x", strings.NewReader("garbage"))
	assert.Empty(t, peekRequestModel(req))
	req.GetBody = func() (io.ReadCloser, error) { return nil, errors.New("x") }
	assert.Empty(t, peekRequestModel(req))
	req, _ = http.NewRequest(http.MethodPost, "http://x", strings.NewReader(`{"model":"gpt"}`))
	assert.Equal(t, "gpt", peekRequestModel(req))
}
