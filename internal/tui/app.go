package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/koh-sh/ccplan/internal/plan"
)

// AppMode represents the current application mode.
type AppMode int

const (
	ModeNormal  AppMode = iota // Step list navigation
	ModeComment                // Comment input
	ModeConfirm                // Confirmation dialog
	ModeHelp                   // Help overlay
)

// Focus represents which pane has focus.
type Focus int

const (
	FocusLeft  Focus = iota
	FocusRight
)

// AppResult is the result returned when the TUI exits.
type AppResult struct {
	Review *plan.ReviewResult
	Status plan.Status
}

// App is the main Bubble Tea model for the TUI.
type App struct {
	plan     *plan.Plan
	stepList *StepList
	detail   *DetailPane
	comment  *CommentEditor
	keymap   KeyMap
	styles   Styles

	mode   AppMode
	focus  Focus
	width  int
	height int
	ready  bool
	opts   AppOptions

	result   AppResult
	pendingG bool // gg chord: true when first 'g' was pressed
}

// AppOptions configures the TUI appearance.
type AppOptions struct {
	Theme    string // "dark" or "light"
	NoColor  bool
	FilePath string // plan file path (displayed in title bar)
}

// NewApp creates a new App model.
func NewApp(p *plan.Plan, opts AppOptions) *App {
	return &App{
		plan:     p,
		stepList: NewStepList(p),
		comment:  NewCommentEditor(),
		keymap:   DefaultKeyMap(),
		styles:   stylesForTheme(opts.Theme, opts.NoColor),
		opts:     opts,
		result: AppResult{
			Status: plan.StatusCancelled,
		},
	}
}

// Result returns the final result after the TUI exits.
func (a *App) Result() AppResult {
	return a.result
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateLayout()
		a.ready = true
		a.refreshDetail()
		return a, nil

	case tea.KeyMsg:
		return a.handleKey(msg)
	}

	if a.mode == ModeComment {
		cmd := a.comment.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch a.mode {
	case ModeNormal:
		return a.handleNormalMode(msg)
	case ModeComment:
		return a.handleCommentMode(msg)
	case ModeConfirm:
		return a.handleConfirmMode(msg)
	case ModeHelp:
		return a.handleHelpMode(msg)
	}
	return a, nil
}

func (a *App) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle 'gg' chord (second g after pending g)
	if a.pendingG {
		a.pendingG = false
		if msg.String() == "g" {
			if a.focus == FocusLeft {
				a.stepList.CursorTop()
				a.refreshDetail()
			} else {
				a.detail.Viewport().GotoTop()
			}
			return a, nil
		}
		// Not 'g' after pending g — fall through to normal handling
	}

	// Check for 'g' (start of gg chord) and 'G' (go to bottom)
	switch msg.String() {
	case "g":
		a.pendingG = true
		return a, nil
	case "G":
		if a.focus == FocusLeft {
			a.stepList.CursorBottom()
			a.refreshDetail()
		} else {
			a.detail.Viewport().GotoBottom()
		}
		return a, nil
	}

	switch {
	case key.Matches(msg, a.keymap.Quit):
		if a.stepList.HasComments() {
			a.mode = ModeConfirm
			return a, nil
		}
		a.result.Status = plan.StatusCancelled
		return a, tea.Quit

	case key.Matches(msg, a.keymap.Help):
		a.mode = ModeHelp
		return a, nil

	case key.Matches(msg, a.keymap.Tab):
		if a.focus == FocusLeft {
			a.focus = FocusRight
		} else {
			a.focus = FocusLeft
		}
		return a, nil

	case key.Matches(msg, a.keymap.Submit):
		return a.submitReview()
	}

	if a.focus == FocusLeft {
		return a.handleLeftPaneKeys(msg)
	}
	return a.handleRightPaneKeys(msg)
}

