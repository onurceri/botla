// Package gate decides whether an eval run is good enough to merge: absolute
// floors and ceilings (hard quality/cost limits) plus regression checks
// against a recorded baseline with per-metric tolerances.
package gate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Thresholds is the parsed evals/thresholds.yaml.
type Thresholds struct {
	// Regression maps a higher-is-better metric to the largest allowed drop
	// (absolute, in metric units) below its baseline value.
	Regression map[string]float64 `yaml:"regression"`
	// Floors are absolute minimums for higher-is-better metrics.
	Floors map[string]float64 `yaml:"floors"`
	// Ceilings are absolute maximums for lower-is-better metrics (cost, latency).
	Ceilings map[string]float64 `yaml:"ceilings"`
}

// LoadThresholds reads a thresholds file.
func LoadThresholds(path string) (*Thresholds, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- operator-supplied config file
	if err != nil {
		return nil, fmt.Errorf("read thresholds: %w", err)
	}
	var t Thresholds
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&t); err != nil {
		return nil, fmt.Errorf("parse thresholds: %w", err)
	}
	for m, tol := range t.Regression {
		if tol < 0 {
			return nil, fmt.Errorf("thresholds: negative tolerance for %s", m)
		}
	}
	return &t, nil
}

// Baseline status values.
const (
	BaselineRecorded    = "recorded"
	BaselineNotRecorded = "not_recorded"
)

// Identity pins what a set of metrics was measured with. Regression checks
// are only meaningful between runs with the same identity.
type Identity struct {
	CorpusHash   string `json:"corpus_hash"`
	GoldenHash   string `json:"golden_hash"`
	JudgeVersion string `json:"judge_version"`
	JudgeHash    string `json:"judge_prompts_hash"`
	GenModel     string `json:"generation_model"`
	JudgeModel   string `json:"judge_model"`
	EmbedModel   string `json:"embedding_model"`
	// VectorStore is informational; it is deliberately not part of Diff
	// because exact (memory) and Qdrant search agree at eval-corpus size.
	VectorStore string `json:"vector_store"`
	// GenRepeats is how many times each case was generated and judged
	// (-gen-repeats); 0 when the generation suite did not run. It is not part
	// of Diff: a mismatch only makes the generation metrics incomparable, so
	// Evaluate skips just those regression checks.
	GenRepeats int `json:"gen_repeats,omitempty"`
}

// EffectiveGenRepeats is GenRepeats for a run that generated, treating a
// missing value (baselines recorded before repeats existed) as 1.
func (a Identity) EffectiveGenRepeats() int {
	if a.GenRepeats < 1 {
		return 1
	}
	return a.GenRepeats
}

// genMetricPrefixes are the metrics whose value depends on generation repeats.
var genMetricPrefixes = []string{"generation.", "latency.generation"}

func isGenMetric(m string) bool {
	for _, p := range genMetricPrefixes {
		if strings.HasPrefix(m, p) {
			return true
		}
	}
	return false
}

// Diff lists identity fields that differ between a and b, ignoring empty fields.
func (a Identity) Diff(b Identity) []string {
	pairs := []struct{ name, x, y string }{
		{"corpus_hash", a.CorpusHash, b.CorpusHash},
		{"golden_hash", a.GoldenHash, b.GoldenHash},
		{"judge_version", a.JudgeVersion, b.JudgeVersion},
		{"judge_prompts_hash", a.JudgeHash, b.JudgeHash},
		{"generation_model", a.GenModel, b.GenModel},
		{"judge_model", a.JudgeModel, b.JudgeModel},
		{"embedding_model", a.EmbedModel, b.EmbedModel},
	}
	var out []string
	for _, p := range pairs {
		if p.x != "" && p.y != "" && p.x != p.y {
			out = append(out, fmt.Sprintf("%s (%s -> %s)", p.name, p.x, p.y))
		}
	}
	return out
}

// Baseline is the parsed evals/baseline.json.
type Baseline struct {
	Status     string             `json:"status"`
	Note       string             `json:"note,omitempty"`
	RecordedAt string             `json:"recorded_at,omitempty"`
	Commit     string             `json:"commit,omitempty"`
	Identity   Identity           `json:"identity"`
	Metrics    map[string]float64 `json:"metrics"`
}

// LoadBaseline reads a baseline file. A missing file is treated as "not recorded".
func LoadBaseline(path string) (*Baseline, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- operator-supplied config file
	if errors.Is(err, fs.ErrNotExist) {
		return &Baseline{Status: BaselineNotRecorded}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read baseline: %w", err)
	}
	var b Baseline
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("parse baseline: %w", err)
	}
	if b.Status == "" {
		b.Status = BaselineNotRecorded
	}
	return &b, nil
}

// Write stores the baseline as indented JSON.
func (b *Baseline) Write(path string) error {
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("encode baseline: %w", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o600); err != nil {
		return fmt.Errorf("write baseline: %w", err)
	}
	return nil
}

