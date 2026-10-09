// Package usage meters tokens, cost and call counts for every LLM/embedding
// request the eval makes, without modifying product code.
//
// The product's OpenAI/OpenRouter clients only surface total tokens, so the
// eval injects an http.RoundTripper (via the clients' existing WithHTTPClient
// options) that reads the provider's "usage" block from each response. Usage
// is attributed to every Meter attached to the request context, which lets one
// request count towards its case, its phase (generation/judge) and the run.
package usage

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Usage is an additive tally of model usage.
type Usage struct {
	Calls            int     `json:"calls"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	CostUSD          float64 `json:"cost_usd"`
	// UnpricedCalls counts calls whose cost could not be determined (provider
	// returned no cost and the model is missing from the pricing table).
	UnpricedCalls int `json:"unpriced_calls,omitempty"`
}

// Add returns u + o.
func (u Usage) Add(o Usage) Usage {
	return Usage{
		Calls:            u.Calls + o.Calls,
		PromptTokens:     u.PromptTokens + o.PromptTokens,
		CompletionTokens: u.CompletionTokens + o.CompletionTokens,
		CostUSD:          u.CostUSD + o.CostUSD,
		UnpricedCalls:    u.UnpricedCalls + o.UnpricedCalls,
	}
}

// TotalTokens is prompt + completion tokens.
func (u Usage) TotalTokens() int { return u.PromptTokens + u.CompletionTokens }

// Meter accumulates usage per model. It is safe for concurrent use.
type Meter struct {
	mu      sync.Mutex
	byModel map[string]Usage
}

// NewMeter returns an empty meter.
func NewMeter() *Meter { return &Meter{byModel: map[string]Usage{}} }

// Record adds usage for a model.
func (m *Meter) Record(model string, u Usage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byModel[model] = m.byModel[model].Add(u)
}

// Total returns the sum over all models.
func (m *Meter) Total() Usage {
	m.mu.Lock()
	defer m.mu.Unlock()
	var t Usage
	for _, u := range m.byModel {
		t = t.Add(u)
	}
	return t
}

// ByModel returns a copy of the per-model tallies.
func (m *Meter) ByModel() map[string]Usage {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[string]Usage, len(m.byModel))
	for k, v := range m.byModel {
		out[k] = v
	}
	return out
}

type metersKey struct{}

// WithMeter returns a context whose requests are also recorded into m.
func WithMeter(ctx context.Context, m *Meter) context.Context {
	prev := Meters(ctx)
	next := make([]*Meter, 0, len(prev)+1)
	next = append(next, prev...)
	next = append(next, m)
	return context.WithValue(ctx, metersKey{}, next)
}

// Meters returns the meters attached to ctx.
func Meters(ctx context.Context) []*Meter {
	ms, _ := ctx.Value(metersKey{}).([]*Meter)
	return ms
}

// Price is USD per one million tokens.
type Price struct {
	Input  float64 `yaml:"input" json:"input"`
	Output float64 `yaml:"output" json:"output"`
}

// Pricing maps model-name prefixes to prices.
type Pricing struct {
	Models map[string]Price `yaml:"models"`
}

// LoadPricing reads a pricing YAML file.
func LoadPricing(path string) (*Pricing, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- operator-supplied config file
	if err != nil {
		return nil, fmt.Errorf("read pricing: %w", err)
	}
	var p Pricing
	if err := yaml.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("parse pricing: %w", err)
	}
	return &p, nil
}

// Lookup finds the price for a model by longest matching prefix, so that a
// dated snapshot such as "gpt-4o-mini-2024-07-18" matches "gpt-4o-mini". A
// provider prefix ("openai/") is also tried without the prefix.
func (p *Pricing) Lookup(model string) (Price, bool) {
	if p == nil || model == "" {
		return Price{}, false
	}
	candidates := []string{model}
	if _, bare, ok := strings.Cut(model, "/"); ok {
		candidates = append(candidates, bare)
	}
	keys := make([]string, 0, len(p.Models))
	for k := range p.Models {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, c := range candidates {
		for _, k := range keys {
			if strings.HasPrefix(c, k) {
				return p.Models[k], true
			}
		}
	}
	return Price{}, false
}

// Cost computes the USD cost of the given token counts.
func (pr Price) Cost(promptTokens, completionTokens int) float64 {
	return (float64(promptTokens)*pr.Input + float64(completionTokens)*pr.Output) / 1_000_000
}
