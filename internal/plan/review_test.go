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

	if !strings.Contains(output, "## S1: First Step\n") {
		t.Errorf("output missing S1 header, got:\n%s", output)
	}
	if !strings.Contains(output, "[suggestion] Change the algorithm.") {
		t.Errorf("output missing S1 body, got:\n%s", output)
	}
	if !strings.Contains(output, "## S2: Second Step\n") {
		t.Errorf("output missing S2 header, got:\n%s", output)
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
	if !strings.Contains(output, "[suggestion] Looks good but needs refactoring.") {
		t.Errorf("output missing action+body, got:\n%s", output)
	}
}

func TestFormatReviewGroupedComments(t *testing.T) {
	p := &Plan{
		Steps: []*Step{
			{ID: "S1", Title: "JWT verification", Level: 2},
			{ID: "S3", Title: "Add tests", Level: 2},
		},
	}
	result := &ReviewResult{
		Comments: []ReviewComment{
			{StepID: "S1", Action: ActionSuggestion, Body: "Switch to HS256."},
			{StepID: "S1", Action: ActionIssue, Body: "Not needed."},
			{StepID: "S3", Action: ActionQuestion, Body: "Coverage target?"},
		},
	}

	output := FormatReview(result, p, "/path/to/plan.md")

	// S1 heading should appear only once
	if strings.Count(output, "## S1: JWT verification") != 1 {
		t.Errorf("expected exactly one S1 heading, got:\n%s", output)
	}
	// Both comments under S1
	if !strings.Contains(output, "[suggestion] Switch to HS256.") {
		t.Errorf("output missing S1 suggestion, got:\n%s", output)
	}
	if !strings.Contains(output, "[issue] Not needed.") {
		t.Errorf("output missing S1 issue, got:\n%s", output)
	}
	// S3 separate heading
	if !strings.Contains(output, "## S3: Add tests\n") {
		t.Errorf("output missing S3 header, got:\n%s", output)
	}
	if !strings.Contains(output, "[question] Coverage target?") {
		t.Errorf("output missing S3 question, got:\n%s", output)
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
