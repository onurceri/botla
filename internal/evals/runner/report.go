package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/evals/gate"
	"github.com/onurceri/botla-app/internal/evals/judge"
	"github.com/onurceri/botla-app/internal/evals/usage"
)

// PhaseUsage is the metered usage of one phase.
type PhaseUsage struct {
	Total   usage.Usage            `json:"total"`
	ByModel map[string]usage.Usage `json:"by_model"`
}

// Report is the complete, machine-readable outcome of a run.
type Report struct {
	RunID       string                `json:"run_id"`
	Commit      string                `json:"commit,omitempty"`
	StartedAt   time.Time             `json:"started_at"`
	FinishedAt  time.Time             `json:"finished_at"`
	DurationSec float64               `json:"duration_sec"`
	Identity    gate.Identity         `json:"identity"`
	Suites      []string              `json:"suites"`
	Metrics     map[string]float64    `json:"metrics"`
	Retrieval   *RetrievalReport      `json:"retrieval,omitempty"`
	Generation  *GenerationReport     `json:"generation,omitempty"`
	Calibration *CalibrationReport    `json:"calibration,omitempty"`
	Usage       map[string]PhaseUsage `json:"usage"`
	Gate        *gate.Result          `json:"gate"`

	phases map[string]*usage.Meter
}

func (r *Report) finishUsage() {
	r.Usage = map[string]PhaseUsage{}
	var total usage.Usage
	for name, m := range r.phases {
		t := m.Total()
		if t.Calls == 0 {
			continue
		}
		r.Usage[name] = PhaseUsage{Total: t, ByModel: m.ByModel()}
		total = total.Add(t)
	}
	r.Metrics["cost.total_usd"] = total.CostUSD
	r.Metrics["tokens.total"] = float64(total.TotalTokens())
	if p, ok := r.Usage[PhaseGeneration]; ok {
		r.Metrics["cost.generation_usd"] = p.Total.CostUSD
	}
	if p, ok := r.Usage[PhaseJudge]; ok {
		r.Metrics["cost.judge_usd"] = p.Total.CostUSD
	}
}

// WriteFiles writes report.json and report.md into dir.
func (r *Report) WriteFiles(dir string) (jsonPath, mdPath string, err error) {
	if mkErr := os.MkdirAll(dir, 0o750); mkErr != nil {
		return "", "", fmt.Errorf("create report dir: %w", mkErr)
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("encode report: %w", err)
	}
	jsonPath = filepath.Join(dir, "report.json")
	if err = os.WriteFile(jsonPath, append(raw, '\n'), 0o600); err != nil {
		return "", "", fmt.Errorf("write report.json: %w", err)
	}
	mdPath = filepath.Join(dir, "report.md")
	if err = os.WriteFile(mdPath, []byte(r.Markdown()), 0o600); err != nil {
		return "", "", fmt.Errorf("write report.md: %w", err)
	}
	return jsonPath, mdPath, nil
}

func f3(v float64) string { return fmt.Sprintf("%.3f", v) }

func rateCell(passed, total int, value, lo, hi float64) string {
	if total == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%s (%d/%d, 95%% CI %s-%s)", fmtPct(value), passed, total, fmtPct(lo), fmtPct(hi))
}

func oneLine(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ReplaceAll(s, "|", "\\|")
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}

// Summary is the short text printed to stdout at the end of a run.
func (r *Report) Summary() string {
	var b strings.Builder
	r.writeHeadline(&b)
	r.writeGate(&b)
	return b.String()
}

// Markdown renders the full human-readable report.
func (r *Report) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Botla RAG eval report\n\n")
	fmt.Fprintf(&b, "- Run: `%s`", r.RunID)
	if r.Commit != "" {
		fmt.Fprintf(&b, " at commit `%s`", r.Commit)
	}
	fmt.Fprintf(&b, "\n- Started: %s (%.0fs)\n", r.StartedAt.Format(time.RFC3339), r.DurationSec)
	id := r.Identity
	fmt.Fprintf(&b, "- Generation model: `%s` | Judge model: `%s` | Embeddings: `%s` | Vector store: `%s`\n", id.GenModel, id.JudgeModel, id.EmbedModel, id.VectorStore)
	fmt.Fprintf(&b, "- Judge prompts: `%s` (hash `%s`) | Corpus `%s` | Golden set `%s`", id.JudgeVersion, id.JudgeHash, id.CorpusHash, id.GoldenHash)
	if id.GenRepeats > 0 {
		fmt.Fprintf(&b, " | Generation repeats: %d", id.GenRepeats)
	}
	b.WriteString("\n\n")

	r.writeHeadline(&b)
	r.writeGate(&b)
	r.writeRetrieval(&b)
	r.writeGeneration(&b)
	r.writeCalibration(&b)
	r.writeUsage(&b)
	return b.String()
}

