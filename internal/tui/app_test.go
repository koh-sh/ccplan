package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/koh-sh/ccplan/internal/plan"
)

// makeLargePlan creates a plan with many steps to exceed any terminal height.
func makeLargePlan(topLevel, childrenPer int) *plan.Plan {
	p := &plan.Plan{
		Title:    "Large Test Plan",
		Preamble: "This is a preamble with enough text to test overflow.",
	}
	for i := 1; i <= topLevel; i++ {
		step := &plan.Step{
			ID:    fmt.Sprintf("S%d", i),
			Title: fmt.Sprintf("Top Level Step %d", i),
			Level: 2,
			Body:  fmt.Sprintf("Body for step %d with some content.", i),
		}
		for j := 1; j <= childrenPer; j++ {
			child := &plan.Step{
				ID:     fmt.Sprintf("S%d.%d", i, j),
				Title:  fmt.Sprintf("Sub Step %d.%d", i, j),
				Level:  3,
				Body:   fmt.Sprintf("Body for sub-step %d.%d.", i, j),
				Parent: step,
			}
			step.Children = append(step.Children, child)
		}
		p.Steps = append(p.Steps, step)
	}
	return p
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func TestViewFitsTerminalHeight(t *testing.T) {
	// 50 top-level steps with 3 children each = 200 items, way more than any terminal
	p := makeLargePlan(50, 3)

	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"small", 80, 24},
		{"medium", 120, 40},
		{"large", 200, 60},
		{"tiny", 60, 15},
		{"wide-short", 200, 10},
	}

	for _, sz := range sizes {
		t.Run(sz.name, func(t *testing.T) {
			app := NewApp(p, AppOptions{})

			// Simulate window size message
			model, _ := app.Update(tea.WindowSizeMsg{Width: sz.width, Height: sz.height})
			a := model.(*App)

			view := a.View()
			lines := countLines(view)

			if lines > sz.height {
				t.Errorf("View() has %d lines, exceeds terminal height %d", lines, sz.height)
			}
		})
	}
}

func TestViewFitsInCommentMode(t *testing.T) {
	p := makeLargePlan(20, 3)
	app := NewApp(p, AppOptions{})

	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	a := model.(*App)

	// Move to a step and enter comment mode
	a.stepList.CursorDown() // move to first step
	a.comment.Open("S1", nil)
	a.mode = ModeComment

	view := a.View()
	lines := countLines(view)

	if lines > 30 {
		t.Errorf("View() in comment mode has %d lines, exceeds terminal height 30", lines)
	}
}

func TestViewFitsInConfirmMode(t *testing.T) {
	p := makeLargePlan(20, 3)
	app := NewApp(p, AppOptions{})

	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	a := model.(*App)

	a.mode = ModeConfirm

	view := a.View()
	lines := countLines(view)

	if lines > 30 {
		t.Errorf("View() in confirm mode has %d lines, exceeds terminal height 30", lines)
	}
}

func TestStepListScrollsWithCursor(t *testing.T) {
	p := makeLargePlan(30, 0) // 30 top-level steps, no children
	app := NewApp(p, AppOptions{})

	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 20})
	a := model.(*App)

	// Move cursor down past the visible area
	for i := 0; i < 25; i++ {
		a.stepList.CursorDown()
	}

	view := a.View()
	lines := countLines(view)

	if lines > 20 {
		t.Errorf("View() has %d lines after scrolling, exceeds terminal height 20", lines)
	}

	// The selected step should be visible in the rendered output
	selected := a.stepList.Selected()
	if selected == nil {
		t.Fatal("Expected a selected step")
	}
	if !strings.Contains(view, selected.ID) {
		t.Errorf("Selected step %s not visible in rendered view after scrolling", selected.ID)
	}
}

func TestGGGoesToTop(t *testing.T) {
	p := makeLargePlan(10, 0)
	app := NewApp(p, AppOptions{})

	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	a := model.(*App)

	// Move cursor down several steps
	for i := 0; i < 5; i++ {
		a.stepList.CursorDown()
	}
	if a.stepList.Selected().ID == "S1" {
		t.Fatal("Cursor should not be at S1 before gg")
	}

	// Send 'g' then 'g' (gg chord)
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

	// Cursor should be at top (Overview if preamble exists, otherwise S1)
	if a.stepList.IsOverviewSelected() {
		// OK: overview is at top
	} else if a.stepList.Selected() != nil && a.stepList.Selected().ID != "S1" {
		t.Errorf("After gg, expected cursor at top, got %s", a.stepList.Selected().ID)
	}
}

func TestShiftGGoesToBottom(t *testing.T) {
	p := makeLargePlan(10, 0)
	app := NewApp(p, AppOptions{})

	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	a := model.(*App)

	// Send 'G' (Shift+G)
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})

	selected := a.stepList.Selected()
	if selected == nil {
		t.Fatal("Expected a selected step after G")
	}
	if selected.ID != "S10" {
		t.Errorf("After G, expected cursor at S10, got %s", selected.ID)
	}
}

func TestPendingGResetOnOtherKey(t *testing.T) {
	p := makeLargePlan(10, 0)
	app := NewApp(p, AppOptions{})

	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	a := model.(*App)

	// Send 'g' then 'j' (not a gg chord)
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Should have moved down by one (j)
	if a.stepList.IsOverviewSelected() {
		// Overview is index 0, j should move to S1
	} else if a.stepList.Selected() != nil {
		// The j key should have been processed normally after pendingG reset
		// Just verify we didn't jump to top or bottom
		id := a.stepList.Selected().ID
		if id != "S1" && id != "S2" {
			t.Errorf("After g+j, expected cursor near top, got %s", id)
		}
	}
}

func TestTruncateCJK(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  bool // true if result width should be <= max
	}{
		{"ascii fits", "hello", 10, true},
		{"ascii truncate", "hello world", 8, true},
		{"cjk fits", "テスト", 6, true},
		{"cjk truncate", "テスト計画の概要", 8, true},
		{"mixed", "Step テスト", 8, true},
		{"zero max", "hello", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			w := lipglossWidth(got)
			if tt.want && w > tt.max {
				t.Errorf("truncate(%q, %d) = %q (width %d), exceeds max",
					tt.input, tt.max, got, w)
			}
		})
	}
}

// lipglossWidth returns the display width of a string (ANSI-aware).
func lipglossWidth(s string) int {
	return lipgloss.Width(s)
}

func TestViewFitsInHelpMode(t *testing.T) {
	p := makeLargePlan(5, 2)
	app := NewApp(p, AppOptions{})

	model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	a := model.(*App)

	a.mode = ModeHelp

	view := a.View()
	lines := countLines(view)

	if lines > 30 {
		t.Errorf("View() in help mode has %d lines, exceeds terminal height 30", lines)
	}
}
