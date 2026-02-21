package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/koh-sh/ccplan/internal/plan"
	"github.com/mattn/go-runewidth"
)

// StepListItem is a flattened step for display in the step list.
type StepListItem struct {
	Step       *plan.Step
	Depth      int
	Expanded   bool
	Visible    bool
	IsOverview bool // true for the overview (preamble) entry
}

// StepList manages the left pane step tree.
type StepList struct {
	items        []StepListItem
	cursor       int
	scrollOffset int
	comments     map[string][]*plan.ReviewComment // stepID -> comments
	viewed       map[string]bool                  // stepID -> viewed flag
	viewedState  *plan.ViewedState
	plan         *plan.Plan
}

// NewStepList creates a new StepList from a parsed plan.
func NewStepList(p *plan.Plan, state *plan.ViewedState) *StepList {
	sl := &StepList{
		comments:    make(map[string][]*plan.ReviewComment),
		viewed:      make(map[string]bool),
		viewedState: state,
		plan:        p,
	}

	// Add overview entry if there's a preamble
	if p.Preamble != "" {
		sl.items = append(sl.items, StepListItem{
			Visible:    true,
			IsOverview: true,
		})
	}

	// Flatten the step tree
	var flatten func(steps []*plan.Step, depth int)
	flatten = func(steps []*plan.Step, depth int) {
		for _, s := range steps {
			sl.items = append(sl.items, StepListItem{
				Step:     s,
				Depth:    depth,
				Expanded: true,
				Visible:  true,
			})
			flatten(s.Children, depth+1)
		}
	}
	flatten(p.Steps, 0)

	// Restore viewed flags from persisted state
	if state != nil {
		for i, item := range sl.items {
			if item.Step != nil && state.IsStepViewed(item.Step) {
				sl.viewed[sl.items[i].Step.ID] = true
			}
		}
	}

	return sl
}

// CursorUp moves the cursor up to the previous visible item.
func (sl *StepList) CursorUp() {
	for i := sl.cursor - 1; i >= 0; i-- {
		if sl.items[i].Visible {
			sl.cursor = i
			return
		}
	}
}

// CursorDown moves the cursor down to the next visible item.
func (sl *StepList) CursorDown() {
	for i := sl.cursor + 1; i < len(sl.items); i++ {
		if sl.items[i].Visible {
			sl.cursor = i
			return
		}
	}
}

// CursorTop moves the cursor to the first visible item.
func (sl *StepList) CursorTop() {
	for i := 0; i < len(sl.items); i++ {
		if sl.items[i].Visible {
			sl.cursor = i
			return
		}
	}
}

// CursorBottom moves the cursor to the last visible item.
func (sl *StepList) CursorBottom() {
	for i := len(sl.items) - 1; i >= 0; i-- {
		if sl.items[i].Visible {
			sl.cursor = i
			return
		}
	}
}

// ToggleExpand toggles the expand/collapse state of the current step.
func (sl *StepList) ToggleExpand() {
	if sl.cursor >= len(sl.items) {
		return
	}
	item := &sl.items[sl.cursor]
	if item.IsOverview || item.Step == nil || len(item.Step.Children) == 0 {
		return
	}
	item.Expanded = !item.Expanded
	sl.updateVisibility()
}

// Expand expands the current step.
func (sl *StepList) Expand() {
	if sl.cursor >= len(sl.items) {
		return
	}
	item := &sl.items[sl.cursor]
	if item.IsOverview || item.Step == nil || len(item.Step.Children) == 0 {
		return
	}
	if !item.Expanded {
		item.Expanded = true
		sl.updateVisibility()
	}
}

// Collapse collapses the current step.
func (sl *StepList) Collapse() {
	if sl.cursor >= len(sl.items) {
		return
	}
	item := &sl.items[sl.cursor]
	if item.IsOverview {
		return
	}

	// If current item has children and is expanded, collapse it
	if item.Step != nil && len(item.Step.Children) > 0 && item.Expanded {
		item.Expanded = false
		sl.updateVisibility()
		return
	}

	// Otherwise, move to parent
	if item.Step != nil && item.Step.Parent != nil {
		for i, it := range sl.items {
			if it.Step == item.Step.Parent {
				sl.cursor = i
				return
			}
		}
	}
}