func (r *Report) writeHeadline(b *strings.Builder) {
	keys := make([]string, 0, len(r.Metrics))
	for k := range r.Metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintf(b, "## Metrics\n\n| Metric | Value |\n|---|---|\n")
	for _, k := range keys {
		fmt.Fprintf(b, "| %s | %s |\n", k, f3(r.Metrics[k]))
	}
	b.WriteString("\n")
}

func (r *Report) writeGate(b *strings.Builder) {
	if r.Gate == nil {
		return
	}
	status := "PASSED"
	if !r.Gate.Passed {
		status = "FAILED"
	}
	fmt.Fprintf(b, "## Regression gate: %s\n\nBaseline: %s\n\n", status, r.Gate.BaselineStatus)
	for _, w := range r.Gate.Warnings {
		fmt.Fprintf(b, "> %s\n\n", w)
	}
	fmt.Fprintf(b, "| Check | Metric | Value | Limit | Baseline | Status |\n|---|---|---|---|---|---|\n")
	for _, c := range r.Gate.Checks {
		val, base, limit := "-", "-", f3(c.Limit)
		if c.Kind == gate.KindRegression && c.Baseline == nil {
			limit = "-"
		}
		if c.Value != nil {
			val = f3(*c.Value)
		}
		if c.Baseline != nil {
			base = f3(*c.Baseline)
		}
		st := c.Status
		if c.Note != "" {
			st += " (" + c.Note + ")"
		}
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s |\n", c.Kind, c.Metric, val, limit, base, st)
	}
	b.WriteString("\n")
}

func (r *Report) writeRetrieval(b *strings.Builder) {
	rr := r.Retrieval
	if rr == nil {
		return
	}
	ks := make([]int, 0, len(rr.Overall.HitAt))
	for k := range rr.Overall.HitAt {
		ks = append(ks, k)
	}
	sort.Ints(ks)
	fmt.Fprintf(b, "## Retrieval\n\nDocument-level metrics over cases with expected sources; `context recall` = expected docs that reached the LLM prompt after the product's tiered thresholds and token budget.\n\n")
	head := "| Slice | n |"
	sep := "|---|---|"
	for _, k := range ks {
		head += fmt.Sprintf(" hit@%d | recall@%d |", k, k)
		sep += "---|---|"
	}
	fmt.Fprintf(b, "%s MRR | context recall |\n%s---|---|\n", head, sep)
	row := func(name string, a RetrievalAgg) {
		fmt.Fprintf(b, "| %s | %d |", name, a.N)
		for _, k := range ks {
			fmt.Fprintf(b, " %s | %s |", f3(a.HitAt[k]), f3(a.RecallAt[k]))
		}
		fmt.Fprintf(b, " %s | %s |\n", f3(a.MRR), f3(a.ContextRecall))
	}
	row("overall", rr.Overall)
	for _, cat := range dataset.Categories {
		if a, ok := rr.ByCategory[cat]; ok {
			row(string(cat), a)
		}
	}
	fmt.Fprintf(b, "\nConfidence tier by category: ")
	var parts []string
	for _, cat := range dataset.Categories {
		if tc, ok := rr.TierCounts[cat]; ok {
			parts = append(parts, fmt.Sprintf("%s high=%d medium=%d low=%d", cat, tc["high"], tc["medium"], tc["low"]))
		}
	}
	fmt.Fprintf(b, "%s\n\n", strings.Join(parts, "; "))
	var misses []string
	for _, c := range rr.Cases {
		if c.Error != "" {
			misses = append(misses, fmt.Sprintf("- `%s`: error: %s", c.ID, c.Error))
			continue
		}
		if c.Scores != nil && c.Scores.RecallAt[ks[len(ks)-1]] < 1 {
			misses = append(misses, fmt.Sprintf("- `%s`: expected %v, ranked %v, in context %v", c.ID, c.ExpectedSources, head5(c.RankedDocs), c.ContextDocs))
		}
	}
	if len(misses) > 0 {
		fmt.Fprintf(b, "Cases with incomplete recall@%d:\n\n%s\n\n", ks[len(ks)-1], strings.Join(misses, "\n"))
	}
}

func head5(xs []string) []string {
	if len(xs) > 5 {
		return xs[:5]
	}
	return xs
}

