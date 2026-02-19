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

func TestDetailPaneHorizontalScroll(t *testing.T) {
	const width = 40
	dp := NewDetailPane(width, 40, "dark")

	longCode := "```\n" + strings.Repeat("x", 100) + "\n```"
	step := &plan.Step{ID: "S1", Title: "Test", Body: longCode}

	dp.ShowStep(step, nil)

	// Verify scroll starts at 0%
	if pct := dp.Viewport().HorizontalScrollPercent(); pct != 0 {
		t.Errorf("initial HorizontalScrollPercent = %f, want 0", pct)
	}

	// Scroll right and verify position changed
	dp.Viewport().ScrollRight(4)
	if pct := dp.Viewport().HorizontalScrollPercent(); pct == 0 {
		t.Error("after ScrollRight(4) HorizontalScrollPercent should be > 0")
	}

	// ShowStep resets scroll position
	dp.Viewport().ScrollRight(10)
	dp.ShowStep(step, nil)
	if pct := dp.Viewport().HorizontalScrollPercent(); pct != 0 {
		t.Errorf("after ShowStep HorizontalScrollPercent = %f, want 0", pct)
	}
}

func TestWrapProse(t *testing.T) {
	tests := []struct {
		name  string
		md    string
		width int
		want  string
	}{
		{"empty string", "", 40, ""},
		{"zero width returns as-is", "some long text here", 0, "some long text here"},
		{"negative width returns as-is", "text", -1, "text"},
		{"short line no wrap", "hello world", 40, "hello world"},
		{"long prose wrapped", "aaa bbb ccc ddd", 7, "aaa bbb\nccc ddd"},
		{
			"backtick code block preserved",
			"```\n" + strings.Repeat("x", 100) + "\n```",
			20,
			"```\n" + strings.Repeat("x", 100) + "\n```",
		},
		{
			"tilde code block preserved",
			"~~~\n" + strings.Repeat("y", 100) + "\n~~~",
			20,
			"~~~\n" + strings.Repeat("y", 100) + "\n~~~",
		},
		{
			"prose outside code block wrapped",
			"```\ncode line\n```\n" + strings.Repeat("word ", 20),
			30,
			"```\ncode line\n```\n" + func() string {
				got := wrapProse(strings.Repeat("word ", 20), 30)
				return got
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapProse(tt.md, tt.width)
			if got != tt.want {
				t.Errorf("wrapProse(%q, %d):\n got: %q\nwant: %q", tt.md, tt.width, got, tt.want)
			}
		})
	}
}

func TestSoftWrapLine(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		width int
		want  []string
	}{
		{"empty line", "   ", 40, []string{"   "}},
		{"single word fits", "hello", 40, []string{"hello"}},
		{"wraps at boundary", "aaa bbb ccc", 7, []string{"aaa bbb", "ccc"}},
		{
			"preserves leading indent",
			"  - item one two three",
			14,
			[]string{"  - item one", "  two three"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := softWrapLine(tt.line, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("softWrapLine(%q, %d) returned %d lines, want %d:\n got: %q\nwant: %q",
					tt.line, tt.width, len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
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
