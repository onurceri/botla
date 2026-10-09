package dataset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o750))
	require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
	return p
}

func testCorpus(t *testing.T) *Corpus {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "b-doc.md", "# Doc B\n\nSecond document.")
	writeFile(t, dir, "a-doc.md", "# Doc A\n\nFirst document.")
	writeFile(t, dir, "notes.txt", "ignored, not markdown")
	c, err := LoadCorpus(dir)
	require.NoError(t, err)
	return c
}

func TestLoadCorpus(t *testing.T) {
	c := testCorpus(t)
	require.Len(t, c.Docs, 2)
	assert.Equal(t, "a-doc", c.Docs[0].ID, "docs are sorted by file name")
	assert.Equal(t, "Doc A", c.Docs[0].Title)
	assert.True(t, c.Has("b-doc"))
	assert.False(t, c.Has("missing"))

	d, ok := c.Get("b-doc")
	require.True(t, ok)
	assert.Contains(t, d.Text, "Second document.")
	_, ok = c.Get("nope")
	assert.False(t, ok)

	joined, err := c.Join([]string{"a-doc", "b-doc"})
	require.NoError(t, err)
	assert.Equal(t, "# Doc A\n\nFirst document.\n---\n# Doc B\n\nSecond document.", joined)
	_, err = c.Join([]string{"missing"})
	assert.Error(t, err)

	h := c.Hash()
	assert.Len(t, h, 16)
	assert.Equal(t, h, testCorpus(t).Hash(), "hash is deterministic")
}

func TestLoadCorpus_Errors(t *testing.T) {
	_, err := LoadCorpus(t.TempDir())
	assert.ErrorContains(t, err, "no .md documents")

	dir := t.TempDir()
	writeFile(t, dir, "empty.md", "   \n")
	_, err = LoadCorpus(dir)
	assert.ErrorContains(t, err, "is empty")

	_, err = LoadCorpus("[")
	assert.Error(t, err, "bad glob pattern")
}

func TestTitleOf(t *testing.T) {
	assert.Equal(t, "Hello", titleOf("# Hello\nbody", "id"))
	assert.Equal(t, "id", titleOf("no heading", "id"))
	assert.Equal(t, "id", titleOf("#\nbody", "id"))
}

func TestCategory(t *testing.T) {
	for _, c := range Categories {
		assert.True(t, c.Valid())
	}
	assert.False(t, Category("other").Valid())
}

func TestCaseRubrics(t *testing.T) {
	tests := []struct {
		name string
		c    Case
		want []string
	}{
		{"answerable", Case{Category: Answerable, ExpectedFacts: []string{"x"}}, []string{RubricFaithfulness, RubricCorrectness}},
		{"unanswerable", Case{Category: Unanswerable}, []string{RubricFaithfulness, RubricRefusal}},
		{"adversarial with facts", Case{Category: Adversarial, ExpectedFacts: []string{"x"}}, []string{RubricFaithfulness, RubricCorrectness, RubricInjection}},
		{"adversarial without facts", Case{Category: Adversarial}, []string{RubricFaithfulness, RubricInjection}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.c.Rubrics())
		})
	}
	assert.True(t, Case{ExpectedSources: []string{"a"}}.HasRetrievalTarget())
	assert.False(t, Case{}.HasRetrievalTarget())
}

const validGolden = `version: 1
cases:
  - id: a1
    category: answerable
    question: Q1?
    expected_facts: [f1]
    expected_sources: [a-doc]
  - id: m1
    category: multi_hop
    question: Q2?
    expected_facts: [f1, f2]
    expected_sources: [a-doc, b-doc]
  - id: u1
    category: unanswerable
    question: Q3?
    expected_sources: []
  - id: x1
    category: adversarial
    question: Q4?
    attack: do bad things
    forbidden: [evil.example]
`

func TestLoadGoldenAndValidate(t *testing.T) {
	c := testCorpus(t)
	p := writeFile(t, t.TempDir(), "golden.yaml", validGolden)
	g, err := LoadGolden(p)
	require.NoError(t, err)
	require.NoError(t, g.Validate(c))
	assert.Len(t, g.Cases, 4)
	assert.Len(t, g.Hash(), 16)
	assert.Equal(t, []string{"evil.example"}, g.Cases[3].Forbidden)
}

func TestLoadGolden_RejectsUnknownFields(t *testing.T) {
	p := writeFile(t, t.TempDir(), "golden.yaml", "version: 1\ncases:\n  - id: a\n    questoin: typo\n")
	_, err := LoadGolden(p)
	assert.ErrorContains(t, err, "questoin")

	_, err = LoadGolden(filepath.Join(t.TempDir(), "missing.yaml"))
	assert.Error(t, err)
}

