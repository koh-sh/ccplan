package plan

import (
	"strings"
	"testing"
)

func TestFormatReview(t *testing.T) {
	p := &Plan{
		Title: "Test Plan",
		Steps: []*Step{
			{ID: "S1", Title: "First Step", Level: 2},
			{ID: "S2", Title: "Second Step", Level: 2},
		},
	}

	result := &ReviewResult{
		Comments: []ReviewComment{
			{StepID: "S1", Action: ActionSuggestion, Body: "Change the algorithm."},
			{StepID: "S2", Action: ActionSuggestion, Body: "Not needed."},
		},
	}

	output := FormatReview(result, p, "/path/to/plan.md")

	if !strings.Contains(output, "## S1: First Step [suggestion]") {
		t.Errorf("output missing S1 modify header, got:\n%s", output)
	}
	if !strings.Contains(output, "Change the algorithm.") {
		t.Errorf("output missing S1 body, got:\n%s", output)
	}
	if !strings.Contains(output, "## S2: Second Step [suggestion]") {
		t.Errorf("output missing S2 modify header, got:\n%s", output)
	}
	if !strings.Contains(output, "/path/to/plan.md") {
		t.Errorf("output missing file path, got:\n%s", output)
	}
}

func TestFormatReviewEmpty(t *testing.T) {
	p := &Plan{Title: "Test"}
	result := &ReviewResult{}
	output := FormatReview(result, p, "")
	if output != "" {
		t.Errorf("expected empty output, got: %q", output)
	}
}

func TestFormatReviewWithBody(t *testing.T) {
	p := &Plan{
		Steps: []*Step{
			{ID: "S1", Title: "Step One", Level: 2},
		},
	}
	result := &ReviewResult{
		Comments: []ReviewComment{
			{StepID: "S1", Action: ActionSuggestion, Body: "Looks good but needs refactoring."},
		},
	}

	output := FormatReview(result, p, "test.md")
	if !strings.Contains(output, "[suggestion]") {
		t.Errorf("output missing modify action, got:\n%s", output)
	}
	if !strings.Contains(output, "Looks good but needs refactoring.") {
		t.Errorf("output missing body, got:\n%s", output)
	}
}

func TestFormatReviewEmptyFilePath(t *testing.T) {
	p := &Plan{
		Steps: []*Step{
			{ID: "S1", Title: "Step One", Level: 2},
		},
	}
	result := &ReviewResult{
		Comments: []ReviewComment{
			{StepID: "S1", Action: ActionSuggestion, Body: "comment"},
		},
	}

	output := FormatReview(result, p, "")
	if !strings.Contains(output, "the file") {
		t.Errorf("expected fallback 'the file' when filePath is empty, got:\n%s", output)
	}
}
