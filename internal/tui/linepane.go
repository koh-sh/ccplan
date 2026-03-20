package tui

import (
	"fmt"
	"strings"

	"github.com/koh-sh/commd/internal/markdown"
	"github.com/mattn/go-runewidth"
)

// sectionRange maps a line range to a section ID.
type sectionRange struct {
	startLine int // 1-based
	endLine   int // 1-based
	sectionID string
}

// LinePane renders raw source with line numbers for line-level commenting.
type LinePane struct {
	lines         []string
	cursor        int // 0-based index into lines
	selectAnchor  int // -1 = no selection
	scrollOffset  int // first visible line index
	viewStart     int // 0-based start of visible range (-1 = show all)
	viewEnd       int // 0-based end of visible range (exclusive)
	width         int
	height        int
	gutterWidth   int
	styles        Styles
	comments      []*markdown.ReviewComment
	sectionRanges []sectionRange
}

// NewLinePane creates a new LinePane.
func NewLinePane(lines []string, width, height int, styles Styles, sections []*markdown.Section) *LinePane {
	lp := &LinePane{
		lines:        lines,
		selectAnchor: -1,
		viewStart:    -1,
		width:        width,
		height:       height,
		styles:       styles,
	}
	lp.gutterWidth = len(fmt.Sprintf("%d", max(len(lines), 1))) + 1
	lp.buildSectionRanges(sections)
	return lp
}

func (lp *LinePane) buildSectionRanges(sections []*markdown.Section) {
	lp.sectionRanges = nil
	for _, s := range sections {
		if s.StartLine > 0 {
			lp.sectionRanges = append(lp.sectionRanges, sectionRange{
				startLine: s.StartLine,
				endLine:   s.EndLine,
				sectionID: s.ID,
			})
		}
	}
}

// SetSize updates the pane dimensions.
func (lp *LinePane) SetSize(width, height int) {
	lp.width = width
	lp.height = height
}

// SetViewRange sets the visible line range (1-based, inclusive).
// Pass 0, 0 to show all lines.
func (lp *LinePane) SetViewRange(startLine, endLine int) {
	if startLine <= 0 || endLine <= 0 {
		lp.viewStart = -1
		lp.viewEnd = 0
		return
	}
	lp.viewStart = startLine - 1
	lp.viewEnd = min(endLine, len(lp.lines))
	lp.clampCursor()
	lp.ensureVisible()
}

// ClearViewRange shows all lines.
func (lp *LinePane) ClearViewRange() {
	lp.viewStart = -1
	lp.viewEnd = 0
}

// rangeStart returns the 0-based start index of the visible range.
func (lp *LinePane) rangeStart() int {
	if lp.viewStart < 0 {
		return 0
	}
	return lp.viewStart
}

// rangeEnd returns the 0-based exclusive end index of the visible range.
func (lp *LinePane) rangeEnd() int {
	if lp.viewStart < 0 {
		return len(lp.lines)
	}
	return lp.viewEnd
}

// clampCursor ensures the cursor is within the visible range.
func (lp *LinePane) clampCursor() {
	lo := lp.rangeStart()
	hi := max(lp.rangeEnd()-1, lo)
	if lp.cursor < lo {
		lp.cursor = lo
	}
	if lp.cursor > hi {
		lp.cursor = hi
	}
}

// CursorUp moves the cursor up one line.
func (lp *LinePane) CursorUp() {
	if lp.cursor > lp.rangeStart() {
		lp.cursor--
		lp.ensureVisible()
	}
}

// CursorDown moves the cursor down one line.
func (lp *LinePane) CursorDown() {
	if lp.cursor < lp.rangeEnd()-1 {
		lp.cursor++
		lp.ensureVisible()
	}
}

// CursorTop moves the cursor to the first visible line.
func (lp *LinePane) CursorTop() {
	lp.cursor = lp.rangeStart()
	lp.scrollOffset = lp.cursor
	lp.clampScroll()
}

// CursorBottom moves the cursor to the last visible line.
func (lp *LinePane) CursorBottom() {
	lp.cursor = max(lp.rangeEnd()-1, lp.rangeStart())
	lp.ensureVisible()
}

