package usage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// ErrBudgetExceeded is returned for any request made after the run's cost
// budget has been spent. It stops a misconfigured run from burning money.
var ErrBudgetExceeded = errors.New("eval cost budget exceeded")

// Budget is a run-wide spending cap in USD. A zero or negative limit disables it.
type Budget struct {
	mu    sync.Mutex
	limit float64
	spent float64
}

// NewBudget returns a budget with the given USD limit.
func NewBudget(limitUSD float64) *Budget { return &Budget{limit: limitUSD} }

// Spend records cost.
func (b *Budget) Spend(usd float64) {
	if b == nil {
		return
	}
	b.mu.Lock()
	b.spent += usd
	b.mu.Unlock()
}

// Exceeded reports whether spending has reached the limit.
func (b *Budget) Exceeded() bool {
	if b == nil {
		return false
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.limit > 0 && b.spent >= b.limit
}

// Spent returns the USD spent so far.
func (b *Budget) Spent() float64 {
	if b == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.spent
}

// Transport is an http.RoundTripper that records the "usage" block of
// OpenAI-compatible responses into the meters on the request context.
type Transport struct {
	Base    http.RoundTripper
	Pricing *Pricing
	Budget  *Budget
}

// maxMeteredBody bounds how much of a response body is buffered for parsing.
const maxMeteredBody = 32 << 20

type usageEnvelope struct {
	Model string `json:"model"`
	Usage *struct {
		PromptTokens     int      `json:"prompt_tokens"`
		CompletionTokens int      `json:"completion_tokens"`
		TotalTokens      int      `json:"total_tokens"`
		Cost             *float64 `json:"cost"` // OpenRouter reports the billed cost
	} `json:"usage"`
}

type requestEnvelope struct {
	Model string `json:"model"`
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Budget.Exceeded() {
		return nil, fmt.Errorf("%w (spent $%.4f)", ErrBudgetExceeded, t.Budget.Spent())
	}
	reqModel := peekRequestModel(req)
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	res, err := base.RoundTrip(req)
	if err != nil || res.StatusCode != http.StatusOK || res.Body == nil {
		return res, err //nolint:wrapcheck // transparent proxy: callers expect the transport's own error
	}
	body, readErr := io.ReadAll(io.LimitReader(res.Body, maxMeteredBody))
	_ = res.Body.Close()
	res.Body = io.NopCloser(bytes.NewReader(body))
	if readErr != nil {
		return res, nil
	}
	t.record(req, reqModel, body)
	return res, nil
}

func (t *Transport) record(req *http.Request, reqModel string, body []byte) {
	var env usageEnvelope
	if json.Unmarshal(body, &env) != nil || env.Usage == nil {
		return
	}
	model := env.Model
	if model == "" {
		model = reqModel
	}
	u := Usage{Calls: 1, PromptTokens: env.Usage.PromptTokens, CompletionTokens: env.Usage.CompletionTokens}
	if u.PromptTokens == 0 && u.CompletionTokens == 0 {
		u.PromptTokens = env.Usage.TotalTokens // embeddings report only total
	}
	switch {
	case env.Usage.Cost != nil:
		u.CostUSD = *env.Usage.Cost
	default:
		price, ok := t.Pricing.Lookup(model)
		if !ok && reqModel != "" {
			price, ok = t.Pricing.Lookup(reqModel)
		}
		if ok {
			u.CostUSD = price.Cost(u.PromptTokens, u.CompletionTokens)
		} else {
			u.UnpricedCalls = 1
		}
	}
	t.Budget.Spend(u.CostUSD)
	for _, m := range Meters(req.Context()) {
		m.Record(model, u)
	}
}

func peekRequestModel(req *http.Request) string {
	if req.GetBody == nil {
		return ""
	}
	rc, err := req.GetBody()
	if err != nil {
		return ""
	}
	defer func() { _ = rc.Close() }()
	raw, err := io.ReadAll(io.LimitReader(rc, maxMeteredBody))
	if err != nil {
		return ""
	}
	var env requestEnvelope
	if json.Unmarshal(raw, &env) != nil {
		return ""
	}
	return env.Model
}
