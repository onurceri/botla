// Package dataset loads and validates the evaluation fixtures: the fixed
// document corpus, the golden question set and the human-labelled judge
// calibration set (see evals/README.md).
package dataset

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Category groups golden cases by the behaviour they probe.
type Category string

const (
	// Answerable cases have their answer stated in a single corpus document.
	Answerable Category = "answerable"
	// MultiHop cases require combining evidence from several passages.
	MultiHop Category = "multi_hop"
	// Unanswerable cases have no answer in the corpus; the bot should decline.
	Unanswerable Category = "unanswerable"
	// Adversarial cases contain a prompt-injection or manipulation attempt.
	Adversarial Category = "adversarial"
)

// Categories lists every category in report order.
var Categories = []Category{Answerable, MultiHop, Unanswerable, Adversarial}

// Valid reports whether c is a known category.
func (c Category) Valid() bool {
	for _, k := range Categories {
		if c == k {
			return true
		}
	}
	return false
}

// Rubric names shared with the judge package. They live here so that the
// dataset can decide which rubrics apply to a case without importing the judge.
const (
	RubricFaithfulness = "faithfulness"
	RubricCorrectness  = "correctness"
	RubricRefusal      = "refusal"
	RubricInjection    = "injection"
)

// Document is one corpus file. ID is the file name without extension.
type Document struct {
	ID    string
	Title string
	Text  string
}

// Corpus is the fixed document set the chatbot is "trained" on for evals.
type Corpus struct {
	Docs []Document
	byID map[string]int
}

// LoadCorpus reads every *.md file in dir, sorted by name for determinism.
func LoadCorpus(dir string) (*Corpus, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, fmt.Errorf("glob corpus: %w", err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no .md documents found in %s", dir)
	}
	sort.Strings(paths)
	c := &Corpus{byID: make(map[string]int, len(paths))}
	for _, p := range paths {
		raw, readErr := os.ReadFile(p) // #nosec G304 -- path comes from a glob over the configured corpus dir
		if readErr != nil {
			return nil, fmt.Errorf("read corpus doc %s: %w", p, readErr)
		}
		id := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		text := strings.TrimSpace(string(raw))
		if text == "" {
			return nil, fmt.Errorf("corpus doc %s is empty", id)
		}
		c.byID[id] = len(c.Docs)
		c.Docs = append(c.Docs, Document{ID: id, Title: titleOf(text, id), Text: text})
	}
	return c, nil
}

func titleOf(text, fallback string) string {
	first, _, _ := strings.Cut(text, "\n")
	if t := strings.TrimSpace(strings.TrimLeft(first, "#")); t != "" && strings.HasPrefix(first, "#") {
		return t
	}
	return fallback
}

// Get returns the document with the given ID.
func (c *Corpus) Get(id string) (Document, bool) {
	i, ok := c.byID[id]
	if !ok {
		return Document{}, false
	}
	return c.Docs[i], true
}

// Has reports whether a document with the given ID exists.
func (c *Corpus) Has(id string) bool {
	_, ok := c.byID[id]
	return ok
}

// Join concatenates the full text of the given documents using the same
// separator the product uses between retrieved chunks.
func (c *Corpus) Join(ids []string) (string, error) {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		d, ok := c.Get(id)
		if !ok {
			return "", fmt.Errorf("unknown corpus doc %q", id)
		}
		parts = append(parts, d.Text)
	}
	return strings.Join(parts, "\n---\n"), nil
}

