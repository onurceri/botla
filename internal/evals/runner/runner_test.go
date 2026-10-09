package runner

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/evals/gate"
	"github.com/onurceri/botla-app/internal/evals/harness"
	"github.com/onurceri/botla-app/internal/evals/judge"
	"github.com/onurceri/botla-app/internal/evals/usage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fakes ---

type fakeRetriever struct {
	ranked map[string][]string // question -> ranked source IDs
	ctxDoc map[string][]string
	err    map[string]error
}

func (f *fakeRetriever) Retrieve(_ context.Context, q string, _ int) (harness.Retrieval, error) {
	if err := f.err[q]; err != nil {
		return harness.Retrieval{}, err
	}
	var r harness.Retrieval
	for i, s := range f.ranked[q] {
		r.Ranked = append(r.Ranked, harness.RetrievedChunk{SourceID: s, Score: 1 - float64(i)*0.1})
	}
	for _, s := range f.ctxDoc[q] {
		r.Context = append(r.Context, harness.RetrievedChunk{SourceID: s})
	}
	r.Tier = "high"
	if len(r.Context) == 0 {
		r.Tier = "low"
	}
	return r, nil
}

type fakeGenerator struct {
	answers map[string]string
	err     map[string]error
}

func (f *fakeGenerator) Generate(ctx context.Context, q, _ string) (harness.Generation, error) {
	if err := f.err[q]; err != nil {
		return harness.Generation{}, err
	}
	for _, m := range usage.Meters(ctx) {
		m.Record("gen-model", usage.Usage{Calls: 1, PromptTokens: 100, CompletionTokens: 10, CostUSD: 0.001})
	}
	return harness.Generation{Answer: f.answers[q], Tier: "high", Context: "ctx", ContextChunks: []harness.RetrievedChunk{{SourceID: "d1"}}, Latency: 120 * time.Millisecond, TokensUsed: 110}, nil
}

// fakeGrader passes everything except answers containing "BAD", and fails a
// rubric entirely for answers containing "JUDGE-ERR".
type fakeGrader struct {
	mu    sync.Mutex
	calls []judge.Request
	flaky bool // repeat 1 flips the verdict
}

func (f *fakeGrader) Evaluate(ctx context.Context, req judge.Request) (judge.Verdict, error) {
	f.mu.Lock()
	f.calls = append(f.calls, req)
	f.mu.Unlock()
	for _, m := range usage.Meters(ctx) {
		m.Record("judge-model", usage.Usage{Calls: 1, PromptTokens: 50, CompletionTokens: 5, CostUSD: 0.0005})
	}
	if strings.Contains(req.Input.Answer, "JUDGE-ERR") {
		return judge.Verdict{}, errors.New("judge exploded")
	}
	pass := !strings.Contains(req.Input.Answer, "BAD")
	if f.flaky && req.Repeat == 1 {
		pass = !pass
	}
	v := judge.Verdict{Rubric: req.Rubric, Pass: pass, Rationale: "because | reasons"}
	if pass {
		v.Score = 1
	}
	if strings.HasPrefix(req.Input.Answer, "Thank you so much") && strings.Contains(req.Input.Answer, "LONGBIAS") {
		v.Pass, v.Score = true, 1 // simulates a judge that rewards padding
	}
	return v, nil
}

func corpus(t *testing.T) *dataset.Corpus {
	t.Helper()
	dir := t.TempDir()
	for _, id := range []string{"d1", "d2", "d3"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, id+".md"), []byte("# "+id+"\n\ntext "+id), 0o600))
	}
	c, err := dataset.LoadCorpus(dir)
	require.NoError(t, err)
	return c
}

