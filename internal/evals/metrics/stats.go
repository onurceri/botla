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
	nf := float64(n)
	p := float64(successes) / nf
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

// NewRate builds a Rate from counts.
func NewRate(passed, total int) Rate {
	r := Rate{Passed: passed, Total: total}
	if total > 0 {
		r.Value = float64(passed) / float64(total)
		r.CILow, r.CIHigh = Wilson(passed, total, 1.96)
	}
	return r
}
