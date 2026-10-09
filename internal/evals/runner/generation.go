package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/evals/judge"
	"github.com/onurceri/botla-app/internal/evals/metrics"
	"github.com/onurceri/botla-app/internal/evals/usage"
)

// GenerationCase is the end-to-end outcome for one golden case.
type GenerationCase struct {
	ID          string                   `json:"id"`
	Category    dataset.Category         `json:"category"`
	Question    string                   `json:"question"`
	Answer      string                   `json:"answer"`
	Tier        string                   `json:"tier"`
	ContextDocs []string                 `json:"context_docs"`
	Verdicts    map[string]judge.Verdict `json:"verdicts"`
	// GuardViolations are forbidden strings found in the answer (deterministic,
	// judge-independent check).
	GuardViolations []string `json:"guard_violations,omitempty"`
	Pass            bool     `json:"pass"`
	LatencyMs       float64  `json:"latency_ms"`
	ProductTokens   int      `json:"product_tokens"`
	// GenUsage and JudgeUsage are metered from provider responses.
	GenUsage    usage.Usage `json:"generation_usage"`
	JudgeUsage  usage.Usage `json:"judge_usage"`
	Error       string      `json:"error,omitempty"`
	JudgeErrors []string    `json:"judge_errors,omitempty"`
}

// RubricAgg summarizes one rubric across cases.
type RubricAgg struct {
	Rate      metrics.Rate `json:"rate"`
	MeanScore float64      `json:"mean_score"`
	// NotApplicable counts verdicts with nothing to grade (they pass).
	NotApplicable int `json:"not_applicable,omitempty"`
}

// GenerationReport is the generation suite section of the report.
type GenerationReport struct {
	Cases       []GenerationCase                  `json:"cases"`
	ByRubric    map[string]RubricAgg              `json:"by_rubric"`
	ByCategory  map[dataset.Category]metrics.Rate `json:"case_pass_by_category"`
	Overall     metrics.Rate                      `json:"case_pass_overall"`
	Injection   metrics.Rate                      `json:"injection_resistance"`
	LatencyP50  float64                           `json:"latency_p50_ms"`
	LatencyP95  float64                           `json:"latency_p95_ms"`
	Errors      int                               `json:"errors"`
	JudgeErrors int                               `json:"judge_errors"`
}

// Guard returns the forbidden substrings present in answer (case-insensitive).
func Guard(answer string, forbidden []string) []string {
	low := strings.ToLower(answer)
	var hits []string
	for _, f := range forbidden {
		if f = strings.TrimSpace(f); f != "" && strings.Contains(low, strings.ToLower(f)) {
			hits = append(hits, f)
		}
	}
	return hits
}

func runGeneration(ctx context.Context, cfg Config, deps Deps, genPhase, judgePhase *usage.Meter) *GenerationReport {
	out := make([]GenerationCase, len(cfg.Cases))
	forEach(ctx, len(cfg.Cases), cfg.Concurrency, func(i int) {
		out[i] = runGenerationCase(ctx, cfg, deps, cfg.Cases[i], genPhase, judgePhase)
	})
	return aggregateGeneration(out)
}

func runGenerationCase(ctx context.Context, cfg Config, deps Deps, cs dataset.Case, genPhase, judgePhase *usage.Meter) GenerationCase {
	gc := GenerationCase{ID: cs.ID, Category: cs.Category, Question: cs.Question, Verdicts: map[string]judge.Verdict{}}

	genMeter := usage.NewMeter()
	gctx := usage.WithMeter(usage.WithMeter(ctx, genPhase), genMeter)
	gen, err := deps.Generator.Generate(gctx, cs.Question, "eval-"+cs.ID)
	gc.GenUsage = genMeter.Total()
	if err != nil {
		gc.Error = err.Error()
		return gc
	}
	gc.Answer, gc.Tier = gen.Answer, gen.Tier
	gc.ContextDocs = metrics.DedupeRanked(gen.ContextDocs())
	gc.LatencyMs = float64(gen.Latency.Microseconds()) / 1000
	gc.ProductTokens = gen.TokensUsed
	for _, msg := range cfg.ErrorMessages {
		if msg != "" && strings.TrimSpace(gen.Answer) == strings.TrimSpace(msg) {
			gc.Error = "product returned its internal-error fallback message (LLM call failed; check provider keys and logs)"
			return gc
		}
	}

	gc.GuardViolations = Guard(gen.Answer, cs.Forbidden)

	judgeMeter := usage.NewMeter()
	jctx := usage.WithMeter(usage.WithMeter(ctx, judgePhase), judgeMeter)
	in := judge.Input{
		Question:      cs.Question,
		Context:       gen.Context,
		Answer:        gen.Answer,
		ExpectedFacts: cs.ExpectedFacts,
		Attack:        cs.Attack,
	}
	pass := len(gc.GuardViolations) == 0
	for _, rubric := range cs.Rubrics() {
		v, jerr := deps.Grader.Evaluate(jctx, judge.Request{Rubric: rubric, Input: in, ExpectedFacts: len(cs.ExpectedFacts)})
		if jerr != nil {
			gc.JudgeErrors = append(gc.JudgeErrors, fmt.Sprintf("%s: %v", rubric, jerr))
			pass = false
			continue
		}
		gc.Verdicts[rubric] = v
		pass = pass && v.Pass
	}
	gc.JudgeUsage = judgeMeter.Total()
	gc.Pass = pass
	return gc
}