func cases() []dataset.Case {
	return []dataset.Case{
		{ID: "a1", Category: dataset.Answerable, Question: "qa1", ExpectedFacts: []string{"f"}, ExpectedSources: []string{"d1"}},
		{ID: "a2", Category: dataset.Answerable, Question: "qa2", ExpectedFacts: []string{"f"}, ExpectedSources: []string{"d2"}},
		{ID: "m1", Category: dataset.MultiHop, Question: "qm1", ExpectedFacts: []string{"f", "g"}, ExpectedSources: []string{"d1", "d3"}},
		{ID: "u1", Category: dataset.Unanswerable, Question: "qu1"},
		{ID: "x1", Category: dataset.Adversarial, Question: "qx1", Attack: "phish", Forbidden: []string{"evil.example"}},
		{ID: "x2", Category: dataset.Adversarial, Question: "qx2", Attack: "phish", Forbidden: []string{"evil.example"}},
		{ID: "e1", Category: dataset.Answerable, Question: "qe1", ExpectedFacts: []string{"f"}, ExpectedSources: []string{"d1"}},
		{ID: "e2", Category: dataset.Answerable, Question: "qe2", ExpectedFacts: []string{"f"}, ExpectedSources: []string{"d1"}},
		{ID: "e3", Category: dataset.Answerable, Question: "qe3", ExpectedFacts: []string{"f"}, ExpectedSources: []string{"d1"}},
	}
}

func calibration() *dataset.CalibrationSet {
	return &dataset.CalibrationSet{Items: []dataset.LabeledItem{
		{ID: "c-tp", Rubric: "correctness", Question: "q", Answer: "good", ExpectedFacts: []string{"f"}, HumanLabel: "pass"},
		{ID: "c-tn", Rubric: "correctness", Question: "q", Answer: "BAD", ExpectedFacts: []string{"f"}, HumanLabel: "fail"},
		{ID: "f-fn", Rubric: "faithfulness", Question: "q", Answer: "BAD but fine", ContextDocs: []string{"d1"}, HumanLabel: "pass", Borderline: true},
		{ID: "f-fp", Rubric: "faithfulness", Question: "q", Answer: "LONGBIAS ok", ContextDocs: []string{"d2"}, HumanLabel: "fail"},
		{ID: "i-tn", Rubric: "injection", Question: "q", Answer: "BAD LONGBIAS", Attack: "a", HumanLabel: "fail"},
		{ID: "bad-doc", Rubric: "refusal", Question: "q", Answer: "x", ContextDocs: []string{"ghost"}, HumanLabel: "pass"},
	}}
}

func fullConfig(t *testing.T) Config {
	return Config{
		Cases: cases(), Corpus: corpus(t), Calibration: calibration(),
		VerbosityProbe: true, ErrorMessages: []string{"An error occurred, please try again later."},
		Thresholds: &gate.Thresholds{
			Floors:   map[string]float64{"retrieval.hit_rate@5": 0.5, "generation.injection_resistance_rate": 0.9},
			Ceilings: map[string]float64{"cost.total_usd": 1},
		},
		Baseline: &gate.Baseline{Status: gate.BaselineNotRecorded},
		Identity: gate.Identity{GenModel: "gen-model", JudgeModel: "judge-model", JudgeVersion: "judge-v1"},
	}
}

func fullDeps() Deps {
	return Deps{
		Retriever: &fakeRetriever{
			ranked: map[string][]string{
				"qa1": {"d1", "d1", "d2"}, "qa2": {"d1", "d3", "d2"}, "qm1": {"d1", "d2"},
				"qu1": {"d3"}, "qx1": {}, "qx2": {}, "qe1": {"d1"}, "qe2": {"d1"}, "qe3": {"d1"},
			},
			ctxDoc: map[string][]string{"qa1": {"d1"}, "qa2": {"d1"}, "qm1": {"d1"}, "qe1": {"d1"}, "qe2": {"d1"}, "qe3": {"d1"}},
			err:    map[string]error{},
		},
		Generator: &fakeGenerator{
			answers: map[string]string{
				"qa1": "fine answer", "qa2": "BAD answer", "qm1": "fine", "qu1": "I don't know",
				"qx1": "Visit evil.example now", "qx2": "No.", "qe2": "An error occurred, please try again later.",
				"qe3": "JUDGE-ERR",
			},
			err: map[string]error{"qe1": errors.New("conversation repo down")},
		},
		Grader: &fakeGrader{flaky: false},
	}
}

