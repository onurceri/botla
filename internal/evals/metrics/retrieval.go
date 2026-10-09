// Package metrics implements the deterministic (no-LLM) measurements used by
// the eval suite: rank-based retrieval metrics, inter-rater agreement for judge
// calibration, and small summary statistics.
package metrics

// DedupeRanked collapses a chunk-level ranking of document IDs into a
// document-level ranking, keeping the first (best) position of each document.
// Retrieval metrics are computed per document because golden cases reference
// source documents, and one document can produce several chunks.
func DedupeRanked(ids []string) []string {
	seen := make(map[string]bool, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

func relevantSet(relevant []string) map[string]bool {
	m := make(map[string]bool, len(relevant))
	for _, r := range relevant {
		m[r] = true
	}
	return m
}

func topK(ranked []string, k int) []string {
	if k < 0 {
		k = 0
	}
	if k > len(ranked) {
		k = len(ranked)
	}
	return ranked[:k]
}

// RecallAtK is the fraction of relevant documents that appear in the first k
// ranked documents. It returns 0 when there are no relevant documents (callers
// should exclude such cases rather than average a meaningless value).
func RecallAtK(ranked, relevant []string, k int) float64 {
	rel := relevantSet(relevant)
	if len(rel) == 0 {
		return 0
	}
	found := 0
	for _, id := range topK(ranked, k) {
		if rel[id] {
			found++
			delete(rel, id) // count each relevant doc once even if ranked twice
		}
	}
	return float64(found) / float64(len(relevantSet(relevant)))
}

// HitAtK is 1 if at least one relevant document is in the first k, else 0.
func HitAtK(ranked, relevant []string, k int) float64 {
	rel := relevantSet(relevant)
	for _, id := range topK(ranked, k) {
		if rel[id] {
			return 1
		}
	}
	return 0
}

// ReciprocalRank is 1/rank of the first relevant document (1-based), or 0 if
// none is ranked. Averaged over cases it gives MRR.
func ReciprocalRank(ranked, relevant []string) float64 {
	rel := relevantSet(relevant)
	for i, id := range ranked {
		if rel[id] {
			return 1 / float64(i+1)
		}
	}
	return 0
}

// Coverage is the fraction of relevant documents present anywhere in docs
// (order-insensitive). Used for "context recall": did the evidence actually
// reach the LLM after the product's score thresholds and token budget?
func Coverage(docs, relevant []string) float64 {
	return RecallAtK(DedupeRanked(docs), relevant, len(docs))
}

// RetrievalScores holds the per-case retrieval metrics.
type RetrievalScores struct {
	HitAt         map[int]float64 `json:"hit_at"`
	RecallAt      map[int]float64 `json:"recall_at"`
	RR            float64         `json:"reciprocal_rank"`
	ContextRecall float64         `json:"context_recall"`
}

// ScoreRetrieval computes all retrieval metrics for one case.
// ranked is the document-level ranking from the vector search; contextDocs are
// the documents whose chunks were placed in the LLM context.
func ScoreRetrieval(ranked, contextDocs, relevant []string, ks []int) RetrievalScores {
	s := RetrievalScores{
		HitAt:         make(map[int]float64, len(ks)),
		RecallAt:      make(map[int]float64, len(ks)),
		RR:            ReciprocalRank(ranked, relevant),
		ContextRecall: Coverage(contextDocs, relevant),
	}
	for _, k := range ks {
		s.HitAt[k] = HitAtK(ranked, relevant, k)
		s.RecallAt[k] = RecallAtK(ranked, relevant, k)
	}
	return s
}
