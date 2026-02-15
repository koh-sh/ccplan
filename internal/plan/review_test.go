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
			{StepID: "S1", Action: ActionModify, Body: "Change the algorithm."},
			{StepID: "S2", Action: ActionDelete, Body: "Not needed."},
		},
	}

	output := FormatReview(result, p)

	if !strings.Contains(output, "## S1: First Step [modify]") {
		t.Errorf("output missing S1 modify header, got:\n%s", output)
	}
	if !strings.Contains(output, "Change the algorithm.") {
		t.Errorf("output missing S1 body, got:\n%s", output)
	}
	if !strings.Contains(output, "## S2: Second Step [delete]") {
		t.Errorf("output missing S2 delete header, got:\n%s", output)
	}
}

func TestFormatReviewEmpty(t *testing.T) {
	p := &Plan{Title: "Test"}
	result := &ReviewResult{}
	output := FormatReview(result, p)
	if output != "" {
		t.Errorf("expected empty output, got: %q", output)
	}
}

func TestFormatReviewApprove(t *testing.T) {
	p := &Plan{
		Steps: []*Step{
			{ID: "S1", Title: "Step One", Level: 2},
		},
	}
	result := &ReviewResult{
		Comments: []ReviewComment{
			{StepID: "S1", Action: ActionApprove, Body: ""},
		},
	}

	output := FormatReview(result, p)
	if !strings.Contains(output, "[approve]") {
		t.Errorf("output missing approve action, got:\n%s", output)
	}
}
