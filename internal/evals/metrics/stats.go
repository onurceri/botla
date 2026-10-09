package metrics

import (
	"math"
	"sort"
)

// Mean returns the arithmetic mean, or 0 for an empty slice.
func Mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

// Percentile returns the p-th percentile (0..100) using the nearest-rank
// method, or 0 for an empty slice. The input is not modified.
func Percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	if p <= 0 {
		return s[0]
	}
	if p >= 100 {
		return s[len(s)-1]
	}
	rank := int(math.Ceil(p / 100 * float64(len(s))))
	if rank < 1 {
		rank = 1
	}
	return s[rank-1]
}

// Wilson returns the Wilson score interval for a binomial proportion at the
// given z (1.96 for 95%). It behaves well for small n and proportions near 0
// or 1, which is exactly the regime of a 40-case eval set, unlike the normal
// approximation. Returns (0, 0) when n == 0.
func Wilson(successes, n int, z float64) (lo, hi float64) {
	if n <= 0 {
		return 0, 0
	}
	return wilsonP(float64(successes)/float64(n), float64(n), z)
}

// wilsonP is the Wilson interval for a proportion p observed over nf trials.
// p need not be a ratio of integers, which averaged (repeated) pass rates need.
func wilsonP(p, nf, z float64) (lo, hi float64) {
	z2 := z * z
	denom := 1 + z2/nf
	center := (p + z2/(2*nf)) / denom
	half := z * math.Sqrt(p*(1-p)/nf+z2/(4*nf*nf)) / denom
	lo, hi = center-half, center+half
	if lo < 0 {
		lo = 0
	}
	if hi > 1 {
		hi = 1
	}
	return lo, hi
}

// Rate is a pass count with its proportion and 95% Wilson interval.
type Rate struct {
	Passed int     `json:"passed"`
	Total  int     `json:"total"`
	Value  float64 `json:"value"`
	CILow  float64 `json:"ci95_low"`
	CIHigh float64 `json:"ci95_high"`
}

// NewRepeatedRate builds a Rate from counts over case-repeats: each of
// total/repeats cases was run repeats times. Value is the mean pass rate over
// all case-repeats. The confidence interval is computed with n = number of
// cases, not case-repeats, because repeats of one case are strongly correlated
// and counting them as independent trials would overstate precision.
func NewRepeatedRate(passed, total, repeats int) Rate {
	if repeats <= 1 {
		return NewRate(passed, total)
	}
	r := Rate{Passed: passed, Total: total}
	if total > 0 {
		r.Value = float64(passed) / float64(total)
		cases := float64(total) / float64(repeats)
		r.CILow, r.CIHigh = wilsonP(r.Value, cases, 1.96)
	}
	return r
}

// Spread summarizes one metric across repeated runs.
type Spread struct {
	Values []float64 `json:"values"`
	Mean   float64   `json:"mean"`
	Min    float64   `json:"min"`
	Max    float64   `json:"max"`
	// Std is the sample standard deviation (n-1); 0 for fewer than 2 values.
	Std float64 `json:"std"`
}

// NewSpread computes the spread of xs (zero value for an empty slice).
func NewSpread(xs []float64) Spread {
	if len(xs) == 0 {
		return Spread{}
	}
	s := Spread{Values: append([]float64(nil), xs...), Mean: Mean(xs), Min: xs[0], Max: xs[0]}
	var ss float64
	for _, x := range xs {
		s.Min = math.Min(s.Min, x)
		s.Max = math.Max(s.Max, x)
		ss += (x - s.Mean) * (x - s.Mean)
	}
	if len(xs) > 1 {
		s.Std = math.Sqrt(ss / float64(len(xs)-1))
	}
	return s
}

// NewRate builds a Rate from counts.
func NewRate(passed, total int) Rate {
	r := Rate{Passed: passed, Total: total}
	if total > 0 {
		r.Value = float64(passed) / float64(total)
		r.CILow, r.CIHigh = Wilson(passed, total, 1.96)
	}
	return r
}