func (a *App) handleLeftPaneKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keymap.Up):
		a.stepList.CursorUp()
		a.refreshDetail()

	case key.Matches(msg, a.keymap.Down):
		a.stepList.CursorDown()
		a.refreshDetail()

	case key.Matches(msg, a.keymap.Toggle):
		a.stepList.ToggleExpand()

	case key.Matches(msg, a.keymap.Expand):
		a.stepList.Expand()

	case key.Matches(msg, a.keymap.Collapse):
		a.stepList.Collapse()

	case key.Matches(msg, a.keymap.Comment):
		if step := a.stepList.Selected(); step != nil {
			existing := a.stepList.GetComment(step.ID)
			a.comment.Open(step.ID, existing)
			a.mode = ModeComment
			return a, a.comment.textarea.Focus()
		}

	case key.Matches(msg, a.keymap.Delete):
		if step := a.stepList.Selected(); step != nil {
			existing := a.stepList.GetComment(step.ID)
			if existing != nil && existing.Action == plan.ActionDelete {
				a.stepList.SetComment(step.ID, nil)
			} else {
				body := ""
				if existing != nil {
					body = existing.Body
				}
				a.stepList.SetComment(step.ID, &plan.ReviewComment{
					StepID: step.ID,
					Action: plan.ActionDelete,
					Body:   body,
				})
			}
			a.refreshDetail()
		}

	case key.Matches(msg, a.keymap.Approve):
		if step := a.stepList.Selected(); step != nil {
			existing := a.stepList.GetComment(step.ID)
			if existing != nil && existing.Action == plan.ActionApprove {
				a.stepList.SetComment(step.ID, nil)
			} else {
				body := ""
				if existing != nil {
					body = existing.Body
				}
				a.stepList.SetComment(step.ID, &plan.ReviewComment{
					StepID: step.ID,
					Action: plan.ActionApprove,
					Body:   body,
				})
			}
			a.refreshDetail()
		}
	}

	return a, nil
}

func (a *App) handleRightPaneKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keymap.Up):
		a.detail.Viewport().LineUp(1)
	case key.Matches(msg, a.keymap.Down):
		a.detail.Viewport().LineDown(1)
	}
	return a, nil
}

func (a *App) handleCommentMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlS:
		result := a.comment.Result()
		a.stepList.SetComment(a.comment.StepID(), result)
		a.comment.Close()
		a.mode = ModeNormal
		a.refreshDetail()
		return a, nil

	case tea.KeyEsc:
		a.comment.Close()
		a.mode = ModeNormal
		return a, nil
	}

	cmd := a.comment.Update(msg)
	return a, cmd
}

func (a *App) handleConfirmMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		a.result.Status = plan.StatusCancelled
		return a, tea.Quit
	case "n", "N":
		a.mode = ModeNormal
		return a, nil
	}
	switch msg.Type {
	case tea.KeyEsc:
		a.mode = ModeNormal
		return a, nil
	case tea.KeyCtrlC:
		a.result.Status = plan.StatusCancelled
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) handleHelpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		a.mode = ModeNormal
		return a, nil
	}
	switch {
	case key.Matches(msg, a.keymap.Help):
		a.mode = ModeNormal
	case msg.String() == "enter", msg.String() == "q":
		a.mode = ModeNormal
	}
	return a, nil
}

func (a *App) submitReview() (tea.Model, tea.Cmd) {
	review := a.stepList.BuildReviewResult()

	hasNonApprove := false
	for _, c := range review.Comments {
		if c.Action != plan.ActionApprove {
			hasNonApprove = true
			break
		}
	}

	if len(review.Comments) == 0 || !hasNonApprove {
		a.result.Status = plan.StatusApproved
	} else {
		a.result.Status = plan.StatusSubmitted
	}
	a.result.Review = review

	return a, tea.Quit
}

func (a *App) refreshDetail() {
	if a.detail == nil {
		return
	}

	if a.stepList.IsOverviewSelected() {
		a.detail.ShowOverview(a.plan)
		return
	}

	if step := a.stepList.Selected(); step != nil {
		comment := a.stepList.GetComment(step.ID)
		a.detail.ShowStep(step, comment)
	}
}

func (a *App) renderTitleBar() string {
	if a.width == 0 {
		return ""
	}

	innerWidth := a.width - 2
	var parts []string
	if a.plan.Title != "" {
		parts = append(parts, a.plan.Title)
	}
	if a.opts.FilePath != "" {
		parts = append(parts, "("+a.opts.FilePath+")")
	}

	if len(parts) == 0 {
		return ""
	}

	content := a.styles.Title.Render(strings.Join(parts, " "))
	return a.styles.InactiveBorder.
		Width(innerWidth).
		Render(content)
}

func (a *App) titleBarHeight() int {
	tb := a.renderTitleBar()
	if tb == "" {
		return 0
	}
	return lipgloss.Height(tb)
}

func (a *App) contentHeight() int {
	h := a.height - a.titleBarHeight() - 3
	if h < 1 {
		h = 1
	}
	return h
}

func (a *App) leftWidth() int {
	if a.width < 80 {
		return a.width - 2
	}
	return a.width * 30 / 100
}

func (a *App) rightWidth() int {
	if a.width < 80 {
		return a.width - 2
	}
	return a.width - a.leftWidth() - 4
}

func (a *App) updateLayout() {
	ch := a.contentHeight()
	rw := a.rightWidth()

	if a.detail == nil {
		a.detail = NewDetailPane(rw, ch, a.opts.Theme)
	} else {
		a.detail.SetSize(rw, ch)
	}

	a.comment.SetWidth(rw - 2)
}