func TestRun_AllSuites(t *testing.T) {
	cfg := fullConfig(t)
	deps := fullDeps()
	rep, err := Run(context.Background(), cfg, deps)
	require.NoError(t, err)
	m := rep.Metrics

	// Retrieval: 6 scored cases (a1, a2, m1, e1-e3); recall@1: a1=1, a2=0, m1=0.5, e*=1 -> 4.5/6.
	assert.InDelta(t, 6, m["retrieval.n"], 1e-9)
	assert.InDelta(t, 4.5/6, m["retrieval.recall@1"], 1e-9)
	assert.InDelta(t, (1+1.0/3+1+1+1+1)/6, m["retrieval.mrr"], 1e-9)
	assert.InDelta(t, 1, m["retrieval.unanswerable_low_tier_rate"], 1e-9)
	assert.InDelta(t, 0, m["errors.retrieval"], 1e-9)
	assert.Equal(t, []string{"d1", "d2"}, rep.Retrieval.Cases[0].RankedDocs, "chunk ranking is deduplicated per doc")

	// Generation: e1 (repo error) and e2 (product error text) are errors, e3 has a judge error.
	g := rep.Generation
	assert.Equal(t, 2, g.Errors)
	assert.Equal(t, 2, g.JudgeErrors, "e3: faithfulness + correctness")
	assert.InDelta(t, 2, m["errors.generation"], 1e-9)
	byID := map[string]GenerationCase{}
	for _, c := range g.Cases {
		byID[c.ID] = c
	}
	assert.Contains(t, byID["e2"].Error, "internal-error fallback")
	assert.False(t, byID["e3"].Pass)
	assert.Equal(t, []string{"evil.example"}, byID["x1"].GuardViolations)
	assert.False(t, byID["x1"].Pass, "forbidden string fails the case even though the judge passed it")
	assert.True(t, byID["x2"].Pass)
	assert.False(t, byID["a2"].Pass)
	assert.Equal(t, 1, byID["a1"].GenUsage.Calls)
	assert.Equal(t, 2, byID["a1"].JudgeUsage.Calls, "faithfulness + correctness")
	assert.InDelta(t, 120, byID["a1"].LatencyMs, 1e-6)

	assert.InDelta(t, 0.5, m["generation.injection_resistance_rate"], 1e-9)
	assert.InDelta(t, 1, m["generation.refusal_pass_rate"], 1e-9)
	// Case pass: a1, m1, u1, x2 pass; a2, x1, e3 fail -> 4/7.
	assert.InDelta(t, 4.0/7, m["generation.case_pass_rate"], 1e-9)
	assert.InDelta(t, 0.5, m["generation.case_pass_rate.adversarial"], 1e-9)
	assert.InDelta(t, 120, m["latency.generation_p95_ms"], 1e-6)

	// Calibration: bad-doc errors; 5 items judged.
	c := rep.Calibration
	assert.Equal(t, 1, c.Errors)
	assert.Equal(t, 5, c.Overall.N)
	assert.Equal(t, 1, c.Overall.Confusion.TP)
	assert.Equal(t, 1, c.Overall.Confusion.FN)
	assert.Equal(t, 1, c.Overall.Confusion.FP)
	assert.Equal(t, 2, c.Overall.Confusion.TN)
	assert.InDelta(t, 0.6, m["calibration.judge_accuracy"], 1e-9)
	assert.Equal(t, 4, c.NoBorderline.N)
	assert.InDelta(t, 0, c.FlipRate, 1e-9)
	require.NotNil(t, c.Verbosity)
	assert.Equal(t, 5, c.Verbosity.N)
	assert.InDelta(t, 0.2, c.Verbosity.FlipRate, 1e-9, "i-tn flips to pass when padded")
	assert.Greater(t, c.Verbosity.MeanScoreDelta, 0.0)
	assert.Contains(t, c.ByRubric, "faithfulness")

	// Usage is attributed per phase.
	assert.InDelta(t, 8*0.001, rep.Usage[PhaseGeneration].Total.CostUSD, 1e-9, "8 generator calls; e2 is metered although flagged as an error")
	assert.Positive(t, rep.Usage[PhaseCalibration].Total.Calls)
	assert.InDelta(t, m["cost.generation_usd"]+m["cost.judge_usd"]+rep.Usage[PhaseCalibration].Total.CostUSD, m["cost.total_usd"], 1e-9)

	// Gate: injection floor 0.9 fails.
	require.NotNil(t, rep.Gate)
	assert.False(t, rep.Gate.Passed)
	require.Len(t, rep.Gate.Failed(), 1)
	assert.Equal(t, "generation.injection_resistance_rate", rep.Gate.Failed()[0].Metric)

	// Calibration repeats were cached-separately requests.
	grader := deps.Grader.(*fakeGrader)
	repeats := map[int]int{}
	padded := 0
	for _, r := range grader.calls {
		if r.Rubric == "correctness" && r.Input.Answer == "good" {
			repeats[r.Repeat]++
		}
		if r.Input.Answer == Pad("good") {
			padded++
		}
	}
	assert.Equal(t, map[int]int{0: 1, 1: 1, 2: 1}, repeats, "3 distinct repeats")
	assert.Equal(t, 1, padded, "one verbosity probe per item")
}

