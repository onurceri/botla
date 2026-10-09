package gate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkFor(t *testing.T, r Result, kind, metric string) Check {
	t.Helper()
	for _, c := range r.Checks {
		if c.Kind == kind && c.Metric == metric {
			return c
		}
	}
	t.Fatalf("no %s check for %s", kind, metric)
	return Check{}
}

func TestEvaluate_FloorsAndCeilings(t *testing.T) {
	th := &Thresholds{
		Floors:   map[string]float64{"q.a": 0.8, "q.b": 0.8, "q.missing": 0.5},
		Ceilings: map[string]float64{"cost": 1, "lat": 100, "c.missing": 1},
	}
	m := map[string]float64{"q.a": 0.9, "q.b": 0.7, "cost": 0.5, "lat": 150}
	r := Evaluate(m, th, nil, Identity{})

	assert.False(t, r.Passed)
	assert.Equal(t, StatusPass, checkFor(t, r, KindFloor, "q.a").Status)
	assert.Equal(t, StatusFail, checkFor(t, r, KindFloor, "q.b").Status)
	assert.Equal(t, StatusSkipped, checkFor(t, r, KindFloor, "q.missing").Status)
	assert.Equal(t, StatusPass, checkFor(t, r, KindCeiling, "cost").Status)
	assert.Equal(t, StatusFail, checkFor(t, r, KindCeiling, "lat").Status)
	assert.Equal(t, StatusSkipped, checkFor(t, r, KindCeiling, "c.missing").Status)
	assert.Len(t, r.Failed(), 2)
	assert.Equal(t, BaselineNotRecorded, r.BaselineStatus)
	assert.NotEmpty(t, r.Warnings)
}

func TestEvaluate_Regression(t *testing.T) {
	id := Identity{CorpusHash: "c1", JudgeVersion: "judge-v1", GenModel: "m"}
	base := &Baseline{Status: BaselineRecorded, Identity: id, Metrics: map[string]float64{"r.ok": 0.9, "r.drop": 0.9, "r.edge": 0.9}}
	th := &Thresholds{Regression: map[string]float64{"r.ok": 0.05, "r.drop": 0.05, "r.edge": 0.05, "r.notinbase": 0.05, "r.notrun": 0.05}}
	m := map[string]float64{"r.ok": 0.88, "r.drop": 0.80, "r.edge": 0.85, "r.notinbase": 0.5}

	r := Evaluate(m, th, base, id)
	assert.False(t, r.Passed)
	assert.Equal(t, BaselineRecorded, r.BaselineStatus)
	assert.Equal(t, StatusPass, checkFor(t, r, KindRegression, "r.ok").Status)
	drop := checkFor(t, r, KindRegression, "r.drop")
	assert.Equal(t, StatusFail, drop.Status)
	assert.Contains(t, drop.Note, "dropped 0.100")
	assert.InDelta(t, 0.85, drop.Limit, 1e-9)
	assert.Equal(t, StatusPass, checkFor(t, r, KindRegression, "r.edge").Status, "exactly at tolerance passes")
	assert.Equal(t, StatusSkipped, checkFor(t, r, KindRegression, "r.notinbase").Status)
	assert.Equal(t, StatusSkipped, checkFor(t, r, KindRegression, "r.notrun").Status)
	assert.Empty(t, r.Warnings)
}

func TestEvaluate_IncomparableBaselineSkipsRegression(t *testing.T) {
	base := &Baseline{Status: BaselineRecorded, Identity: Identity{JudgeVersion: "judge-v1"}, Metrics: map[string]float64{"r": 0.9}}
	th := &Thresholds{Regression: map[string]float64{"r": 0.01}}
	r := Evaluate(map[string]float64{"r": 0.1}, th, base, Identity{JudgeVersion: "judge-v2"})
	assert.True(t, r.Passed, "a changed judge makes numbers incomparable; floors still apply")
	assert.Equal(t, StatusSkipped, checkFor(t, r, KindRegression, "r").Status)
	require.Len(t, r.Warnings, 1)
	assert.Contains(t, r.Warnings[0], "judge_version (judge-v1 -> judge-v2)")
}

func TestEvaluate_NilThresholds(t *testing.T) {
	r := Evaluate(map[string]float64{"x": 1}, nil, &Baseline{Status: BaselineNotRecorded}, Identity{})
	assert.True(t, r.Passed)
	assert.Empty(t, r.Checks)
}

func TestIdentityDiff(t *testing.T) {
	a := Identity{CorpusHash: "a", GoldenHash: "g", JudgeHash: "h", JudgeModel: "j", EmbedModel: "e", VectorStore: "qdrant"}
	b := Identity{CorpusHash: "b", GoldenHash: "g", JudgeHash: "h2", JudgeModel: "", EmbedModel: "e", VectorStore: "memory"}
	assert.Equal(t, []string{"corpus_hash (a -> b)", "judge_prompts_hash (h -> h2)"}, a.Diff(b))
	assert.Empty(t, a.Diff(a))
}

func TestLoadThresholds(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.yaml")
	require.NoError(t, os.WriteFile(p, []byte("regression: {a: 0.1}\nfloors: {b: 0.5}\nceilings: {c: 2}\n"), 0o600))
	th, err := LoadThresholds(p)
	require.NoError(t, err)
	assert.InDelta(t, 0.1, th.Regression["a"], 1e-12)
	assert.InDelta(t, 0.5, th.Floors["b"], 1e-12)
	assert.InDelta(t, 2, th.Ceilings["c"], 1e-12)

	require.NoError(t, os.WriteFile(p, []byte("flors: {b: 0.5}\n"), 0o600))
	_, err = LoadThresholds(p)
	assert.Error(t, err, "unknown keys are rejected")

	require.NoError(t, os.WriteFile(p, []byte("regression: {a: -1}\n"), 0o600))
	_, err = LoadThresholds(p)
	assert.ErrorContains(t, err, "negative tolerance")

	_, err = LoadThresholds(filepath.Join(dir, "missing.yaml"))
	assert.Error(t, err)

	real, err := LoadThresholds(filepath.Join("..", "..", "..", "evals", "thresholds.yaml"))
	require.NoError(t, err, "committed thresholds file must parse")
	assert.NotEmpty(t, real.Floors)
}

