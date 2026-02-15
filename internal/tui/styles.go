package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds all the lipgloss styles for the TUI.
type Styles struct {
	// Panes
	LeftPane       lipgloss.Style
	RightPane      lipgloss.Style
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style

	// Step list
	Title          lipgloss.Style
	SelectedStep   lipgloss.Style
	NormalStep     lipgloss.Style
	StepBadge      lipgloss.Style
	DeleteBadge    lipgloss.Style
	ApproveBadge   lipgloss.Style

	// Status bar
	StatusBar      lipgloss.Style
	StatusKey      lipgloss.Style

	// Comment
	CommentBorder  lipgloss.Style

	// Help
	HelpStyle      lipgloss.Style
}

// DefaultStyles returns the default dark theme styles.
func DefaultStyles() Styles {
	return Styles{
		LeftPane: lipgloss.NewStyle().
			Padding(0, 1),
		RightPane: lipgloss.NewStyle().
			Padding(0, 1),
		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")),
		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")),
		SelectedStep: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		NormalStep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		StepBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")),
		DeleteBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")),
		ApproveBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")),
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		StatusKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true),
		CommentBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")),
		HelpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
	}
}
