package metrics

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const eps = 1e-9

func TestDedupeRanked(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c"}, DedupeRanked([]string{"a", "b", "a", "", "c", "b"}))
	assert.Empty(t, DedupeRanked(nil))
}

func TestRetrievalMetrics(t *testing.T) {
	ranked := []string{"x", "a", "y", "b", "z"}
	rel := []string{"a", "b"}

	tests := []struct {
		k           int
		hit, recall float64
	}{
		{0, 0, 0},
		{1, 0, 0},
		{2, 1, 0.5},
		{4, 1, 1},
		{50, 1, 1}, // k beyond list length is clamped
	}
	for _, tt := range tests {
		assert.InDelta(t, tt.hit, HitAtK(ranked, rel, tt.k), eps, "hit@%d", tt.k)
		assert.InDelta(t, tt.recall, RecallAtK(ranked, rel, tt.k), eps, "recall@%d", tt.k)
	}
	assert.InDelta(t, 0, RecallAtK(ranked, rel, -1), eps)
	assert.InDelta(t, 0.5, ReciprocalRank(ranked, rel), eps)
	assert.InDelta(t, 0, ReciprocalRank(ranked, []string{"q"}), eps)
	assert.InDelta(t, 1, ReciprocalRank([]string{"a"}, rel), eps)

	// No relevant docs: recall is defined as 0 and callers must exclude the case.
	assert.InDelta(t, 0, RecallAtK(ranked, nil, 3), eps)
	// A relevant doc ranked twice is counted once.
	assert.InDelta(t, 0.5, RecallAtK([]string{"a", "a"}, rel, 2), eps)
}

func TestCoverageAndScoreRetrieval(t *testing.T) {
	assert.InDelta(t, 0.5, Coverage([]string{"b", "b", "q"}, []string{"a", "b"}), eps)
	assert.InDelta(t, 0, Coverage(nil, []string{"a"}), eps)

	s := ScoreRetrieval([]string{"x", "a", "b"}, []string{"a"}, []string{"a", "b"}, []int{1, 3})
	assert.InDelta(t, 0, s.HitAt[1], eps)
	assert.InDelta(t, 1, s.HitAt[3], eps)
	assert.InDelta(t, 1, s.RecallAt[3], eps)
	assert.InDelta(t, 0.5, s.RR, eps)
	assert.InDelta(t, 0.5, s.ContextRecall, eps)
}

func TestConfusionAndKappa(t *testing.T) {
	// Textbook example: 50 items, raters agree on 20 pass + 15 fail.
	// p_o = 0.7; human pass 25/50, judge pass 30/50 -> p_e = 0.5*0.6+0.5*0.4 = 0.5; kappa = 0.4.
	c := Confusion{TP: 20, FN: 5, FP: 10, TN: 15}
	assert.Equal(t, 50, c.N())
	assert.InDelta(t, 0.7, c.Accuracy(), eps)
	assert.InDelta(t, 0.4, c.CohensKappa(), eps)
	assert.False(t, c.Degenerate())

	human := []bool{true, true, false, false}
	judge := []bool{true, false, true, false}
	built, err := NewConfusion(human, judge)
	require.NoError(t, err)
	assert.Equal(t, Confusion{TP: 1, FN: 1, FP: 1, TN: 1}, built)
	assert.InDelta(t, 0, built.CohensKappa(), eps, "chance-level agreement")

	_, err = NewConfusion([]bool{true}, nil)
	assert.ErrorIs(t, err, ErrLengthMismatch)

	assert.Equal(t, Confusion{TP: 21, FN: 6, FP: 11, TN: 16}, c.Merge(built))
}

func TestKappaEdgeCases(t *testing.T) {
	assert.InDelta(t, 0, Confusion{}.CohensKappa(), eps)
	assert.InDelta(t, 0, Confusion{}.Accuracy(), eps)

	perfect := Confusion{TP: 3, TN: 2}
	assert.InDelta(t, 1, perfect.CohensKappa(), eps)

	// Both raters say "pass" for everything: p_e = 1, kappa undefined -> 1 on full agreement.
	allPass := Confusion{TP: 5}
	assert.InDelta(t, 1, allPass.CohensKappa(), eps)
	assert.True(t, allPass.Degenerate())
	// Raters each use one (different) label: p_e = 0 -> kappa = 0.
	assert.InDelta(t, 0, Confusion{FN: 4}.CohensKappa(), eps)

	// Systematic disagreement is negative.
	assert.Less(t, Confusion{FN: 3, FP: 3}.CohensKappa(), 0.0)
}