// updateVisibility updates the Visible field for all items based on parent expansion state.
func (sl *StepList) updateVisibility() {
	collapsed := make(map[*plan.Step]bool)
	for _, item := range sl.items {
		if item.Step != nil && !item.Expanded {
			collapsed[item.Step] = true
		}
	}

	for i := range sl.items {
		if sl.items[i].IsOverview {
			sl.items[i].Visible = true
			continue
		}
		if sl.items[i].Step == nil {
			continue
		}

		visible := true
		parent := sl.items[i].Step.Parent
		for parent != nil {
			if collapsed[parent] {
				visible = false
				break
			}
			parent = parent.Parent
		}
		sl.items[i].Visible = visible
	}
}

// Selected returns the currently selected step (or nil for overview).
func (sl *StepList) Selected() *plan.Step {
	if sl.cursor >= len(sl.items) {
		return nil
	}
	return sl.items[sl.cursor].Step
}

// IsOverviewSelected returns true if the overview entry is selected.
func (sl *StepList) IsOverviewSelected() bool {
	if sl.cursor >= len(sl.items) {
		return false
	}
	return sl.items[sl.cursor].IsOverview
}

// AddComment appends a comment for a step.
func (sl *StepList) AddComment(stepID string, comment *plan.ReviewComment) {
	if comment == nil || comment.Body == "" {
		return
	}
	sl.comments[stepID] = append(sl.comments[stepID], comment)
}

// UpdateComment replaces a comment at the given index for a step.
func (sl *StepList) UpdateComment(stepID string, index int, comment *plan.ReviewComment) {
	comments := sl.comments[stepID]
	if index < 0 || index >= len(comments) {
		return
	}
	if comment == nil || comment.Body == "" {
		sl.DeleteComment(stepID, index)
		return
	}
	sl.comments[stepID][index] = comment
}

// DeleteComment removes a comment at the given index for a step.
func (sl *StepList) DeleteComment(stepID string, index int) {
	comments := sl.comments[stepID]
	if index < 0 || index >= len(comments) {
		return
	}
	sl.comments[stepID] = append(comments[:index], comments[index+1:]...)
	if len(sl.comments[stepID]) == 0 {
		delete(sl.comments, stepID)
	}
}

// ToggleViewed toggles the viewed flag for a step.
func (sl *StepList) ToggleViewed(stepID string) {
	sl.viewed[stepID] = !sl.viewed[stepID]

	// Sync with persisted state
	if sl.viewedState != nil {
		if step := sl.plan.FindStep(stepID); step != nil {
			if sl.viewed[stepID] {
				sl.viewedState.MarkViewed(step)
			} else {
				sl.viewedState.UnmarkViewed(step)
			}
		}
	}
}

// ViewedState returns the underlying ViewedState for persistence.
func (sl *StepList) ViewedState() *plan.ViewedState {
	return sl.viewedState
}

// IsViewed returns whether a step is marked as viewed.
func (sl *StepList) IsViewed(stepID string) bool {
	return sl.viewed[stepID]
}

// GetComments returns all comments for a step.
func (sl *StepList) GetComments(stepID string) []*plan.ReviewComment {
	return sl.comments[stepID]
}

// HasComments returns true if there are any comments.
func (sl *StepList) HasComments() bool {
	for _, comments := range sl.comments {
		if len(comments) > 0 {
			return true
		}
	}
	return false
}

// BuildReviewResult creates a ReviewResult from all comments.
func (sl *StepList) BuildReviewResult() *plan.ReviewResult {
	result := &plan.ReviewResult{}

	// Walk steps in order to maintain consistent ordering
	allSteps := sl.plan.AllSteps()
	for _, s := range allSteps {
		for _, c := range sl.comments[s.ID] {
			result.Comments = append(result.Comments, *c)
		}
	}

	return result
}

