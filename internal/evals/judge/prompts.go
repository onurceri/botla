// Package judge implements LLM-as-a-judge scoring with versioned rubric
// prompts, strict structured-output parsing and a content-hash verdict cache.
//
// Design choices (see evals/README.md for the rationale):
//   - The judge never emits a free-form 1-10 score. It emits structured
//     sub-judgements (per-claim support, per-fact coverage, booleans) and the
//     score and pass/fail are computed here, deterministically, from those.
//   - Prompts are files under evals/judge/<version>/; their hash is part of the
//     cache key and of every report.
package judge

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/onurceri/botla-app/internal/evals/dataset"
	"gopkg.in/yaml.v3"
)

// Rubric names.
const (
	Faithfulness = dataset.RubricFaithfulness
	Correctness  = dataset.RubricCorrectness
	Refusal      = dataset.RubricRefusal
	Injection    = dataset.RubricInjection
)

// Rubric is one grading prompt plus its pass rule.
type Rubric struct {
	Name         string
	PassMinScore float64
	tmpl         *template.Template
}

// PromptSet is a versioned set of rubric prompts loaded from a directory.
type PromptSet struct {
	Version string
	System  string
	Rubrics map[string]*Rubric
	// Hash covers the manifest and every prompt file, so any edit to wording
	// invalidates cached verdicts even if the version string is not bumped.
	Hash string
}

type manifest struct {
	Version      string `yaml:"version"`
	SystemPrompt string `yaml:"system_prompt"`
	Rubrics      map[string]struct {
		Template     string  `yaml:"template"`
		PassMinScore float64 `yaml:"pass_min_score"`
	} `yaml:"rubrics"`
}

// Input is everything a rubric template may reference.
type Input struct {
	Question      string
	Context       string
	Answer        string
	ExpectedFacts []string
	Attack        string
}

var templateFuncs = template.FuncMap{"inc": func(i int) int { return i + 1 }}

// LoadPromptSet loads manifest.yaml and the prompt files it references from dir.
func LoadPromptSet(dir string) (*PromptSet, error) {
	h := sha256.New()
	read := func(name string) (string, error) {
		raw, err := os.ReadFile(filepath.Join(dir, filepath.Clean(name))) // #nosec G304 -- files named by the versioned manifest
		if err != nil {
			return "", fmt.Errorf("read judge prompt %s: %w", name, err)
		}
		_, _ = fmt.Fprintf(h, "%s:%d:", name, len(raw))
		_, _ = h.Write(raw)
		return string(raw), nil
	}
	rawManifest, err := read("manifest.yaml")
	if err != nil {
		return nil, err
	}
	var m manifest
	if yamlErr := yaml.Unmarshal([]byte(rawManifest), &m); yamlErr != nil {
		return nil, fmt.Errorf("parse judge manifest: %w", yamlErr)
	}
	if m.Version == "" || m.SystemPrompt == "" || len(m.Rubrics) == 0 {
		return nil, fmt.Errorf("judge manifest in %s must set version, system_prompt and rubrics", dir)
	}
	system, err := read(m.SystemPrompt)
	if err != nil {
		return nil, err
	}
	ps := &PromptSet{Version: m.Version, System: strings.TrimSpace(system), Rubrics: map[string]*Rubric{}}
	names := make([]string, 0, len(m.Rubrics))
	for name := range m.Rubrics {
		names = append(names, name)
	}
	sort.Strings(names) // hash files in a stable order
	for _, name := range names {
		spec := m.Rubrics[name]
		if spec.PassMinScore <= 0 || spec.PassMinScore > 1 {
			return nil, fmt.Errorf("rubric %s: pass_min_score must be in (0, 1]", name)
		}
		body, readErr := read(spec.Template)
		if readErr != nil {
			return nil, readErr
		}
		tmpl, parseErr := template.New(name).Funcs(templateFuncs).Option("missingkey=error").Parse(body)
		if parseErr != nil {
			return nil, fmt.Errorf("parse rubric template %s: %w", name, parseErr)
		}
		ps.Rubrics[name] = &Rubric{Name: name, PassMinScore: spec.PassMinScore, tmpl: tmpl}
	}
	ps.Hash = hex.EncodeToString(h.Sum(nil))[:16]
	return ps, nil
}

// Render produces the user prompt for a rubric.
func (p *PromptSet) Render(rubric string, in Input) (string, error) {
	r, ok := p.Rubrics[rubric]
	if !ok {
		return "", fmt.Errorf("unknown rubric %q", rubric)
	}
	var b strings.Builder
	if err := r.tmpl.Execute(&b, in); err != nil {
		return "", fmt.Errorf("render rubric %s: %w", rubric, err)
	}
	return strings.TrimSpace(b.String()), nil
}
