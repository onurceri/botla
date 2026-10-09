package judge

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/onurceri/botla-app/internal/evals/cache"
)

const cacheNamespace = "judge"

// Judge scores answers with a rubric prompt set and a model client.
type Judge struct {
	Prompts *PromptSet
	Client  ChatClient
	Cache   *cache.Cache
	// ParseRetries is how many extra calls to make when the output does not
	// match the rubric schema (default 1).
	ParseRetries int
}

// Request identifies one judgement. Repeat distinguishes intentional repeated
// runs (for consistency measurement) so they are cached separately.
type Request struct {
	Rubric        string
	Input         Input
	Repeat        int
	ExpectedFacts int
}

type cachedEntry struct {
	Raw string `json:"raw"`
}

// CacheKey returns the content hash under which a judgement is cached. It
// covers the prompt-set version and hash, the model fingerprint, the fully
// rendered prompt and the repeat index, so any change to any of them misses.
func (j *Judge) CacheKey(rubric, rendered string, repeat int) string {
	return cache.Key(j.Prompts.Version, j.Prompts.Hash, j.Client.Fingerprint(), rubric, rendered, strconv.Itoa(repeat))
}

// Evaluate grades one answer. Raw judge output is cached (not the parsed
// verdict), so improving the parser never requires paying for calls again.
func (j *Judge) Evaluate(ctx context.Context, req Request) (Verdict, error) {
	rubric, ok := j.Prompts.Rubrics[req.Rubric]
	if !ok {
		return Verdict{}, fmt.Errorf("unknown rubric %q", req.Rubric)
	}
	rendered, err := j.Prompts.Render(req.Rubric, req.Input)
	if err != nil {
		return Verdict{}, err
	}
	key := j.CacheKey(req.Rubric, rendered, req.Repeat)

	var hit cachedEntry
	if found, cacheErr := j.Cache.Get(cacheNamespace, key, &hit); cacheErr == nil && found {
		if v, parseErr := Parse(req.Rubric, hit.Raw, rubric.PassMinScore, req.ExpectedFacts); parseErr == nil {
			v.Cached = true
			return v, nil
		}
		// A cached entry that no longer parses (schema changed) is re-fetched.
	}

	retries := j.ParseRetries
	if retries <= 0 {
		retries = 1
	}
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		raw, callErr := j.Client.Complete(ctx, j.Prompts.System, rendered)
		if callErr != nil {
			return Verdict{}, fmt.Errorf("judge %s: %w", req.Rubric, callErr)
		}
		v, parseErr := Parse(req.Rubric, raw, rubric.PassMinScore, req.ExpectedFacts)
		if parseErr == nil {
			// The cache is best-effort: a failed write only costs a future re-call.
			_ = j.Cache.Put(cacheNamespace, key, cachedEntry{Raw: raw})
			return v, nil
		}
		lastErr = parseErr
		if !errors.Is(parseErr, ErrMalformed) {
			break
		}
	}
	return Verdict{}, fmt.Errorf("judge %s: %w", req.Rubric, lastErr)
}
