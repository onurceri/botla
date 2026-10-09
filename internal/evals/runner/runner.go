// Package runner orchestrates the eval suites (retrieval, generation,
// calibration), aggregates metrics, applies the regression gate and renders
// JSON + Markdown reports. It depends only on small interfaces so the whole
// flow is unit-testable with fakes; cmd/eval wires in the real clients.
package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/evals/gate"
	"github.com/onurceri/botla-app/internal/evals/harness"
	"github.com/onurceri/botla-app/internal/evals/judge"
	"github.com/onurceri/botla-app/internal/evals/usage"
)

// Suite names.
const (
	SuiteRetrieval   = "retrieval"
	SuiteGeneration  = "generation"
	SuiteCalibration = "calibration"
)

// AllSuites lists the suites in execution order.
var AllSuites = []string{SuiteRetrieval, SuiteGeneration, SuiteCalibration}

// Retriever runs the retrieval stage (implemented by harness.Pipeline).
type Retriever interface {
	Retrieve(ctx context.Context, question string, topN int) (harness.Retrieval, error)
}

// Generator runs one end-to-end chat turn (implemented by harness.Pipeline).
type Generator interface {
	Generate(ctx context.Context, question, sessionID string) (harness.Generation, error)
}

// Grader scores an answer with a rubric (implemented by *judge.Judge).
type Grader interface {
	Evaluate(ctx context.Context, req judge.Request) (judge.Verdict, error)
}

// Config controls a run.
type Config struct {
	Suites      []string
	Cases       []dataset.Case
	Corpus      *dataset.Corpus
	Calibration *dataset.CalibrationSet
	// Ks are the cut-offs for hit rate and recall (default 1, 3, 5).
	Ks []int
	// SearchDepth is how many chunks the raw ranking retrieves (default 10).
	SearchDepth int
	Concurrency int
	// JudgeRepeats is how many times each calibration item is judged, to
	// measure judge self-consistency (default 3; 1 disables the check).
	JudgeRepeats int
	// GenRepeats runs every case's generation (and judging) this many times
	// and reports means over case-repeats plus the per-repeat spread
	// (default 1).
	GenRepeats int
	// VerbosityProbe re-judges each calibration item with claim-free padding
	// added to the answer, to detect verbosity bias.
	VerbosityProbe bool
	// ErrorMessages are product fallback strings that indicate the pipeline
	// failed internally (the product swallows LLM errors into these).
	ErrorMessages []string
	Identity      gate.Identity
	Thresholds    *gate.Thresholds
	Baseline      *gate.Baseline
}

// Deps are the collaborators a run needs. Nil deps disable the suites that need them.
type Deps struct {
	Retriever Retriever
	Generator Generator
	Grader    Grader
}

func (c *Config) defaults() {
	if len(c.Ks) == 0 {
		c.Ks = []int{1, 3, 5}
	}
	if c.SearchDepth <= 0 {
		c.SearchDepth = 10
	}
	if c.Concurrency <= 0 {
		c.Concurrency = 4
	}
	if c.JudgeRepeats <= 0 {
		c.JudgeRepeats = 3
	}
	if c.GenRepeats <= 0 {
		c.GenRepeats = 1
	}
	if len(c.Suites) == 0 {
		c.Suites = AllSuites
	}
}

func (c *Config) has(suite string) bool {
	for _, s := range c.Suites {
		if s == suite {
			return true
		}
	}
	return false
}

// Phase meters, recorded in the report's usage section.
const (
	PhaseRetrieval   = "retrieval"
	PhaseGeneration  = "generation"
	PhaseJudge       = "judge"
	PhaseCalibration = "calibration_judge"
)

// Run executes the configured suites and returns the report. It returns an
// error only for configuration problems; per-case failures are recorded in
// the report (and surface as error-count metrics the gate can fail on).
func Run(ctx context.Context, cfg Config, deps Deps) (*Report, error) {
	cfg.defaults()
	if cfg.has(SuiteRetrieval) && deps.Retriever == nil {
		return nil, errors.New("retrieval suite requires a retriever")
	}
	if cfg.has(SuiteGeneration) && (deps.Generator == nil || deps.Grader == nil) {
		return nil, errors.New("generation suite requires a generator and a grader")
	}
	if cfg.has(SuiteCalibration) && (deps.Grader == nil || cfg.Calibration == nil || cfg.Corpus == nil) {
		return nil, errors.New("calibration suite requires a grader, a calibration set and the corpus")
	}

	rep := &Report{
		StartedAt: time.Now().UTC(),
		Identity:  cfg.Identity,
		Suites:    cfg.Suites,
		Metrics:   map[string]float64{},
		phases:    map[string]*usage.Meter{},
	}
	for _, p := range []string{PhaseRetrieval, PhaseGeneration, PhaseJudge, PhaseCalibration} {
		rep.phases[p] = usage.NewMeter()
	}

	if cfg.has(SuiteRetrieval) {
		rctx := usage.WithMeter(ctx, rep.phases[PhaseRetrieval])
		rep.Retrieval = runRetrieval(rctx, cfg, deps.Retriever)
		rep.Retrieval.addMetrics(rep.Metrics, cfg.Ks)
	}
	if cfg.has(SuiteGeneration) {
		rep.Generation = runGeneration(ctx, cfg, deps, rep.phases[PhaseGeneration], rep.phases[PhaseJudge])
		rep.Generation.addMetrics(rep.Metrics)
	}
	if cfg.has(SuiteCalibration) {
		cctx := usage.WithMeter(ctx, rep.phases[PhaseCalibration])
		rep.Calibration = runCalibration(cctx, cfg, deps.Grader)
		rep.Calibration.addMetrics(rep.Metrics)
	}

	rep.finishUsage()
	rep.FinishedAt = time.Now().UTC()
	rep.DurationSec = rep.FinishedAt.Sub(rep.StartedAt).Seconds()
	g := gate.Evaluate(rep.Metrics, cfg.Thresholds, cfg.Baseline, cfg.Identity)
	rep.Gate = &g
	return rep, nil
}

// forEach runs fn(i) for i in [0, n) on up to workers goroutines, stopping
// early (without starting new items) once ctx is cancelled.
func forEach(ctx context.Context, n, workers int, fn func(i int)) {
	if workers > n {
		workers = n
	}
	var wg sync.WaitGroup
	next := make(chan int)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				fn(i)
			}
		}()
	}
	for i := 0; i < n; i++ {
		if ctx.Err() != nil {
			break
		}
		next <- i
	}
	close(next)
	wg.Wait()
}

func fmtPct(v float64) string { return fmt.Sprintf("%.1f%%", v*100) }
