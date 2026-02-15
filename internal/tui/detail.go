package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	styles "github.com/charmbracelet/glamour/styles"
	"github.com/koh-sh/ccplan/internal/plan"
)

// DetailPane manages the right pane that shows step details.
type DetailPane struct {
	viewport viewport.Model
	renderer *glamour.TermRenderer
	width    int
	height   int
}

// customDarkStyle returns a dark style with red background removed from
// Chroma error tokens. Japanese text can be misidentified as error tokens
// by Chroma, causing distracting red backgrounds.
func customDarkStyle() ansi.StyleConfig {
	style := styles.DarkStyleConfig
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
func NewDetailPane(width, height int) *DetailPane {
	vp := viewport.New(width, height)
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStyles(customDarkStyle()),
		glamour.WithWordWrap(width-4),
	)

	return &DetailPane{
		viewport: vp,
		renderer: renderer,
		width:    width,
		height:   height,
	}
}

// SetSize updates the pane size.
func (d *DetailPane) SetSize(width, height int) {
	if width == d.width && height == d.height {
		return
	}
	d.width = width
	d.height = height
	d.viewport.Width = width
	d.viewport.Height = height

	d.renderer, _ = glamour.NewTermRenderer(
		glamour.WithStyles(customDarkStyle()),
		glamour.WithWordWrap(width-4),
	)
}

// ShowStep renders and displays a step's content.
func (d *DetailPane) ShowStep(step *plan.Step, comment *plan.ReviewComment) {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("## %s: %s\n\n", step.ID, step.Title))

	if step.Body != "" {
		content.WriteString(step.Body + "\n")
	}

	if comment != nil {
		content.WriteString("\n---\n\n")
		content.WriteString(fmt.Sprintf("**Review Comment** [%s]\n\n", comment.Action))
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
		content.WriteString(fmt.Sprintf("# %s\n\n", p.Title))
	}
	if p.Preamble != "" {
		content.WriteString(p.Preamble + "\n")
	}

	d.renderContent(content.String())
}

// renderContent renders Markdown content into the viewport.
func (d *DetailPane) renderContent(md string) {
	rendered := md
	if d.renderer != nil {
		if r, err := d.renderer.Render(md); err == nil {
			rendered = r
		}
	}
	d.viewport.SetContent(rendered)
	d.viewport.GotoTop()
}

// View returns the viewport view.
func (d *DetailPane) View() string {
	return d.viewport.View()
}

// Viewport returns a pointer to the viewport for event handling.
func (d *DetailPane) Viewport() *viewport.Model {
	return &d.viewport
}