// HalfPageDown moves the cursor down by half the viewport height.
func (lp *LinePane) HalfPageDown() {
	lp.cursor = min(lp.cursor+lp.height/2, max(lp.rangeEnd()-1, lp.rangeStart()))
	lp.ensureVisible()
}

// HalfPageUp moves the cursor up by half the viewport height.
func (lp *LinePane) HalfPageUp() {
	lp.cursor = max(lp.cursor-lp.height/2, lp.rangeStart())
	lp.ensureVisible()
}

// PageDown moves the cursor down by a full viewport height.
func (lp *LinePane) PageDown() {
	lp.cursor = min(lp.cursor+lp.height, max(lp.rangeEnd()-1, lp.rangeStart()))
	lp.ensureVisible()
}

// PageUp moves the cursor up by a full viewport height.
func (lp *LinePane) PageUp() {
	lp.cursor = max(lp.cursor-lp.height, lp.rangeStart())
	lp.ensureVisible()
}

// StartVisualSelect begins visual line selection from the current cursor position.
func (lp *LinePane) StartVisualSelect() {
	lp.selectAnchor = lp.cursor
}

// IsVisualSelect returns true if visual selection is active.
func (lp *LinePane) IsVisualSelect() bool {
	return lp.selectAnchor >= 0
}

// CancelVisualSelect cancels visual line selection.
func (lp *LinePane) CancelVisualSelect() {
	lp.selectAnchor = -1
}

// SelectedRange returns the 1-based line range for commenting.
// Without visual selection: returns (cursor+1, 0) for single line.
// With visual selection: returns (min+1, max+1) for range.
func (lp *LinePane) SelectedRange() (startLine, endLine int) {
	if lp.selectAnchor < 0 {
		return lp.cursor + 1, 0
	}
	lo := min(lp.selectAnchor, lp.cursor)
	hi := max(lp.selectAnchor, lp.cursor)
	return lo + 1, hi + 1
}

// ScrollToLine scrolls the viewport so the given 1-based line is visible,
// centered if possible.
func (lp *LinePane) ScrollToLine(line int) {
	idx := max(line-1, 0)
	if idx >= len(lp.lines) {
		idx = max(len(lp.lines)-1, 0)
	}
	lp.cursor = idx
	// Center the cursor in the viewport
	lp.scrollOffset = max(idx-lp.height/2, 0)
	lp.clampScroll()
}

// SectionIDAtLine returns the section ID containing the given 1-based line number.
// Assumes sectionRanges is sorted by startLine in ascending order.
// Returns OverviewSectionID if the line is before any section.
func (lp *LinePane) SectionIDAtLine(line int) string {
	result := markdown.OverviewSectionID
	for _, sr := range lp.sectionRanges {
		if line >= sr.startLine {
			result = sr.sectionID
		} else {
			break
		}
	}
	return result
}

// Cursor returns the current 0-based cursor position.
func (lp *LinePane) Cursor() int {
	return lp.cursor
}

// LineCount returns the total number of source lines.
func (lp *LinePane) LineCount() int {
	return len(lp.lines)
}

// AtRangeTop returns true if the cursor is at the top of the visible range.
func (lp *LinePane) AtRangeTop() bool {
	return lp.cursor <= lp.rangeStart()
}

// AtRangeBottom returns true if the cursor is at the bottom of the visible range.
func (lp *LinePane) AtRangeBottom() bool {
	return lp.cursor >= lp.rangeEnd()-1
}

// SetComments sets the comments for inline display.
func (lp *LinePane) SetComments(comments []*markdown.ReviewComment) {
	lp.comments = comments
}

