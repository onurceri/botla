package runner

import (
	"context"
	"fmt"
	"sort"

	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/evals/judge"
	"github.com/onurceri/botla-app/internal/evals/metrics"
)

// Claim-free padding used by the verbosity probe. It adds length and
// politeness but no factual content, so no rubric verdict should change.
const (
	verbosityPrefix = "Thank you so much for reaching out, and that is a great question! I'm really glad to help you with this.\n\n"
	verbositySuffix = "\n\nI hope this explanation is helpful and clear. Please don't hesitate to ask if anything else comes up; I'm always happy to help!"
)

// Pad wraps an answer in claim-free padding for the verbosity probe.
func Pad(answer string) string { return verbosityPrefix + answer + verbositySuffix }

// CalibrationItem is the judge-vs-human outcome for one labelled answer.
type CalibrationItem struct {
	ID         string    `json:"id"`
	Rubric     string    `json:"rubric"`
	HumanPass  bool      `json:"human_pass"`
	JudgeRuns  []bool    `json:"judge_runs"`
	Scores     []float64 `json:"scores"`
	JudgePass  bool      `json:"judge_majority_pass"`
	Agree      bool      `json:"agree"`
	Borderline bool      `json:"borderline,omitempty"`
	Rationale  string    `json:"rationale"`
	// PaddedPass / PaddedScore are the verdict on the padded answer (verbosity probe).
	PaddedPass  *bool    `json:"padded_pass,omitempty"`
	PaddedScore *float64 `json:"padded_score,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// Agreement summarizes judge-vs-human agreement on a slice of items.
type Agreement struct {
	N              int               `json:"n"`
	Confusion      metrics.Confusion `json:"confusion"`
	Accuracy       float64           `json:"accuracy"`
	Kappa          float64           `json:"cohens_kappa"`
	Interpretation string            `json:"interpretation"`
	// Degenerate is set when one rater used a single label, making kappa uninformative.
	Degenerate bool `json:"degenerate,omitempty"`
}

func newAgreement(c metrics.Confusion) Agreement {
	k := c.CohensKappa()
	return Agreement{N: c.N(), Confusion: c, Accuracy: c.Accuracy(), Kappa: k, Interpretation: metrics.KappaInterpretation(k), Degenerate: c.Degenerate()}
}

// VerbosityResult summarizes the verbosity-bias probe.
type VerbosityResult struct {
	N int `json:"n"`
	// FlipRate is the fraction of items whose verdict changed when padded.
	FlipRate float64 `json:"flip_rate"`
	// MeanScoreDelta is mean(padded score - original score); positive means the
	// judge rewards length.
	MeanScoreDelta float64 `json:"mean_score_delta"`
}

// CalibrationReport is the judge meta-evaluation section of the report.
type CalibrationReport struct {
	Items        []CalibrationItem    `json:"items"`
	Overall      Agreement            `json:"overall"`
	ByRubric     map[string]Agreement `json:"by_rubric"`
	NoBorderline Agreement            `json:"excluding_borderline"`
	Repeats      int                  `json:"repeats"`
	FlipRate     float64              `json:"repeat_flip_rate"`
	Verbosity    *VerbosityResult     `json:"verbosity,omitempty"`
	Errors       int                  `json:"errors"`
}

func runCalibration(ctx context.Context, cfg Config, g Grader) *CalibrationReport {
	items := cfg.Calibration.Items
	out := make([]CalibrationItem, len(items))
	forEach(ctx, len(items), cfg.Concurrency, func(i int) {
		out[i] = runCalibrationItem(ctx, cfg, g, items[i])
	})
	return aggregateCalibration(out, cfg.JudgeRepeats)
}

func runCalibrationItem(ctx context.Context, cfg Config, g Grader, it dataset.LabeledItem) CalibrationItem {
	ci := CalibrationItem{ID: it.ID, Rubric: it.Rubric, HumanPass: it.HumanPass(), Borderline: it.Borderline}
	contextText, err := cfg.Corpus.Join(it.ContextDocs)
	if err != nil {
		ci.Error = err.Error()
		return ci
	}
	in := judge.Input{Question: it.Question, Context: contextText, Answer: it.Answer, ExpectedFacts: it.ExpectedFacts, Attack: it.Attack}
	for r := 0; r < cfg.JudgeRepeats; r++ {
		v, jerr := g.Evaluate(ctx, judge.Request{Rubric: it.Rubric, Input: in, Repeat: r, ExpectedFacts: len(it.ExpectedFacts)})
		if jerr != nil {
			ci.Error = fmt.Sprintf("repeat %d: %v", r, jerr)
			return ci
		}
		if r == 0 {
			ci.Rationale = v.Rationale
		}
		ci.JudgeRuns = append(ci.JudgeRuns, v.Pass)
		ci.Scores = append(ci.Scores, v.Score)
	}
	ci.JudgePass = majorityOf(ci.JudgeRuns)
	ci.Agree = ci.JudgePass == ci.HumanPass
	if cfg.VerbosityProbe {
		padded := in
		padded.Answer = Pad(it.Answer)
		v, jerr := g.Evaluate(ctx, judge.Request{Rubric: it.Rubric, Input: padded, ExpectedFacts: len(it.ExpectedFacts)})
		if jerr != nil {
			ci.Error = fmt.Sprintf("verbosity probe: %v", jerr)
			return ci
		}
		ci.PaddedPass, ci.PaddedScore = &v.Pass, &v.Score
	}
	return ci
}

// majorityOf is the per-item majority over repeats; ties resolve to fail.
func majorityOf(runs []bool) bool {
	pass := 0
	for _, r := range runs {
		if r {
			pass++
		}
	}
	return pass*2 > len(runs)
}

func aggregateCalibration(items []CalibrationItem, repeats int) *CalibrationReport {
	rep := &CalibrationReport{Items: items, ByRubric: map[string]Agreement{}, Repeats: repeats}
	var overall, noBorder metrics.Confusion
	perRubric := map[string]*metrics.Confusion{}
	var runs [][]bool
	var vN, vFlips int
	var deltas []float64
	for _, it := range items {
		if it.ID == "" {
			continue
		}
		if it.Error != "" {
			rep.Errors++
			continue
		}
		overall.Add(it.HumanPass, it.JudgePass)
		if !it.Borderline {
			noBorder.Add(it.HumanPass, it.JudgePass)
		}
		if perRubric[it.Rubric] == nil {
			perRubric[it.Rubric] = &metrics.Confusion{}
		}
		perRubric[it.Rubric].Add(it.HumanPass, it.JudgePass)
		runs = append(runs, it.JudgeRuns)
		if it.PaddedPass != nil && len(it.JudgeRuns) > 0 {
			vN++
			if *it.PaddedPass != it.JudgeRuns[0] {
				vFlips++
			}
			deltas = append(deltas, *it.PaddedScore-it.Scores[0])
		}
	}
	rep.Overall = newAgreement(overall)
	rep.NoBorderline = newAgreement(noBorder)
	for name, c := range perRubric {
		rep.ByRubric[name] = newAgreement(*c)
	}
	rep.FlipRate = metrics.FlipRate(transpose(runs))
	if vN > 0 {
		rep.Verbosity = &VerbosityResult{N: vN, FlipRate: float64(vFlips) / float64(vN), MeanScoreDelta: metrics.Mean(deltas)}
	}
	return rep
}

// transpose turns per-item run lists into per-run item lists for FlipRate.
func transpose(perItem [][]bool) [][]bool {
	maxRuns := 0
	for _, r := range perItem {
		if len(r) > maxRuns {
			maxRuns = len(r)
		}
	}
	out := make([][]bool, maxRuns)
	for r := 0; r < maxRuns; r++ {
		out[r] = make([]bool, len(perItem))
		for i, item := range perItem {
			if r < len(item) {
				out[r][i] = item[r]
			} else if len(item) > 0 {
				out[r][i] = item[0] // missing repeat: treat as agreeing with the first run
			}
		}
	}
	return out
}

func (c *CalibrationReport) addMetrics(m map[string]float64) {
	m["calibration.n"] = float64(c.Overall.N)
	m["calibration.judge_accuracy"] = c.Overall.Accuracy
	m["calibration.judge_kappa"] = c.Overall.Kappa
	m["calibration.repeat_flip_rate"] = c.FlipRate
	if c.Verbosity != nil {
		m["calibration.verbosity_flip_rate"] = c.Verbosity.FlipRate
	}
	m["errors.calibration"] = float64(c.Errors)
	names := make([]string, 0, len(c.ByRubric))
	for n := range c.ByRubric {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		m["calibration.judge_accuracy."+n] = c.ByRubric[n].Accuracy
	}
}