// Render renders the step list for display within the given height.
func (sl *StepList) Render(width, height int, styles Styles) string {
	// Build list of visible item indices
	var visibleIndices []int
	for i, item := range sl.items {
		if item.Visible {
			visibleIndices = append(visibleIndices, i)
		}
	}

	// Find cursor position in visible list
	cursorPos := 0
	for vi, idx := range visibleIndices {
		if idx == sl.cursor {
			cursorPos = vi
			break
		}
	}

	// Calculate available lines for items
	itemLines := max(height, 1)

	// Adjust scroll offset to keep cursor visible
	if cursorPos < sl.scrollOffset {
		sl.scrollOffset = cursorPos
	}
	if cursorPos >= sl.scrollOffset+itemLines {
		sl.scrollOffset = cursorPos - itemLines + 1
	}
	if sl.scrollOffset < 0 {
		sl.scrollOffset = 0
	}

	var sb strings.Builder

	// Only render items in the visible window
	end := min(sl.scrollOffset+itemLines, len(visibleIndices))

	for vi := sl.scrollOffset; vi < end; vi++ {
		i := visibleIndices[vi]
		item := sl.items[i]

		var line string
		if item.IsOverview {
			line = "  Overview"
		} else if item.Step != nil {
			indent := strings.Repeat("  ", item.Depth)
			prefix := " "
			if len(item.Step.Children) > 0 {
				if item.Expanded {
					prefix = "▼"
				} else {
					prefix = "▶"
				}
			}

			badge := sl.renderBadge(item.Step.ID, styles)
			stepText := fmt.Sprintf("%s%s %s %s", indent, prefix, item.Step.ID, item.Step.Title)
			line = truncate(stepText, width-4-lipgloss.Width(badge)) + badge
		}

		if i == sl.cursor {
			line = styles.SelectedStep.Render("> " + line)
		} else {
			line = styles.NormalStep.Render("  " + line)
		}

		sb.WriteString(line + "\n")
	}

	return sb.String()
}

// renderBadge renders the badge for a step (comment indicator, viewed mark).
func (sl *StepList) renderBadge(stepID string, styles Styles) string {
	commentCount := len(sl.comments[stepID])
	isViewed := sl.viewed[stepID]

	var badge string
	if commentCount == 1 {
		badge += styles.StepBadge.Render(" [*]")
	} else if commentCount > 1 {
		badge += styles.StepBadge.Render(fmt.Sprintf(" [*%d]", commentCount))
	}
	if isViewed {
		badge += styles.ViewedBadge.Render(" [✓]")
	}
	return badge
}

// FilterByQuery filters the step list to show only steps matching the query.
// Matching is case-insensitive against step ID + Title.
// If a child matches, its ancestors are shown. If a parent matches, its children are shown.
func (sl *StepList) FilterByQuery(query string) {
	if query == "" {
		sl.ClearFilter()
		return
	}

	query = strings.ToLower(query)

	// First pass: mark direct matches
	matched := make(map[int]bool)
	for i, item := range sl.items {
		if item.IsOverview {
			if strings.Contains("overview", query) { //nolint:gocritic // intentional: match when query is a substring of "overview"
				matched[i] = true
			}
			continue
		}
		if item.Step == nil {
			continue
		}
		text := strings.ToLower(item.Step.ID + " " + item.Step.Title)
		if strings.Contains(text, query) {
			matched[i] = true
		}
	}

	// Second pass: if a step matches, show its ancestors
	ancestorVisible := make(map[*plan.Step]bool)
	for i, item := range sl.items {
		if !matched[i] || item.Step == nil {
			continue
		}
		parent := item.Step.Parent
		for parent != nil {
			ancestorVisible[parent] = true
			parent = parent.Parent
		}
	}

	// Third pass: if a step matches, show its descendants
	descendantVisible := make(map[*plan.Step]bool)
	for i, item := range sl.items {
		if !matched[i] || item.Step == nil {
			continue
		}
		var markDescendants func(steps []*plan.Step)
		markDescendants = func(steps []*plan.Step) {
			for _, s := range steps {
				descendantVisible[s] = true
				markDescendants(s.Children)
			}
		}
		markDescendants(item.Step.Children)
	}

	// Apply visibility
	for i := range sl.items {
		item := &sl.items[i]
		if item.IsOverview {
			item.Visible = matched[i]
			continue
		}
		if item.Step == nil {
			item.Visible = false
			continue
		}
		item.Visible = matched[i] || ancestorVisible[item.Step] || descendantVisible[item.Step]
	}

	// Move cursor to first visible item if current is hidden
	if sl.cursor < len(sl.items) && !sl.items[sl.cursor].Visible {
		sl.CursorTop()
	}
}

// ClearFilter resets visibility to respect only expansion state.
func (sl *StepList) ClearFilter() {
	sl.updateVisibility()
}

// truncate truncates a string to fit within max display-width cells, with ellipsis.
func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return runewidth.Truncate(s, maxWidth, "")
	}
	return runewidth.Truncate(s, maxWidth, "...")
}