// View renders the line pane content.
func (lp *LinePane) View() string {
	if len(lp.lines) == 0 {
		return ""
	}

	// Build a map of line number -> comments for inline display
	commentMap := lp.buildCommentMap()

	contentWidth := max(
		// gutter + separator + padding
		lp.width-lp.gutterWidth-3, 1)

	var sb strings.Builder
	rEnd := lp.rangeEnd()
	visibleEnd := min(lp.scrollOffset+lp.height, rEnd)
	linesRendered := 0

	for i := lp.scrollOffset; i < visibleEnd && linesRendered < lp.height; i++ {
		lineNum := i + 1
		gutter := lp.styles.LineGutter.Render(fmt.Sprintf("%*d ", lp.gutterWidth, lineNum))
		separator := lp.styles.LineGutter.Render("│")

		content := lp.lines[i]

		// Apply cursor or selection styling
		var styledContent string
		switch {
		case i == lp.cursor:
			styledContent = lp.styles.LineCursor.Render(fitToWidth(content, contentWidth))
		case lp.isInSelection(i):
			styledContent = lp.styles.LineSelected.Render(fitToWidth(content, contentWidth))
		default:
			styledContent = fitToWidth(content, contentWidth)
		}

		sb.WriteString(gutter + separator + " " + styledContent)
		linesRendered++

		if linesRendered < lp.height {
			sb.WriteString("\n")
		}

		// Render inline comment boxes after their target lines
		if comments, ok := commentMap[lineNum]; ok && linesRendered < lp.height {
			for _, c := range comments {
				if linesRendered >= lp.height {
					break
				}
				box := lp.renderInlineCommentBox(c, contentWidth)
				boxLines := strings.SplitSeq(box, "\n")
				for boxLine := range boxLines {
					if linesRendered >= lp.height {
						break
					}
					padding := strings.Repeat(" ", lp.gutterWidth+2)
					sb.WriteString(padding + " " + boxLine + "\n")
					linesRendered++
				}
			}
		}
	}

	// Pad remaining lines if viewport is not full
	for linesRendered < lp.height {
		sb.WriteString(strings.Repeat(" ", lp.gutterWidth+1) + lp.styles.LineGutter.Render("│"))
		linesRendered++
		if linesRendered < lp.height {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (lp *LinePane) ensureVisible() {
	if lp.cursor < lp.scrollOffset {
		lp.scrollOffset = lp.cursor
	}
	if lp.cursor >= lp.scrollOffset+lp.height {
		lp.scrollOffset = lp.cursor - lp.height + 1
	}
	lp.clampScroll()
}

func (lp *LinePane) clampScroll() {
	lo := lp.rangeStart()
	hi := lp.rangeEnd()
	maxOffset := max(hi-lp.height, lo)
	if lp.scrollOffset > maxOffset {
		lp.scrollOffset = maxOffset
	}
	if lp.scrollOffset < lo {
		lp.scrollOffset = lo
	}
}

func (lp *LinePane) isInSelection(idx int) bool {
	if lp.selectAnchor < 0 {
		return false
	}
	lo := min(lp.selectAnchor, lp.cursor)
	hi := max(lp.selectAnchor, lp.cursor)
	return idx >= lo && idx <= hi
}

// buildCommentMap groups line-level comments by the line they should be displayed after.
// For single-line comments, they appear after StartLine.
// For range comments, they appear after EndLine.
func (lp *LinePane) buildCommentMap() map[int][]*markdown.ReviewComment {
	m := make(map[int][]*markdown.ReviewComment)
	for _, c := range lp.comments {
		if c.StartLine == 0 {
			continue // section-level comment, skip
		}
		displayLine := c.StartLine
		if c.EndLine > 0 {
			displayLine = c.EndLine
		}
		m[displayLine] = append(m[displayLine], c)
	}
	return m
}

func (lp *LinePane) renderInlineCommentBox(c *markdown.ReviewComment, maxWidth int) string {
	lineRef := c.FormatLineRef()
	header := fmt.Sprintf("Review Comment [%s]", c.FormatLabel())
	if lineRef != "" {
		header += " (" + lineRef + ")"
	}

	content := header
	if c.Body != "" {
		content += "\n" + c.Body
	}

	boxWidth := max(maxWidth, 20)

	style := lp.styles.CommentBorder.
		Width(boxWidth).
		Padding(0, 1)

	return style.Render(content)
}

// fitToWidth truncates or pads a string to exactly the given display width.
func fitToWidth(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw > width {
		return runewidth.Truncate(s, width, "")
	}
	if sw < width {
		return s + strings.Repeat(" ", width-sw)
	}
	return s
}
