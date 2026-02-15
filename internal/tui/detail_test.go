package tui

import (
	"strings"
	"testing"

	"github.com/koh-sh/ccplan/internal/plan"
)

func TestNewDetailPane(t *testing.T) {
	dp := NewDetailPane(80, 24, "dark")
	if dp == nil {
		t.Fatal("NewDetailPane returned nil")
	}
	if dp.width != 80 {
		t.Errorf("width = %d, want 80", dp.width)
	}
	if dp.height != 24 {
		t.Errorf("height = %d, want 24", dp.height)
	}
}

func TestDetailPaneShowStep(t *testing.T) {
	dp := NewDetailPane(80, 40, "dark")
	step := &plan.Step{ID: "S1", Title: "Test Step", Body: "Unique test body here"}

	dp.ShowStep(step, nil)
	content := dp.View()
	if !strings.Contains(content, "S1") {
		t.Error("view should contain step ID")
	}
	// glamour wraps text and inserts ANSI codes; check individual words
	if !strings.Contains(content, "Unique") {
		t.Errorf("view should contain step body word, got:\n%s", content)
	}
}

func TestDetailPaneShowStepWithComments(t *testing.T) {
	dp := NewDetailPane(80, 40, "dark")
	step := &plan.Step{ID: "S1", Title: "Test Step", Body: "Body"}
	comments := []*plan.ReviewComment{
		{StepID: "S1", Action: plan.ActionSuggestion, Body: "Review text"},
	}

	dp.ShowStep(step, comments)
	content := dp.View()
	if !strings.Contains(content, "Review") {
		t.Error("view should contain 'Review'")
	}
}

func TestDetailPaneShowStepWithMultipleComments(t *testing.T) {
	dp := NewDetailPane(80, 40, "dark")
	step := &plan.Step{ID: "S1", Title: "Test Step", Body: "Body"}
	comments := []*plan.ReviewComment{
		{StepID: "S1", Action: plan.ActionSuggestion, Body: "First"},
		{StepID: "S1", Action: plan.ActionIssue, Body: "Second"},
	}

	dp.ShowStep(step, comments)
	content := dp.View()
	if !strings.Contains(content, "#1") {
		t.Error("view should contain '#1' for numbered comments")
	}
}

func TestDetailPaneShowOverview(t *testing.T) {
	dp := NewDetailPane(80, 40, "dark")
	p := &plan.Plan{Title: "My Plan", Preamble: "Unique overview text"}

	dp.ShowOverview(p)
	content := dp.View()
	if !strings.Contains(content, "Unique") {
		t.Errorf("view should contain preamble word, got:\n%s", content)
	}
}

func TestDetailPaneSetSize(t *testing.T) {
	dp := NewDetailPane(80, 24, "dark")

	// Same size should be no-op
	dp.SetSize(80, 24)
	if dp.width != 80 || dp.height != 24 {
		t.Error("same size should not change")
	}

	// Different size
	dp.SetSize(100, 30)
	if dp.width != 100 {
		t.Errorf("width = %d, want 100", dp.width)
	}
	if dp.height != 30 {
		t.Errorf("height = %d, want 30", dp.height)
	}
}

func TestCustomStyle(t *testing.T) {
	dark := customStyle("dark")
	light := customStyle("light")

	// Both should not have BackgroundColor on Error token
	if dark.CodeBlock.Chroma != nil && dark.CodeBlock.Chroma.Error.BackgroundColor != nil {
		t.Error("dark style should not have error background color")
	}
	if light.CodeBlock.Chroma != nil && light.CodeBlock.Chroma.Error.BackgroundColor != nil {
		t.Error("light style should not have error background color")
	}
}