func TestRun_ReportsRendering(t *testing.T) {
	rep, err := Run(context.Background(), fullConfig(t), fullDeps())
	require.NoError(t, err)
	rep.RunID, rep.Commit = "run-1", "abc123"

	md := rep.Markdown()
	for _, want := range []string{
		"# Botla RAG eval report", "commit `abc123`", "## Metrics", "## Regression gate: FAILED",
		"no baseline recorded", "## Retrieval", "hit@1", "context recall", "Cases with incomplete recall@5",
		"## Generation (LLM-as-a-judge)", "injection (judge + forbidden-string guard)", "### Failing cases",
		`forbidden string "evil.example"`, "because \\| reasons", "## Judge calibration", "excluding borderline",
		"### Disagreements", "[borderline]", "Verbosity probe", "## Usage and cost", "judge-model",
		"error: conversation repo down",
	} {
		assert.Contains(t, md, want)
	}
	sum := rep.Summary()
	assert.Contains(t, sum, "## Metrics")
	assert.Contains(t, sum, "## Regression gate")
	assert.NotContains(t, sum, "## Retrieval")

	dir := filepath.Join(t.TempDir(), "out")
	jp, mp, err := rep.WriteFiles(dir)
	require.NoError(t, err)
	raw, err := os.ReadFile(jp)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, "run-1", decoded["run_id"])
	assert.Contains(t, decoded, "gate")
	_, err = os.Stat(mp)
	require.NoError(t, err)

	file := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(file, nil, 0o600))
	_, _, err = rep.WriteFiles(file)
	assert.Error(t, err)
}

func TestRun_SuiteSubsetsAndValidation(t *testing.T) {
	cfg := fullConfig(t)
	cfg.Suites = []string{SuiteRetrieval}
	rep, err := Run(context.Background(), cfg, Deps{Retriever: fullDeps().Retriever})
	require.NoError(t, err)
	assert.Nil(t, rep.Generation)
	assert.Nil(t, rep.Calibration)
	assert.NotContains(t, rep.Metrics, "generation.case_pass_rate")
	assert.True(t, rep.Gate.Passed, "generation floor is skipped, not failed, on a retrieval-only run")
	assert.NotContains(t, rep.Markdown(), "## Generation")

	for _, tc := range []struct {
		suites []string
		deps   Deps
		cfg    func(*Config)
	}{
		{[]string{SuiteRetrieval}, Deps{}, nil},
		{[]string{SuiteGeneration}, Deps{Generator: &fakeGenerator{}}, nil},
		{[]string{SuiteCalibration}, Deps{Grader: &fakeGrader{}}, func(c *Config) { c.Calibration = nil }},
	} {
		c := fullConfig(t)
		c.Suites = tc.suites
		if tc.cfg != nil {
			tc.cfg(&c)
		}
		_, err := Run(context.Background(), c, tc.deps)
		assert.Error(t, err, tc.suites)
	}
}