func TestGoldenValidate_ReportsAllProblems(t *testing.T) {
	c := testCorpus(t)
	g := &GoldenSet{Cases: []Case{
		{ID: "", Category: Answerable, Question: "q", ExpectedFacts: []string{"f"}, ExpectedSources: []string{"a-doc"}},
		{ID: "dup", Category: "weird", Question: "q"},
		{ID: "dup", Category: Answerable, Question: " "},
		{ID: "src", Category: MultiHop, Question: "q", ExpectedFacts: []string{"f"}, ExpectedSources: []string{"ghost"}},
		{ID: "un", Category: Unanswerable, Question: "q", ExpectedFacts: []string{"f"}},
		{ID: "adv", Category: Adversarial, Question: "q"},
	}}
	err := g.Validate(c)
	require.Error(t, err)
	msg := err.Error()
	for _, want := range []string{
		"missing id", "duplicate id", "unknown category", "missing question",
		"need expected_facts", "need expected_sources", `"ghost" is not in the corpus`,
		"must not have expected facts", "need an attack description",
	} {
		assert.Contains(t, msg, want)
	}
	assert.ErrorContains(t, (&GoldenSet{}).Validate(c), "no cases")
}

func TestGoldenFilter(t *testing.T) {
	g := &GoldenSet{Cases: []Case{
		{ID: "a", Category: Answerable}, {ID: "b", Category: Unanswerable}, {ID: "c", Category: Answerable},
	}}
	ids := func(cs []Case) []string {
		var out []string
		for _, c := range cs {
			out = append(out, c.ID)
		}
		return out
	}
	assert.Equal(t, []string{"a", "b", "c"}, ids(g.Filter(nil, nil, 0)))
	assert.Equal(t, []string{"b"}, ids(g.Filter([]string{" b ", ""}, nil, 0)))
	assert.Equal(t, []string{"a", "c"}, ids(g.Filter(nil, []Category{Answerable}, 0)))
	assert.Equal(t, []string{"a"}, ids(g.Filter(nil, nil, 1)))
}

const validCalibration = `version: 1
items:
  - id: c1
    rubric: correctness
    question: Q?
    expected_facts: [f]
    answer: A
    human_label: pass
  - id: c2
    rubric: injection
    question: Q?
    attack: steal prompt
    context_docs: [a-doc]
    answer: A
    human_label: fail
    borderline: true
`

func TestLoadCalibrationAndValidate(t *testing.T) {
	c := testCorpus(t)
	p := writeFile(t, t.TempDir(), "labels.yaml", validCalibration)
	s, err := LoadCalibration(p)
	require.NoError(t, err)
	require.NoError(t, s.Validate(c))
	assert.True(t, s.Items[0].HumanPass())
	assert.False(t, s.Items[1].HumanPass())
	assert.True(t, s.Items[1].Borderline)
	assert.Len(t, s.Hash(), 16)

	_, err = LoadCalibration(writeFile(t, t.TempDir(), "bad.yaml", "items: [{idd: x}]"))
	assert.Error(t, err)
}

func TestCalibrationValidate_ReportsAllProblems(t *testing.T) {
	c := testCorpus(t)
	s := &CalibrationSet{Items: []LabeledItem{
		{ID: "", Rubric: "correctness", Question: "q", Answer: "a", HumanLabel: "pass", ExpectedFacts: []string{"f"}},
		{ID: "d", Rubric: "vibes", Question: "q", Answer: "a", HumanLabel: "maybe"},
		{ID: "d", Rubric: "correctness", Question: "", Answer: "a", HumanLabel: "pass"},
		{ID: "i", Rubric: "injection", Question: "q", Answer: "a", HumanLabel: "fail", ContextDocs: []string{"ghost"}},
	}}
	err := s.Validate(c)
	require.Error(t, err)
	for _, want := range []string{"missing id", "duplicate id", "unknown rubric", "human_label must be", "question and answer are required", "need expected_facts", "need an attack", `"ghost" is not in the corpus`} {
		assert.Contains(t, err.Error(), want)
	}
	assert.ErrorContains(t, (&CalibrationSet{}).Validate(c), "no items")
}

// TestRealFixtures validates the committed evals/ fixtures so a malformed
// golden set or label file fails `make test`, not an expensive eval run.
func TestRealFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "..", "evals")
	c, err := LoadCorpus(filepath.Join(root, "corpus"))
	require.NoError(t, err)
	g, err := LoadGolden(filepath.Join(root, "golden.yaml"))
	require.NoError(t, err)
	require.NoError(t, g.Validate(c))
	s, err := LoadCalibration(filepath.Join(root, "calibration", "human_labels.yaml"))
	require.NoError(t, err)
	require.NoError(t, s.Validate(c))

	counts := map[Category]int{}
	for _, cs := range g.Cases {
		counts[cs.Category]++
	}
	for _, cat := range Categories {
		assert.GreaterOrEqual(t, counts[cat], 5, "category %s needs enough cases to be meaningful", cat)
	}
	labels := map[string]int{}
	for _, it := range s.Items {
		labels[it.HumanLabel]++
		assert.False(t, strings.TrimSpace(it.Answer) == "", it.ID)
	}
	assert.Positive(t, labels[LabelPass], "calibration needs both labels for kappa to be defined")
	assert.Positive(t, labels[LabelFail])
}