func (r *Report) writeGeneration(b *strings.Builder) {
	g := r.Generation
	if g == nil {
		return
	}
	fmt.Fprintf(b, "## Generation (LLM-as-a-judge)\n\n")
	if g.Repeats > 1 {
		fmt.Fprintf(b, "Each case was generated and judged %d times. Rates and means are over case-repeats (counts below are case-repeats); 95%% intervals use the number of cases, because repeats of one case are correlated.\n\n", g.Repeats)
	}
	fmt.Fprintf(b, "| Rubric | Pass rate | Mean score |\n|---|---|---|\n")
	for _, name := range []string{judge.Faithfulness, judge.Correctness, judge.Refusal, judge.Injection} {
		a, ok := g.ByRubric[name]
		if !ok {
			continue
		}
		fmt.Fprintf(b, "| %s | %s | %s |\n", name, rateCell(a.Rate.Passed, a.Rate.Total, a.Rate.Value, a.Rate.CILow, a.Rate.CIHigh), f3(a.MeanScore))
	}
	in := g.Injection
	fmt.Fprintf(b, "| injection (judge + forbidden-string guard) | %s | - |\n\n", rateCell(in.Passed, in.Total, in.Value, in.CILow, in.CIHigh))
	fmt.Fprintf(b, "| Category | Case pass rate (all rubrics + guards) |\n|---|---|\n")
	for _, cat := range dataset.Categories {
		if rt, ok := g.ByCategory[cat]; ok {
			fmt.Fprintf(b, "| %s | %s |\n", cat, rateCell(rt.Passed, rt.Total, rt.Value, rt.CILow, rt.CIHigh))
		}
	}
	o := g.Overall
	fmt.Fprintf(b, "| **overall** | %s |\n\n", rateCell(o.Passed, o.Total, o.Value, o.CILow, o.CIHigh))
	fmt.Fprintf(b, "Latency per case (full ProcessChat turn): p50 %.0f ms, p95 %.0f ms. Errors: generation %d, judge %d.\n\n", g.LatencyP50, g.LatencyP95, g.Errors, g.JudgeErrors)

	writeRepeatSpread(b, g)

	var fails []string
	for _, c := range g.Cases {
		if c.ID == "" || c.Pass {
			continue
		}
		label := "`" + c.ID + "`"
		if g.Repeats > 1 {
			label += fmt.Sprintf(" repeat %d/%d", c.Repeat+1, g.Repeats)
		}
		var why []string
		if c.Error != "" {
			why = append(why, "error: "+c.Error)
		}
		why = append(why, c.JudgeErrors...)
		for _, v := range c.GuardViolations {
			why = append(why, fmt.Sprintf("forbidden string %q", v))
		}
		names := make([]string, 0, len(c.Verdicts))
		for n := range c.Verdicts {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			if v := c.Verdicts[n]; !v.Pass {
				why = append(why, fmt.Sprintf("%s %.2f: %s", n, v.Score, oneLine(v.Rationale, 220)))
			}
		}
		fails = append(fails, fmt.Sprintf("- %s (%s, tier %s)\n  - Answer: %s\n  - %s", label, c.Category, c.Tier, oneLine(c.Answer, 260), strings.Join(why, "\n  - ")))
	}
	if len(fails) > 0 {
		fmt.Fprintf(b, "### Failing cases\n\n%s\n\n", strings.Join(fails, "\n"))
	}
}