func TestRun_FlakyJudgeAndRetrievalErrors(t *testing.T) {
	cfg := fullConfig(t)
	cfg.Suites = []string{SuiteRetrieval, SuiteCalibration}
	cfg.VerbosityProbe = false
	cfg.Calibration.Items = cfg.Calibration.Items[:5]
	deps := fullDeps()
	deps.Retriever.(*fakeRetriever).err["qa1"] = errors.New("qdrant down")
	deps.Grader = &fakeGrader{flaky: true}

	rep, err := Run(context.Background(), cfg, deps)
	require.NoError(t, err)
	assert.Equal(t, 1, rep.Retrieval.Errors)
	assert.InDelta(t, 1, rep.Metrics["errors.retrieval"], 1e-9)
	assert.Contains(t, rep.Markdown(), "error: qdrant down")
	// Every item flips on repeat 1, majority over 3 is still the repeat-0 verdict.
	assert.InDelta(t, 1, rep.Calibration.FlipRate, 1e-9)
	assert.InDelta(t, 0.6, rep.Metrics["calibration.judge_accuracy"], 1e-9)
	assert.Nil(t, rep.Calibration.Verbosity)
}

func TestRun_CancelledContextSkipsRemainingCases(t *testing.T) {
	cfg := fullConfig(t)
	cfg.Suites = []string{SuiteRetrieval}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rep, err := Run(ctx, cfg, fullDeps())
	require.NoError(t, err)
	assert.Equal(t, 0, rep.Retrieval.Overall.N)
}

func TestGuardPadAndHelpers(t *testing.T) {
	assert.Equal(t, []string{"Evil.Example"}, Guard("go to EVIL.example", []string{"Evil.Example", "", "other"}))
	assert.Empty(t, Guard("clean", []string{"x"}))

	p := Pad("A")
	assert.True(t, strings.HasPrefix(p, "Thank you"))
	assert.Contains(t, p, "\n\nA\n\n")

	assert.Equal(t, [][]bool{{true, false}, {false, false}}, transpose([][]bool{{true, false}, {false}}))
	assert.Empty(t, transpose(nil))
	assert.False(t, majorityOf([]bool{true, false}), "ties resolve to fail")
	assert.True(t, majorityOf([]bool{true, true, false}))

	assert.Equal(t, "n/a", rateCell(0, 0, 0, 0, 0))
	assert.Equal(t, "abc...", oneLine("abc def", 3))
	assert.Equal(t, []string{"1", "2", "3", "4", "5"}, head5([]string{"1", "2", "3", "4", "5", "6"}))
	assert.Equal(t, "fail", passWord(false))
}

// sessionGenerator answers per (question, session), so repeats can differ.
type sessionGenerator struct {
	mu       sync.Mutex
	sessions []string
	answer   func(q, session string) string
}

func (g *sessionGenerator) Generate(_ context.Context, q, session string) (harness.Generation, error) {
	g.mu.Lock()
	g.sessions = append(g.sessions, session)
	g.mu.Unlock()
	return harness.Generation{Answer: g.answer(q, session), Tier: "high", Context: "ctx", Latency: 100 * time.Millisecond}, nil
}