// View implements tea.Model.
func (a *App) View() string {
	if !a.ready {
		return "Loading..."
	}

	// Full-screen overlay modes
	switch a.mode {
	case ModeHelp:
		return a.renderHelp()
	case ModeConfirm:
		return a.renderConfirm()
	}

	ch := a.contentHeight()
	lw := a.leftWidth()
	rw := a.rightWidth()
	singlePane := a.width < 80

	// Left pane: clip content BEFORE applying border
	leftContent := clipLines(a.stepList.Render(lw, ch, a.styles), ch)
	leftBorder := a.styles.InactiveBorder
	if a.focus == FocusLeft {
		leftBorder = a.styles.ActiveBorder
	}
	leftPane := leftBorder.
		Width(lw).
		Height(ch).
		Render(leftContent)

	// Title bar (full width, above panes)
	titleBar := a.renderTitleBar()

	if singlePane {
		var pane string
		if a.focus == FocusRight {
			rightContent := clipLines(a.renderRightContent(rw, ch), ch)
			rightBorder := a.styles.ActiveBorder
			pane = rightBorder.Width(rw).Height(ch).Render(rightContent)
		} else {
			pane = leftPane
		}
		if titleBar != "" {
			return titleBar + "\n" + pane + "\n" + a.renderStatusBar()
		}
		return pane + "\n" + a.renderStatusBar()
	}

	// Right pane: clip content BEFORE applying border
	rightContent := clipLines(a.renderRightContent(rw, ch), ch)
	rightBorder := a.styles.InactiveBorder
	if a.focus == FocusRight {
		rightBorder = a.styles.ActiveBorder
	}
	rightPane := rightBorder.
		Width(rw).
		Height(ch).
		Render(rightContent)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	if titleBar != "" {
		return titleBar + "\n" + content + "\n" + a.renderStatusBar()
	}
	return content + "\n" + a.renderStatusBar()
}

func (a *App) renderRightContent(width, height int) string {
	if a.mode == ModeComment {
		commentHeight := 7
		detailHeight := height - commentHeight - 2
		if detailHeight < 1 {
			detailHeight = 1
		}

		a.detail.SetSize(width, detailHeight)
		detailView := a.detail.View()

		separator := a.styles.CommentBorder.Width(width - 2).Render("Comment")
		commentView := a.comment.View()

		return detailView + "\n" + separator + "\n" + commentView
	}

	return a.detail.View()
}

func (a *App) renderStatusBar() string {
	if a.mode == ModeComment {
		return a.styles.StatusBar.Render(
			a.styles.StatusKey.Render("ctrl+s") + " save  " +
				a.styles.StatusKey.Render("esc") + " cancel",
		)
	}

	return a.styles.StatusBar.Render(
		a.styles.StatusKey.Render("j/k") + " navigate  " +
			a.styles.StatusKey.Render("gg/G") + " top/bottom  " +
			a.styles.StatusKey.Render("enter") + " toggle  " +
			a.styles.StatusKey.Render("c") + " comment  " +
			a.styles.StatusKey.Render("d") + " delete  " +
			a.styles.StatusKey.Render("a") + " approve  " +
			a.styles.StatusKey.Render("s") + " submit  " +
			a.styles.StatusKey.Render("tab") + " switch  " +
			a.styles.StatusKey.Render("?") + " help  " +
			a.styles.StatusKey.Render("q") + " quit",
	)
}

// renderConfirm renders a full-screen confirmation dialog.
func (a *App) renderConfirm() string {
	dialog := lipgloss.NewStyle().
		Width(40).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("170")).
		Render(
			"You have review comments.\n\nQuit without submitting?\n\n" +
				a.styles.StatusKey.Render("y") + " yes   " +
				a.styles.StatusKey.Render("n") + " no   " +
				a.styles.StatusKey.Render("esc") + " cancel",
		)

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, dialog)
}

// clipLines truncates a string to at most maxLines lines.
func clipLines(s string, maxLines int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}

func (a *App) renderHelp() string {
	help := fmt.Sprintf(`%s

  Navigation:
    j/k, Up/Down   Move cursor up/down
    gg              Go to top
    G               Go to bottom
    Enter           Toggle expand/collapse
    l/Right         Expand step
    h/Left          Collapse step (or go to parent)
    Tab             Switch between left/right pane

  Review:
    c               Add/edit comment on selected step
    d               Toggle delete mark
    a               Toggle approve mark
    s               Submit review

  Comment Editor:
    Ctrl+S          Save comment
    Esc             Cancel editing

  Other:
    ?               Toggle this help
    q, Ctrl+C       Quit

  Press Esc or ? or q to close this help.
`, a.styles.Title.Render("ccplan review - Help"))

	return clipLines(help, a.height)
}