func TestKappaInterpretation(t *testing.T) {
	cases := map[float64]string{
		-0.1: "poor (worse than chance)", 0.1: "slight", 0.3: "fair",
		0.5: "moderate", 0.7: "substantial", 0.9: "almost perfect",
	}
	for k, want := range cases {
		assert.Equal(t, want, KappaInterpretation(k), "%v", k)
	}
}

func TestFlipRate(t *testing.T) {
	assert.InDelta(t, 0, FlipRate(nil), eps)
	assert.InDelta(t, 0, FlipRate([][]bool{{true}}), eps)
	assert.InDelta(t, 0, FlipRate([][]bool{{}, {}}), eps)
	runs := [][]bool{
		{true, false, true, false},
		{true, true, true, false},
		{true, false, false},
	}
	// item1 flips (run 2), item2 flips (run 3); item3 missing in run 3 is ignored.
	assert.InDelta(t, 0.5, FlipRate(runs), eps)
}

func TestStats(t *testing.T) {
	assert.InDelta(t, 0, Mean(nil), eps)
	assert.InDelta(t, 2, Mean([]float64{1, 2, 3}), eps)

	xs := []float64{5, 1, 4, 2, 3}
	assert.InDelta(t, 0, Percentile(nil, 50), eps)
	assert.InDelta(t, 1, Percentile(xs, 0), eps)
	assert.InDelta(t, 3, Percentile(xs, 50), eps)
	assert.InDelta(t, 5, Percentile(xs, 95), eps)
	assert.InDelta(t, 5, Percentile(xs, 100), eps)
	assert.InDelta(t, 1, Percentile(xs, 1), eps)
	assert.Equal(t, []float64{5, 1, 4, 2, 3}, xs, "input not mutated")
}

func TestWilson(t *testing.T) {
	lo, hi := Wilson(0, 0, 1.96)
	assert.Zero(t, lo)
	assert.Zero(t, hi)

	// Reference values for 8/10 at 95%: (0.4902, 0.9433).
	lo, hi = Wilson(8, 10, 1.96)
	assert.InDelta(t, 0.4902, lo, 1e-3)
	assert.InDelta(t, 0.9433, hi, 1e-3)

	lo, hi = Wilson(10, 10, 1.96)
	assert.Less(t, lo, 1.0)
	assert.InDelta(t, 1, hi, eps)
	lo, _ = Wilson(0, 10, 1.96)
	assert.InDelta(t, 0, lo, eps)

	r := NewRate(3, 4)
	assert.InDelta(t, 0.75, r.Value, eps)
	assert.True(t, r.CILow < 0.75 && r.CIHigh > 0.75)
	assert.Equal(t, Rate{}, NewRate(0, 0))
	assert.False(t, math.IsNaN(NewRate(0, 3).CILow))
}

func TestNewRepeatedRate(t *testing.T) {
	assert.Equal(t, NewRate(3, 4), NewRepeatedRate(3, 4, 1), "one repeat is a plain rate")
	assert.Equal(t, NewRate(3, 4), NewRepeatedRate(3, 4, 0))

	r := NewRepeatedRate(14, 30, 3) // 10 cases x 3 repeats
	assert.Equal(t, 14, r.Passed)
	assert.Equal(t, 30, r.Total)
	assert.InDelta(t, 14.0/30, r.Value, 1e-12, "value is the mean over case-repeats")
	naive := NewRate(14, 30)
	assert.Less(t, r.CILow, naive.CILow, "the interval uses n = cases, so it is wider than counting repeats as trials")
	assert.Greater(t, r.CIHigh, naive.CIHigh)
	assert.True(t, r.CILow <= r.Value && r.Value <= r.CIHigh)

	assert.Equal(t, Rate{}, NewRepeatedRate(0, 0, 3))
}

func TestNewSpread(t *testing.T) {
	assert.Equal(t, Spread{}, NewSpread(nil))
	one := NewSpread([]float64{0.5})
	assert.InDelta(t, 0.5, one.Mean, 1e-12)
	assert.Zero(t, one.Std)

	s := NewSpread([]float64{0.4, 0.6, 0.5})
	assert.Equal(t, []float64{0.4, 0.6, 0.5}, s.Values)
	assert.InDelta(t, 0.5, s.Mean, 1e-12)
	assert.InDelta(t, 0.4, s.Min, 1e-12)
	assert.InDelta(t, 0.6, s.Max, 1e-12)
	assert.InDelta(t, 0.1, s.Std, 1e-12, "sample standard deviation")
}