func aggregateGeneration(cases []GenerationCase) *GenerationReport {
	rep := &GenerationReport{Cases: cases, ByRubric: map[string]RubricAgg{}, ByCategory: map[dataset.Category]metrics.Rate{}}
	type tally struct {
		pass, total, na int
		scores          []float64
	}
	rubrics := map[string]*tally{}
	catPass := map[dataset.Category][2]int{}
	var overallPass, overallTotal, injPass, injTotal int
	var latencies []float64

	for _, c := range cases {
		if c.ID == "" {
			continue
		}
		if c.Error != "" {
			rep.Errors++
			continue
		}
		if len(c.JudgeErrors) > 0 {
			rep.JudgeErrors += len(c.JudgeErrors)
		}
		latencies = append(latencies, c.LatencyMs)
		for name, v := range c.Verdicts {
			t := rubrics[name]
			if t == nil {
				t = &tally{}
				rubrics[name] = t
			}
			t.total++
			t.scores = append(t.scores, v.Score)
			if v.Pass {
				t.pass++
			}
			if v.NotApplicable {
				t.na++
			}
		}
		cp := catPass[c.Category]
		cp[1]++
		overallTotal++
		if c.Pass {
			cp[0]++
			overallPass++
		}
		catPass[c.Category] = cp
		if c.Category == dataset.Adversarial {
			injTotal++
			if v, ok := c.Verdicts[judge.Injection]; ok && v.Pass && len(c.GuardViolations) == 0 {
				injPass++
			}
		}
	}
	for name, t := range rubrics {
		rep.ByRubric[name] = RubricAgg{Rate: metrics.NewRate(t.pass, t.total), MeanScore: metrics.Mean(t.scores), NotApplicable: t.na}
	}
	for cat, cp := range catPass {
		rep.ByCategory[cat] = metrics.NewRate(cp[0], cp[1])
	}
	rep.Overall = metrics.NewRate(overallPass, overallTotal)
	rep.Injection = metrics.NewRate(injPass, injTotal)
	rep.LatencyP50 = metrics.Percentile(latencies, 50)
	rep.LatencyP95 = metrics.Percentile(latencies, 95)
	return rep
}

func (g *GenerationReport) addMetrics(m map[string]float64) {
	if a, ok := g.ByRubric[judge.Faithfulness]; ok {
		m["generation.faithfulness_pass_rate"] = a.Rate.Value
		m["generation.faithfulness_mean"] = a.MeanScore
	}
	if a, ok := g.ByRubric[judge.Correctness]; ok {
		m["generation.correctness_pass_rate"] = a.Rate.Value
		m["generation.correctness_mean"] = a.MeanScore
	}
	if a, ok := g.ByRubric[judge.Refusal]; ok {
		m["generation.refusal_pass_rate"] = a.Rate.Value
	}
	if g.Injection.Total > 0 {
		m["generation.injection_resistance_rate"] = g.Injection.Value
	}
	m["generation.case_pass_rate"] = g.Overall.Value
	for cat, r := range g.ByCategory {
		m["generation.case_pass_rate."+string(cat)] = r.Value
	}
	m["latency.generation_p50_ms"] = g.LatencyP50
	m["latency.generation_p95_ms"] = g.LatencyP95
	m["errors.generation"] = float64(g.Errors)
	m["errors.judge"] = float64(g.JudgeErrors)
}
