package dataset

import (
	"errors"
	"fmt"
	"strings"
)

// Human labels for calibration items.
const (
	LabelPass = "pass"
	LabelFail = "fail"
)

// LabeledItem is one human-graded answer used to measure judge agreement.
type LabeledItem struct {
	ID            string   `yaml:"id" json:"id"`
	Rubric        string   `yaml:"rubric" json:"rubric"`
	Question      string   `yaml:"question" json:"question"`
	ContextDocs   []string `yaml:"context_docs" json:"context_docs,omitempty"`
	ExpectedFacts []string `yaml:"expected_facts" json:"expected_facts,omitempty"`
	Attack        string   `yaml:"attack" json:"attack,omitempty"`
	Answer        string   `yaml:"answer" json:"answer"`
	HumanLabel    string   `yaml:"human_label" json:"human_label"`
	Borderline    bool     `yaml:"borderline" json:"borderline,omitempty"`
	Note          string   `yaml:"note" json:"note,omitempty"`
}

// HumanPass reports whether the human grader marked the item as passing.
func (it LabeledItem) HumanPass() bool { return it.HumanLabel == LabelPass }

// CalibrationSet is the parsed human_labels.yaml.
type CalibrationSet struct {
	Version int           `yaml:"version"`
	Items   []LabeledItem `yaml:"items"`
	hash    string
}

// Hash is a content hash of the calibration file.
func (s *CalibrationSet) Hash() string { return s.hash }

// LoadCalibration parses a calibration file, rejecting unknown fields.
func LoadCalibration(path string) (*CalibrationSet, error) {
	var s CalibrationSet
	h, err := decodeStrict(path, &s)
	if err != nil {
		return nil, err
	}
	s.hash = h
	return &s, nil
}

var knownRubrics = map[string]bool{
	RubricFaithfulness: true,
	RubricCorrectness:  true,
	RubricRefusal:      true,
	RubricInjection:    true,
}

// Validate checks labels, rubric names, rubric-specific inputs and corpus references.
func (s *CalibrationSet) Validate(c *Corpus) error {
	var errs []error
	if len(s.Items) == 0 {
		errs = append(errs, errors.New("calibration set has no items"))
	}
	seen := make(map[string]bool, len(s.Items))
	for i, it := range s.Items {
		where := fmt.Sprintf("item #%d (%s)", i, it.ID)
		if strings.TrimSpace(it.ID) == "" {
			errs = append(errs, fmt.Errorf("item #%d: missing id", i))
		} else if seen[it.ID] {
			errs = append(errs, fmt.Errorf("%s: duplicate id", where))
		}
		seen[it.ID] = true
		if !knownRubrics[it.Rubric] {
			errs = append(errs, fmt.Errorf("%s: unknown rubric %q", where, it.Rubric))
		}
		if it.HumanLabel != LabelPass && it.HumanLabel != LabelFail {
			errs = append(errs, fmt.Errorf("%s: human_label must be %q or %q", where, LabelPass, LabelFail))
		}
		if strings.TrimSpace(it.Question) == "" || strings.TrimSpace(it.Answer) == "" {
			errs = append(errs, fmt.Errorf("%s: question and answer are required", where))
		}
		if it.Rubric == RubricCorrectness && len(it.ExpectedFacts) == 0 {
			errs = append(errs, fmt.Errorf("%s: correctness items need expected_facts", where))
		}
		if it.Rubric == RubricInjection && strings.TrimSpace(it.Attack) == "" {
			errs = append(errs, fmt.Errorf("%s: injection items need an attack description", where))
		}
		for _, d := range it.ContextDocs {
			if !c.Has(d) {
				errs = append(errs, fmt.Errorf("%s: context doc %q is not in the corpus", where, d))
			}
		}
	}
	return errors.Join(errs...)
}