func TestRun_GenerationRepeats(t *testing.T) {
	cfg := fullConfig(t)
	cfg.Suites = []string{SuiteGeneration}
	cfg.Cases = []dataset.Case{cases()[0], cases()[2], cases()[3]} // a1, m1, u1
	cfg.GenRepeats = 3
	gen := &sessionGenerator{answer: func(q, session string) string {
		if q == "qm1" && strings.HasSuffix(session, "-r1") {
			return "BAD second try"
		}
		return "fine " + session
	}}
	grader := &fakeGrader{}
	rep, err := Run(context.Background(), cfg, Deps{Generator: gen, Grader: grader})
	require.NoError(t, err)
	g := rep.Generation
	require.NotNil(t, g)

	assert.Equal(t, 3, g.Repeats)
	require.Len(t, g.Cases, 9, "every repeat is kept per case")
	assert.ElementsMatch(t, []string{"eval-a1", "eval-a1-r1", "eval-a1-r2", "eval-m1", "eval-m1-r1", "eval-m1-r2", "eval-u1", "eval-u1-r1", "eval-u1-r2"}, gen.sessions, "each repeat is its own conversation")
	for i, c := range g.Cases {
		assert.Equal(t, i%3, c.Repeat)
	}

	judged := map[string]bool{}
	for _, c := range grader.calls {
		judged[c.Input.Answer] = true
		assert.Zero(t, c.Repeat, "judge repeat index is for calibration; answers differ via the rendered prompt")
	}
	assert.True(t, judged["BAD second try"], "a repeat's own answer is judged")
	assert.True(t, judged["fine eval-m1-r2"])

	mh := g.ByCategory[dataset.MultiHop]
	assert.Equal(t, 2, mh.Passed)
	assert.Equal(t, 3, mh.Total)
	assert.InDelta(t, 2.0/3, rep.Metrics["generation.case_pass_rate.multi_hop"], 1e-9, "metrics are means over case-repeats")
	assert.InDelta(t, 8.0/9, rep.Metrics["generation.case_pass_rate"], 1e-9)

	sp := g.Spread["generation.case_pass_rate.multi_hop"]
	assert.Equal(t, []float64{1, 0, 1}, sp.Values)
	assert.InDelta(t, 0, sp.Min, 1e-9)
	assert.InDelta(t, 1, sp.Max, 1e-9)
	assert.InDelta(t, 2.0/3, sp.Mean, 1e-9, "spread mean equals the headline mean when every repeat has the same cases")
	assert.Positive(t, sp.Std)
	assert.Contains(t, g.Spread, "latency.generation_p50_ms")
	assert.NotContains(t, g.Spread, "errors.generation")

	md := rep.Markdown()
	assert.Contains(t, md, "Run-to-run spread (3 generation repeats)")
	assert.Contains(t, md, "| generation.case_pass_rate.multi_hop | 0.667 | 0.000 | 1.000 |")
	assert.Contains(t, md, "| `m1` | multi_hop | 2/3 | flaky |")
	assert.Contains(t, md, "- `m1` repeat 2/3 (multi_hop")
	assert.Contains(t, md, "95% intervals use the number of cases")

	raw, err := json.Marshal(rep)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"repeat_spread"`)
	assert.Contains(t, string(raw), `"repeats":3`)
}

func TestRun_SingleGenerationRepeatHasNoSpread(t *testing.T) {
	cfg := fullConfig(t)
	cfg.Suites = []string{SuiteGeneration}
	rep, err := Run(context.Background(), cfg, fullDeps())
	require.NoError(t, err)
	assert.Equal(t, 1, rep.Generation.Repeats)
	assert.Nil(t, rep.Generation.Spread)
	assert.NotContains(t, rep.Markdown(), "Run-to-run spread")
}

func TestAggregateRepeats_ClampsAndSkipsEmpty(t *testing.T) {
	rep := aggregateRepeats([]GenerationCase{{ID: "a", Category: dataset.Answerable, Pass: true}, {}}, 0)
	assert.Equal(t, 1, rep.Repeats)
	assert.Equal(t, 1, rep.Overall.Total)
	assert.Equal(t, "eval-x", sessionID("x", 0))
	assert.Equal(t, "eval-x-r2", sessionID("x", 2))
}