// Check kinds and statuses.
const (
	KindFloor      = "floor"
	KindCeiling    = "ceiling"
	KindRegression = "regression"

	StatusPass    = "pass"
	StatusFail    = "fail"
	StatusSkipped = "skipped"
)

// Check is one gate condition.
type Check struct {
	Metric   string   `json:"metric"`
	Kind     string   `json:"kind"`
	Status   string   `json:"status"`
	Value    *float64 `json:"value,omitempty"`
	Limit    float64  `json:"limit"`
	Baseline *float64 `json:"baseline,omitempty"`
	Note     string   `json:"note,omitempty"`
}

// Result is the overall gate outcome.
type Result struct {
	Passed         bool     `json:"passed"`
	BaselineStatus string   `json:"baseline_status"`
	Checks         []Check  `json:"checks"`
	Warnings       []string `json:"warnings,omitempty"`
}

// Failed returns the failing checks.
func (r Result) Failed() []Check {
	var out []Check
	for _, c := range r.Checks {
		if c.Status == StatusFail {
			out = append(out, c)
		}
	}
	return out
}

func ptr(v float64) *float64 { return &v }

func sortedKeys(m map[string]float64) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// Evaluate applies thresholds and the baseline to the run's metrics.
//
// A metric that the run did not produce (for example, generation metrics on a
// retrieval-only run) is skipped, not failed: the gate judges what was
// measured. Regression checks are skipped with a warning when the baseline
// was recorded with a different corpus, golden set, judge or model, because
// such numbers are not comparable. Generation metrics are additionally
// skipped when the baseline and the run used a different number of
// generation repeats (means over 3 runs and single runs have different noise).
func Evaluate(metrics map[string]float64, th *Thresholds, base *Baseline, current Identity) Result {
	res := Result{Passed: true, BaselineStatus: BaselineNotRecorded}
	if th == nil {
		th = &Thresholds{}
	}
	add := func(c Check) {
		if c.Status == StatusFail {
			res.Passed = false
		}
		res.Checks = append(res.Checks, c)
	}

	for _, m := range sortedKeys(th.Floors) {
		limit := th.Floors[m]
		v, ok := metrics[m]
		if !ok {
			add(Check{Metric: m, Kind: KindFloor, Status: StatusSkipped, Limit: limit, Note: "metric not produced by this run"})
			continue
		}
		st := StatusPass
		if v < limit {
			st = StatusFail
		}
		add(Check{Metric: m, Kind: KindFloor, Status: st, Value: ptr(v), Limit: limit})
	}
	for _, m := range sortedKeys(th.Ceilings) {
		limit := th.Ceilings[m]
		v, ok := metrics[m]
		if !ok {
			add(Check{Metric: m, Kind: KindCeiling, Status: StatusSkipped, Limit: limit, Note: "metric not produced by this run"})
			continue
		}
		st := StatusPass
		if v > limit {
			st = StatusFail
		}
		add(Check{Metric: m, Kind: KindCeiling, Status: st, Value: ptr(v), Limit: limit})
	}

	comparable := base != nil && base.Status == BaselineRecorded
	if base != nil {
		res.BaselineStatus = base.Status
	}
	if !comparable {
		res.Warnings = append(res.Warnings, "no baseline recorded: only absolute floors/ceilings were applied")
	} else if diff := base.Identity.Diff(current); len(diff) > 0 {
		comparable = false
		res.Warnings = append(res.Warnings, fmt.Sprintf("baseline not comparable, regression checks skipped; changed: %v. Re-record the baseline (-update-baseline) in the same PR.", diff))
	}
	genComparable := comparable
	if comparable && current.GenRepeats > 0 && base.Identity.EffectiveGenRepeats() != current.GenRepeats {
		genComparable = false
		res.Warnings = append(res.Warnings, fmt.Sprintf("baseline was recorded with gen_repeats=%d but this run used %d: generation regression checks skipped (run with -gen-repeats %d to compare)",
			base.Identity.EffectiveGenRepeats(), current.GenRepeats, base.Identity.EffectiveGenRepeats()))
	}

	for _, m := range sortedKeys(th.Regression) {
		tol := th.Regression[m]
		v, ok := metrics[m]
		c := Check{Metric: m, Kind: KindRegression}
		switch {
		case !ok:
			c.Status, c.Note = StatusSkipped, "metric not produced by this run"
		case !comparable:
			c.Status, c.Value, c.Note = StatusSkipped, ptr(v), "no comparable baseline"
		case !genComparable && isGenMetric(m):
			c.Status, c.Value, c.Note = StatusSkipped, ptr(v), "different gen_repeats than the baseline"
		default:
			b, inBase := base.Metrics[m]
			if !inBase {
				c.Status, c.Value, c.Note = StatusSkipped, ptr(v), "metric missing from baseline"
				break
			}
			c.Value, c.Baseline, c.Limit = ptr(v), ptr(b), b-tol
			c.Status = StatusPass
			if v < b-tol-1e-9 {
				c.Status = StatusFail
				c.Note = fmt.Sprintf("dropped %.3f (tolerance %.3f)", b-v, tol)
			}
		}
		add(c)
	}
	return res
}