// writeRepeatSpread renders, for runs with several generation repeats, the
// per-repeat spread of each generation metric and the cases whose outcome
// differed between repeats.
func writeRepeatSpread(b *strings.Builder, g *GenerationReport) {
	if g.Repeats <= 1 {
		return
	}
	keys := make([]string, 0, len(g.Spread))
	for k := range g.Spread {
		if strings.HasPrefix(k, "errors.") {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintf(b, "### Run-to-run spread (%d generation repeats)\n\nEach metric computed separately for every repeat. The headline value is the mean.\n\n| Metric | Mean | Min | Max | Std | Per repeat |\n|---|---|---|---|---|---|\n", g.Repeats)
	for _, k := range keys {
		sp := g.Spread[k]
		vals := make([]string, len(sp.Values))
		for i, v := range sp.Values {
			vals[i] = f3(v)
		}
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s |\n", k, f3(sp.Mean), f3(sp.Min), f3(sp.Max), f3(sp.Std), strings.Join(vals, ", "))
	}
	b.WriteString("\n")

	type tally struct {
		cat          dataset.Category
		passed, runs int
	}
	byCase := map[string]*tally{}
	var order []string
	for _, c := range g.Cases {
		if c.ID == "" {
			continue
		}
		t := byCase[c.ID]
		if t == nil {
			t = &tally{cat: c.Category}
			byCase[c.ID] = t
			order = append(order, c.ID)
		}
		t.runs++
		if c.Pass {
			t.passed++
		}
	}
	var rows []string
	for _, id := range order {
		if t := byCase[id]; t.passed < t.runs {
			kind := "always fails"
			if t.passed > 0 {
				kind = "flaky"
			}
			rows = append(rows, fmt.Sprintf("| `%s` | %s | %d/%d | %s |", id, t.cat, t.passed, t.runs, kind))
		}
	}
	if len(rows) > 0 {
		fmt.Fprintf(b, "### Cases failing in at least one repeat\n\n| Case | Category | Passed | |\n|---|---|---|---|\n%s\n\n", strings.Join(rows, "\n"))
	}
}

func (r *Report) writeCalibration(b *strings.Builder) {
	c := r.Calibration
	if c == nil {
		return
	}
	fmt.Fprintf(b, "## Judge calibration (meta-eval vs. human labels)\n\n")
	fmt.Fprintf(b, "Judge verdict = majority over %d repeat(s). Positive class = pass.\n\n", c.Repeats)
	fmt.Fprintf(b, "| Slice | n | Accuracy | Cohen's kappa | Interpretation |\n|---|---|---|---|---|\n")
	row := func(name string, a Agreement) {
		interp := a.Interpretation
		if a.Degenerate {
			interp = "undefined (a rater used a single label on this slice)"
		}
		fmt.Fprintf(b, "| %s | %d | %s | %s | %s |\n", name, a.N, f3(a.Accuracy), f3(a.Kappa), interp)
	}
	row("overall", c.Overall)
	row("excluding borderline", c.NoBorderline)
	names := make([]string, 0, len(c.ByRubric))
	for n := range c.ByRubric {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		row(n, c.ByRubric[n])
	}
	cm := c.Overall.Confusion
	fmt.Fprintf(b, "\nConfusion (human x judge): both pass %d, human pass/judge fail %d, human fail/judge pass %d, both fail %d.\n\n", cm.TP, cm.FN, cm.FP, cm.TN)
	fmt.Fprintf(b, "Repeat consistency: %s of items changed verdict across %d runs.\n", fmtPct(c.FlipRate), c.Repeats)
	if v := c.Verbosity; v != nil {
		fmt.Fprintf(b, "Verbosity probe: %s of %d items changed verdict when claim-free padding was added; mean score delta %+.3f.\n", fmtPct(v.FlipRate), v.N, v.MeanScoreDelta)
	}
	var dis []string
	for _, it := range c.Items {
		if it.Error != "" {
			dis = append(dis, fmt.Sprintf("- `%s`: error: %s", it.ID, it.Error))
		} else if it.ID != "" && !it.Agree {
			dis = append(dis, fmt.Sprintf("- `%s` (%s): human %s, judge %s%s. Judge rationale: %s", it.ID, it.Rubric, passWord(it.HumanPass), passWord(it.JudgePass), borderTag(it.Borderline), oneLine(it.Rationale, 240)))
		}
	}
	if len(dis) > 0 {
		fmt.Fprintf(b, "\n### Disagreements\n\n%s\n", strings.Join(dis, "\n"))
	}
	b.WriteString("\n")
}

func passWord(p bool) string {
	if p {
		return "pass"
	}
	return "fail"
}

func borderTag(b bool) string {
	if b {
		return " [borderline]"
	}
	return ""
}

func (r *Report) writeUsage(b *strings.Builder) {
	if len(r.Usage) == 0 {
		return
	}
	fmt.Fprintf(b, "## Usage and cost\n\n| Phase | Model | Calls | Prompt tokens | Completion tokens | Cost (USD) |\n|---|---|---|---|---|---|\n")
	phases := make([]string, 0, len(r.Usage))
	for p := range r.Usage {
		phases = append(phases, p)
	}
	sort.Strings(phases)
	for _, p := range phases {
		models := make([]string, 0, len(r.Usage[p].ByModel))
		for m := range r.Usage[p].ByModel {
			models = append(models, m)
		}
		sort.Strings(models)
		for _, m := range models {
			u := r.Usage[p].ByModel[m]
			cost := fmt.Sprintf("%.4f", u.CostUSD)
			if u.UnpricedCalls > 0 {
				cost += fmt.Sprintf(" (+%d unpriced calls)", u.UnpricedCalls)
			}
			fmt.Fprintf(b, "| %s | %s | %d | %d | %d | %s |\n", p, m, u.Calls, u.PromptTokens, u.CompletionTokens, cost)
		}
	}
	fmt.Fprintf(b, "\nTotal metered cost: $%.4f. Cache hits (embeddings, capability summaries, judge verdicts) are free and not counted.\n", r.Metrics["cost.total_usd"])
}
