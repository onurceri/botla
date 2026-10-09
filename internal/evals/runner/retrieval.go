package runner

import (
	"context"
	"fmt"

	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/evals/metrics"
)

// RetrievalCase is the retrieval outcome for one golden case.
type RetrievalCase struct {
	ID              string                   `json:"id"`
	Category        dataset.Category         `json:"category"`
	ExpectedSources []string                 `json:"expected_sources"`
	RankedDocs      []string                 `json:"ranked_docs"`
	ContextDocs     []string                 `json:"context_docs"`
	Tier            string                   `json:"tier"`
	HighestScore    float64                  `json:"highest_score"`
	Scores          *metrics.RetrievalScores `json:"scores,omitempty"`
	Error           string                   `json:"error,omitempty"`
}

// RetrievalAgg is a mean over cases with a retrieval target.
type RetrievalAgg struct {
	N             int             `json:"n"`
	HitAt         map[int]float64 `json:"hit_rate_at"`
	RecallAt      map[int]float64 `json:"recall_at"`
	MRR           float64         `json:"mrr"`
	ContextRecall float64         `json:"context_recall"`
}

// RetrievalReport is the retrieval suite section of the report.
type RetrievalReport struct {
	Cases      []RetrievalCase                   `json:"cases"`
	Overall    RetrievalAgg                      `json:"overall"`
	ByCategory map[dataset.Category]RetrievalAgg `json:"by_category"`
	// TierCounts counts the product's confidence tier per category.
	TierCounts map[dataset.Category]map[string]int `json:"tier_counts"`
	Errors     int                                 `json:"errors"`
}

func runRetrieval(ctx context.Context, cfg Config, r Retriever) *RetrievalReport {
	out := make([]RetrievalCase, len(cfg.Cases))
	forEach(ctx, len(cfg.Cases), cfg.Concurrency, func(i int) {
		cs := cfg.Cases[i]
		rc := RetrievalCase{ID: cs.ID, Category: cs.Category, ExpectedSources: cs.ExpectedSources}
		res, err := r.Retrieve(ctx, cs.Question, cfg.SearchDepth)
		if err != nil {
			rc.Error = err.Error()
			out[i] = rc
			return
		}
		rc.RankedDocs = metrics.DedupeRanked(res.RankedDocs())
		rc.ContextDocs = metrics.DedupeRanked(res.ContextDocs())
		rc.Tier, rc.HighestScore = res.Tier, res.HighestScore
		if cs.HasRetrievalTarget() {
			s := metrics.ScoreRetrieval(rc.RankedDocs, rc.ContextDocs, cs.ExpectedSources, cfg.Ks)
			rc.Scores = &s
		}
		out[i] = rc
	})
	rep := &RetrievalReport{Cases: out, ByCategory: map[dataset.Category]RetrievalAgg{}, TierCounts: map[dataset.Category]map[string]int{}}
	byCat := map[dataset.Category][]RetrievalCase{}
	var scored []RetrievalCase
	for _, rc := range out {
		if rc.ID == "" { // not reached because the context was cancelled
			continue
		}
		if rc.Error != "" {
			rep.Errors++
			continue
		}
		if rep.TierCounts[rc.Category] == nil {
			rep.TierCounts[rc.Category] = map[string]int{}
		}
		rep.TierCounts[rc.Category][rc.Tier]++
		if rc.Scores != nil {
			scored = append(scored, rc)
			byCat[rc.Category] = append(byCat[rc.Category], rc)
		}
	}
	rep.Overall = aggregateRetrieval(scored, cfg.Ks)
	for cat, cs := range byCat {
		rep.ByCategory[cat] = aggregateRetrieval(cs, cfg.Ks)
	}
	return rep
}

func aggregateRetrieval(cs []RetrievalCase, ks []int) RetrievalAgg {
	agg := RetrievalAgg{N: len(cs), HitAt: map[int]float64{}, RecallAt: map[int]float64{}}
	if len(cs) == 0 {
		return agg
	}
	var rr, ctxRecall []float64
	hits := map[int][]float64{}
	recalls := map[int][]float64{}
	for _, c := range cs {
		rr = append(rr, c.Scores.RR)
		ctxRecall = append(ctxRecall, c.Scores.ContextRecall)
		for _, k := range ks {
			hits[k] = append(hits[k], c.Scores.HitAt[k])
			recalls[k] = append(recalls[k], c.Scores.RecallAt[k])
		}
	}
	agg.MRR = metrics.Mean(rr)
	agg.ContextRecall = metrics.Mean(ctxRecall)
	for _, k := range ks {
		agg.HitAt[k] = metrics.Mean(hits[k])
		agg.RecallAt[k] = metrics.Mean(recalls[k])
	}
	return agg
}

func (r *RetrievalReport) addMetrics(m map[string]float64, ks []int) {
	m["retrieval.n"] = float64(r.Overall.N)
	m["retrieval.mrr"] = r.Overall.MRR
	m["retrieval.context_recall"] = r.Overall.ContextRecall
	for _, k := range ks {
		m[fmt.Sprintf("retrieval.hit_rate@%d", k)] = r.Overall.HitAt[k]
		m[fmt.Sprintf("retrieval.recall@%d", k)] = r.Overall.RecallAt[k]
	}
	m["errors.retrieval"] = float64(r.Errors)
	// How often the product's similarity thresholds route an unanswerable
	// question to the low-confidence fallback (higher is better).
	if tc := r.TierCounts[dataset.Unanswerable]; len(tc) > 0 {
		total := 0
		for _, n := range tc {
			total += n
		}
		m["retrieval.unanswerable_low_tier_rate"] = float64(tc["low"]) / float64(total)
	}
}