// Hash is a content hash of the whole corpus, recorded in reports so results
// from different corpus versions are never compared by accident.
func (c *Corpus) Hash() string {
	h := sha256.New()
	for _, d := range c.Docs {
		_, _ = fmt.Fprintf(h, "%d:%s\n%d:%s\n", len(d.ID), d.ID, len(d.Text), d.Text)
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// Case is one golden evaluation case.
type Case struct {
	ID              string   `yaml:"id" json:"id"`
	Category        Category `yaml:"category" json:"category"`
	Question        string   `yaml:"question" json:"question"`
	ExpectedFacts   []string `yaml:"expected_facts" json:"expected_facts,omitempty"`
	ExpectedSources []string `yaml:"expected_sources" json:"expected_sources"`
	Attack          string   `yaml:"attack" json:"attack,omitempty"`
	Forbidden       []string `yaml:"forbidden" json:"forbidden,omitempty"`
	Notes           string   `yaml:"notes" json:"notes,omitempty"`
}

// HasRetrievalTarget reports whether retrieval metrics are defined for the case.
func (c Case) HasRetrievalTarget() bool { return len(c.ExpectedSources) > 0 }

// Rubrics returns the judge rubrics that apply to the case, in a fixed order.
//
//   - Every answer is checked for faithfulness to its retrieved context.
//   - Cases with reference facts are checked for correctness.
//   - Unanswerable cases are checked for a clean refusal.
//   - Adversarial cases are checked for injection resistance.
func (c Case) Rubrics() []string {
	r := []string{RubricFaithfulness}
	if len(c.ExpectedFacts) > 0 {
		r = append(r, RubricCorrectness)
	}
	switch c.Category {
	case Unanswerable:
		r = append(r, RubricRefusal)
	case Adversarial:
		r = append(r, RubricInjection)
	}
	return r
}

// GoldenSet is the parsed golden.yaml.
type GoldenSet struct {
	Version int    `yaml:"version"`
	Cases   []Case `yaml:"cases"`
	hash    string
}

// Hash is a content hash of the golden file.
func (g *GoldenSet) Hash() string { return g.hash }

// LoadGolden parses a golden set file. Unknown YAML fields are rejected so
// that typos in field names fail loudly instead of silently dropping data.
func LoadGolden(path string) (*GoldenSet, error) {
	var g GoldenSet
	h, err := decodeStrict(path, &g)
	if err != nil {
		return nil, err
	}
	g.hash = h
	return &g, nil
}

// LoadGoldenFiles loads one or more golden set files and concatenates their
// cases in order (e.g. golden.yaml plus a held-out file). With a single path
// it is identical to LoadGolden. The combined hash covers every file, so a run
// over a different set of files never shares a baseline identity with another.
// Duplicate IDs across files are reported by Validate.
func LoadGoldenFiles(paths ...string) (*GoldenSet, error) {
	if len(paths) == 0 {
		return nil, errors.New("no golden set files given")
	}
	if len(paths) == 1 {
		return LoadGolden(paths[0])
	}
	out := &GoldenSet{}
	h := sha256.New()
	for _, p := range paths {
		g, err := LoadGolden(p)
		if err != nil {
			return nil, err
		}
		if out.Version == 0 {
			out.Version = g.Version
		}
		out.Cases = append(out.Cases, g.Cases...)
		_, _ = fmt.Fprintf(h, "%s\n", g.hash)
	}
	out.hash = hex.EncodeToString(h.Sum(nil))[:16]
	return out, nil
}

// Validate checks internal consistency and that every referenced document
// exists in the corpus. It returns all problems at once.
func (g *GoldenSet) Validate(c *Corpus) error {
	var errs []error
	if len(g.Cases) == 0 {
		errs = append(errs, errors.New("golden set has no cases"))
	}
	seen := make(map[string]bool, len(g.Cases))
	for i, cs := range g.Cases {
		where := fmt.Sprintf("case #%d (%s)", i, cs.ID)
		if strings.TrimSpace(cs.ID) == "" {
			errs = append(errs, fmt.Errorf("case #%d: missing id", i))
		} else if seen[cs.ID] {
			errs = append(errs, fmt.Errorf("%s: duplicate id", where))
		}
		seen[cs.ID] = true
		if !cs.Category.Valid() {
			errs = append(errs, fmt.Errorf("%s: unknown category %q", where, cs.Category))
		}
		if strings.TrimSpace(cs.Question) == "" {
			errs = append(errs, fmt.Errorf("%s: missing question", where))
		}
		for _, s := range cs.ExpectedSources {
			if !c.Has(s) {
				errs = append(errs, fmt.Errorf("%s: expected source %q is not in the corpus", where, s))
			}
		}
		switch cs.Category {
		case Answerable, MultiHop:
			if len(cs.ExpectedFacts) == 0 {
				errs = append(errs, fmt.Errorf("%s: %s cases need expected_facts", where, cs.Category))
			}
			if len(cs.ExpectedSources) == 0 {
				errs = append(errs, fmt.Errorf("%s: %s cases need expected_sources", where, cs.Category))
			}
		case Unanswerable:
			if len(cs.ExpectedFacts) > 0 || len(cs.ExpectedSources) > 0 {
				errs = append(errs, fmt.Errorf("%s: unanswerable cases must not have expected facts or sources", where))
			}
		case Adversarial:
			if strings.TrimSpace(cs.Attack) == "" {
				errs = append(errs, fmt.Errorf("%s: adversarial cases need an attack description", where))
			}
		}
	}
	return errors.Join(errs...)
}

// Filter returns the cases matching the given IDs and categories (empty means
// "all"), truncated to limit when limit > 0.
func (g *GoldenSet) Filter(ids []string, cats []Category, limit int) []Case {
	idSet := toSet(ids)
	catSet := make(map[Category]bool, len(cats))
	for _, c := range cats {
		catSet[c] = true
	}
	var out []Case
	for _, cs := range g.Cases {
		if len(idSet) > 0 && !idSet[cs.ID] {
			continue
		}
		if len(catSet) > 0 && !catSet[cs.Category] {
			continue
		}
		out = append(out, cs)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func toSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		if x = strings.TrimSpace(x); x != "" {
			m[x] = true
		}
	}
	return m
}

// decodeStrict reads a YAML file into v, rejecting unknown fields, and returns
// a short content hash of the file.
func decodeStrict(path string, v any) (string, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- path is an operator-supplied fixture file
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])[:16], nil
}
