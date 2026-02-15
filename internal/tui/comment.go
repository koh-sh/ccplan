package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/koh-sh/ccplan/internal/plan"
)

// CommentEditor wraps a textarea for entering review comments.
type CommentEditor struct {
	textarea textarea.Model
	stepID   string
	action   plan.ActionType
	active   bool
}

// NewCommentEditor creates a new CommentEditor.
func NewCommentEditor() *CommentEditor {
	ta := textarea.New()
	ta.Placeholder = "Enter review comment... (Ctrl+S to save, Esc to cancel)"
	ta.ShowLineNumbers = false
	ta.SetHeight(5)
	ta.CharLimit = 0

	return &CommentEditor{
		textarea: ta,
	}
}

// Open opens the comment editor for a step, optionally pre-filling with existing comment.
func (c *CommentEditor) Open(stepID string, existing *plan.ReviewComment) {
	c.stepID = stepID
	c.active = true

	if existing != nil {
		c.action = existing.Action
		c.textarea.SetValue(existing.Body)
	} else {
		c.action = plan.ActionModify
		c.textarea.SetValue("")
	}

	c.textarea.Focus()
}

// Close closes the comment editor.
func (c *CommentEditor) Close() {
	c.active = false
	c.textarea.Blur()
}

// IsActive returns whether the editor is active.
func (c *CommentEditor) IsActive() bool {
	return c.active
}

// StepID returns the step ID being edited.
func (c *CommentEditor) StepID() string {
	return c.stepID
}

// Result returns the review comment from the editor content.
// Returns nil if the body is empty and action is modify.
func (c *CommentEditor) Result() *plan.ReviewComment {
	body := strings.TrimSpace(c.textarea.Value())

	// Parse action prefix from body
	action := c.action
	for _, prefix := range []struct {
		text   string
		action plan.ActionType
	}{
		{"modify:", plan.ActionModify},
		{"delete:", plan.ActionDelete},
		{"approve:", plan.ActionApprove},
		{"insert-after:", plan.ActionInsertAfter},
		{"insert-before:", plan.ActionInsertBefore},
	} {
		if strings.HasPrefix(body, prefix.text) {
			action = prefix.action
			body = strings.TrimSpace(strings.TrimPrefix(body, prefix.text))
			break
		}
	}

	if body == "" && action == plan.ActionModify {
		return nil
	}

	return &plan.ReviewComment{
		StepID: c.stepID,
		Action: action,
		Body:   body,
	}
}

// Update handles tea messages for the textarea.
func (c *CommentEditor) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.textarea, cmd = c.textarea.Update(msg)
	return cmd
}

// View renders the comment editor.
func (c *CommentEditor) View() string {
	return c.textarea.View()
}

// SetWidth sets the width of the textarea.
func (c *CommentEditor) SetWidth(w int) {
	c.textarea.SetWidth(w)
}

// SetAction sets the action type for the comment.
func (c *CommentEditor) SetAction(action plan.ActionType) {
	c.action = action
}