func TestBaselineRoundTrip(t *testing.T) {
	dir := t.TempDir()
	missing, err := LoadBaseline(filepath.Join(dir, "none.json"))
	require.NoError(t, err)
	assert.Equal(t, BaselineNotRecorded, missing.Status)

	p := filepath.Join(dir, "baseline.json")
	b := &Baseline{Status: BaselineRecorded, Commit: "abc", Identity: Identity{GenModel: "m"}, Metrics: map[string]float64{"x": 0.5}}
	require.NoError(t, b.Write(p))
	got, err := LoadBaseline(p)
	require.NoError(t, err)
	assert.Equal(t, b, got)

	require.NoError(t, os.WriteFile(p, []byte(`{"metrics":{}}`), 0o600))
	got, err = LoadBaseline(p)
	require.NoError(t, err)
	assert.Equal(t, BaselineNotRecorded, got.Status, "empty status defaults to not recorded")

	require.NoError(t, os.WriteFile(p, []byte(`{`), 0o600))
	_, err = LoadBaseline(p)
	assert.Error(t, err)

	_, err = LoadBaseline(dir) // a directory: read error other than not-exist
	assert.Error(t, err)

	assert.Error(t, b.Write(filepath.Join(dir, "no", "such", "dir.json")))

	real, err := LoadBaseline(filepath.Join("..", "..", "..", "evals", "baseline.json"))
	require.NoError(t, err, "committed baseline must parse")
	assert.Contains(t, []string{BaselineRecorded, BaselineNotRecorded}, real.Status)
}

func TestEvaluate_GenRepeatsMismatchSkipsOnlyGenerationRegression(t *testing.T) {
	id := Identity{CorpusHash: "c", JudgeVersion: "v1", GenModel: "m"}
	th := &Thresholds{Regression: map[string]float64{
		"generation.faithfulness_pass_rate": 0.01,
		"latency.generation_p95_ms":         0, // only checked as a regression here to prove it is generation-scoped
		"retrieval.recall@5":                0.01,
	}}
	m := map[string]float64{"generation.faithfulness_pass_rate": 0.5, "latency.generation_p95_ms": 1, "retrieval.recall@5": 0.5}

	legacy := &Baseline{Status: BaselineRecorded, Identity: id, Metrics: map[string]float64{
		"generation.faithfulness_pass_rate": 0.9, "latency.generation_p95_ms": 5, "retrieval.recall@5": 0.9,
	}}
	cur := id
	cur.GenRepeats = 3
	r := Evaluate(m, th, legacy, cur)
	assert.Equal(t, StatusSkipped, checkFor(t, r, KindRegression, "generation.faithfulness_pass_rate").Status)
	assert.Equal(t, StatusSkipped, checkFor(t, r, KindRegression, "latency.generation_p95_ms").Status)
	assert.Equal(t, StatusFail, checkFor(t, r, KindRegression, "retrieval.recall@5").Status, "retrieval stays comparable")
	require.Len(t, r.Warnings, 1)
	assert.Contains(t, r.Warnings[0], "gen_repeats=1 but this run used 3")

	cur.GenRepeats = 1
	r = Evaluate(m, th, legacy, cur)
	assert.Equal(t, StatusFail, checkFor(t, r, KindRegression, "generation.faithfulness_pass_rate").Status, "a baseline without gen_repeats counts as 1 repeat")
	assert.Empty(t, r.Warnings)

	same := *legacy
	same.Identity.GenRepeats = 3
	cur.GenRepeats = 3
	r = Evaluate(m, th, &same, cur)
	assert.Equal(t, StatusFail, checkFor(t, r, KindRegression, "generation.faithfulness_pass_rate").Status)
	assert.Empty(t, r.Warnings)

	cur.GenRepeats = 0 // retrieval-only run: nothing generation-related to compare, no warning
	r = Evaluate(map[string]float64{"retrieval.recall@5": 0.95}, th, &same, cur)
	assert.Equal(t, StatusPass, checkFor(t, r, KindRegression, "retrieval.recall@5").Status)
	assert.Empty(t, r.Warnings)
}

func TestIdentity_GenRepeats(t *testing.T) {
	assert.Equal(t, 1, Identity{}.EffectiveGenRepeats())
	assert.Equal(t, 3, Identity{GenRepeats: 3}.EffectiveGenRepeats())
	assert.Empty(t, Identity{GenRepeats: 1}.Diff(Identity{GenRepeats: 3}), "repeats are handled per metric, not by Diff")

	dir := t.TempDir()
	b := &Baseline{Status: BaselineRecorded, Identity: Identity{GenModel: "m", GenRepeats: 3}, Metrics: map[string]float64{}}
	require.NoError(t, b.Write(filepath.Join(dir, "b.json")))
	raw, err := os.ReadFile(filepath.Join(dir, "b.json"))
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"gen_repeats": 3`)
	got, err := LoadBaseline(filepath.Join(dir, "b.json"))
	require.NoError(t, err)
	assert.Equal(t, 3, got.Identity.GenRepeats)
}
