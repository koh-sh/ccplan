package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	styles "github.com/charmbracelet/glamour/styles"
	"github.com/koh-sh/ccplan/internal/plan"
	"github.com/mattn/go-runewidth"
)

// glamourHorizontalOverhead accounts for glamour's default left/right
// margin (2 each) and padding (2 each) = 8 total.
const glamourHorizontalOverhead = 8

// DetailPane manages the right pane that shows step details.
type DetailPane struct {
	viewport  viewport.Model
	renderer  *glamour.TermRenderer
	width     int
	height    int
	theme     string
	wrapWidth int // prose wrap width (excludes glamour margins)
}

// customStyle returns a glamour style with red background removed from
// Chroma error tokens. Japanese text can be misidentified as error tokens
// by Chroma, causing distracting red backgrounds.
func customStyle(theme string) ansi.StyleConfig {
	var style ansi.StyleConfig
	if theme == "light" {
		style = styles.LightStyleConfig
	} else {
		style = styles.DarkStyleConfig
	}
	if style.CodeBlock.Chroma != nil {
		chroma := *style.CodeBlock.Chroma
		chroma.Error = ansi.StylePrimitive{
			Color: chroma.Error.Color,
		}
		style.CodeBlock.Chroma = &chroma
	}
	return style
}

// NewDetailPane creates a new DetailPane.
func NewDetailPane(width, height int, theme string) *DetailPane {
	vp := viewport.New(width, height)
	// Intentionally ignore error: renderContent falls back to plain text when renderer is nil.
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(customStyle(theme)),
		glamour.WithWordWrap(0),
	)

	return &DetailPane{
		viewport:  vp,
		renderer:  renderer,
		width:     width,
		height:    height,
		theme:     theme,
		wrapWidth: width - glamourHorizontalOverhead,
	}
}

// SetSize updates the pane size. It does not re-render current content;
// call ShowStep or ShowOverview after resizing to refresh the viewport.
func (d *DetailPane) SetSize(width, height int) {
	if width == d.width && height == d.height {
		return
	}
	d.width = width
	d.height = height
	d.viewport.Width = width
	d.viewport.Height = height

	d.wrapWidth = width - glamourHorizontalOverhead

	// Intentionally ignore error: renderContent falls back to plain text when renderer is nil.
	d.renderer, _ = glamour.NewTermRenderer(
		glamour.WithStyles(customStyle(d.theme)),
		glamour.WithWordWrap(0),
	)
}

// ShowStep renders and displays a step's content.
func (d *DetailPane) ShowStep(step *plan.Step, comments []*plan.ReviewComment) {
	var content strings.Builder

	fmt.Fprintf(&content, "## %s: %s\n\n", step.ID, step.Title)

	if step.Body != "" {
		content.WriteString(step.Body + "\n")
	}

	for i, comment := range comments {
		content.WriteString("\n---\n\n")
		if len(comments) == 1 {
			fmt.Fprintf(&content, "**Review Comment** [%s]\n\n", comment.Action)
		} else {
			fmt.Fprintf(&content, "**Review Comment #%d** [%s]\n\n", i+1, comment.Action)
		}
		if comment.Body != "" {
			content.WriteString(comment.Body + "\n")
		}
	}

	d.renderContent(content.String())
}

// ShowOverview renders and displays the plan overview (preamble).
func (d *DetailPane) ShowOverview(p *plan.Plan) {
	var content strings.Builder

	if p.Title != "" {
		fmt.Fprintf(&content, "# %s\n\n", p.Title)
	}
	if p.Preamble != "" {
		content.WriteString(p.Preamble + "\n")
	}

	d.renderContent(content.String())
}

// renderContent renders Markdown content into the viewport.
func (d *DetailPane) renderContent(md string) {
	md = wrapProse(md, d.wrapWidth)
	rendered := md
	if d.renderer != nil {
		if r, err := d.renderer.Render(md); err == nil {
			rendered = r
		}
	}
	d.viewport.SetContent(rendered)
	d.viewport.SetXOffset(0)
	d.viewport.GotoTop()
}

// wrapProse wraps prose lines in Markdown to the given width while preserving
// code blocks (fenced with ``` or ~~~) as-is. This allows glamour to render
// with WordWrap(0) so code blocks keep their original formatting.
func wrapProse(md string, width int) string {
	if width <= 0 {
		return md
	}
	lines := strings.Split(md, "\n")
	var result []string
	var fenceMarker string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if fenceMarker == "" &&
			(strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			fenceMarker = trimmed[:3]
			result = append(result, line)
			continue
		}
		if fenceMarker != "" && strings.HasPrefix(trimmed, fenceMarker) {
			fenceMarker = ""
			result = append(result, line)
			continue
		}
		if fenceMarker != "" || runewidth.StringWidth(line) <= width {
			result = append(result, line)
			continue
		}
		result = append(result, softWrapLine(line, width)...)
	}
	return strings.Join(result, "\n")
}

// softWrapLine breaks a single long line at word boundaries, preserving
// any leading whitespace indent on the first and continuation lines.
// Uses display width (runewidth) so multibyte characters wrap correctly.
func softWrapLine(line string, width int) []string {
	// Preserve leading whitespace.
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]

	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return []string{line}
	}

	indentWidth := runewidth.StringWidth(indent)
	effectiveWidth := width - indentWidth
	if effectiveWidth <= 0 {
		effectiveWidth = 1
	}

	var lines []string
	current := words[0]
	currentWidth := runewidth.StringWidth(current)
	for _, word := range words[1:] {
		ww := runewidth.StringWidth(word)
		if currentWidth+1+ww > effectiveWidth {
			lines = append(lines, indent+current)
			current = word
			currentWidth = ww
		} else {
			current += " " + word
			currentWidth += 1 + ww
		}
	}
	lines = append(lines, indent+current)
	return lines
}

// View returns the viewport view.
func (d *DetailPane) View() string {
	return d.viewport.View()
}

// Viewport returns a pointer to the viewport for event handling.
func (d *DetailPane) Viewport() *viewport.Model {
	return &d.viewport
}
