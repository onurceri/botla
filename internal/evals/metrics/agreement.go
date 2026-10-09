package metrics

import (
	"errors"
	"math"
)

// Confusion is a 2x2 confusion matrix for binary pass/fail labels, with the
// human label as ground truth and "pass" as the positive class.
type Confusion struct {
	TP int `json:"both_pass"`             // human pass, judge pass
	FN int `json:"human_pass_judge_fail"` // human pass, judge fail
	FP int `json:"human_fail_judge_pass"` // human fail, judge pass
	TN int `json:"both_fail"`             // human fail, judge fail
}

// ErrLengthMismatch is returned when label slices differ in length.
var ErrLengthMismatch = errors.New("human and judge label slices differ in length")

// NewConfusion builds a confusion matrix from paired labels.
func NewConfusion(human, judge []bool) (Confusion, error) {
	if len(human) != len(judge) {
		return Confusion{}, ErrLengthMismatch
	}
	var c Confusion
	for i := range human {
		c.Add(human[i], judge[i])
	}
	return c, nil
}

// Add records one paired observation.
func (c *Confusion) Add(human, judge bool) {
	switch {
	case human && judge:
		c.TP++
	case human && !judge:
		c.FN++
	case !human && judge:
		c.FP++
	default:
		c.TN++
	}
}

// Merge returns the element-wise sum of two matrices.
func (c Confusion) Merge(o Confusion) Confusion {
	return Confusion{TP: c.TP + o.TP, FN: c.FN + o.FN, FP: c.FP + o.FP, TN: c.TN + o.TN}
}

// N is the number of observations.
func (c Confusion) N() int { return c.TP + c.FN + c.FP + c.TN }

// Accuracy is the observed agreement (TP+TN)/N, or 0 for an empty matrix.
func (c Confusion) Accuracy() float64 {
	n := c.N()
	if n == 0 {
		return 0
	}
	return float64(c.TP+c.TN) / float64(n)
}

// CohensKappa corrects observed agreement for agreement expected by chance:
//
//	kappa = (p_o - p_e) / (1 - p_e)
//
// where p_e is computed from each rater's marginal pass rate. Kappa is 1 for
// perfect agreement, 0 for chance-level agreement and negative for systematic
// disagreement. When p_e == 1 (both raters used a single label for every item)
// kappa is undefined; we return 1 if they agree perfectly and 0 otherwise, and
// callers should report such slices as degenerate.
func (c Confusion) CohensKappa() float64 {
	n := float64(c.N())
	if n == 0 {
		return 0
	}
	po := float64(c.TP+c.TN) / n
	humanPass := float64(c.TP+c.FN) / n
	judgePass := float64(c.TP+c.FP) / n
	pe := humanPass*judgePass + (1-humanPass)*(1-judgePass)
	if math.Abs(1-pe) < 1e-12 {
		if po == 1 {
			return 1
		}
		return 0
	}
	return (po - pe) / (1 - pe)
}

// Degenerate reports whether either rater used only one label, which makes
// kappa uninformative.
func (c Confusion) Degenerate() bool {
	humanPass, humanFail := c.TP+c.FN, c.FP+c.TN
	judgePass, judgeFail := c.TP+c.FP, c.FN+c.TN
	return humanPass == 0 || humanFail == 0 || judgePass == 0 || judgeFail == 0
}

// KappaInterpretation maps kappa to the Landis & Koch (1977) bands. The bands
// are a convention, not a statistical test, but are the usual shorthand.
func KappaInterpretation(k float64) string {
	switch {
	case k < 0:
		return "poor (worse than chance)"
	case k <= 0.20:
		return "slight"
	case k <= 0.40:
		return "fair"
	case k <= 0.60:
		return "moderate"
	case k <= 0.80:
		return "substantial"
	default:
		return "almost perfect"
	}
}

// FlipRate is the fraction of items whose verdict was not unanimous across
// repeated judge runs. runs[r][i] is the verdict of repeat r on item i. Items
// missing from a shorter run are ignored for that run.
func FlipRate(runs [][]bool) float64 {
	if len(runs) < 2 {
		return 0
	}
	n := len(runs[0])
	if n == 0 {
		return 0
	}
	flipped := 0
	for i := 0; i < n; i++ {
		first := runs[0][i]
		for _, r := range runs[1:] {
			if i < len(r) && r[i] != first {
				flipped++
				break
			}
		}
	}
	return float64(flipped) / float64(n)
}
