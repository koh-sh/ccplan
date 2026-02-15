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
	Title        lipgloss.Style
	SelectedStep lipgloss.Style
	NormalStep   lipgloss.Style
	StepBadge   lipgloss.Style
	ViewedBadge lipgloss.Style

	// Status bar
	StatusBar lipgloss.Style
	StatusKey lipgloss.Style

	// Comment
	CommentBorder lipgloss.Style

	// Help
	HelpStyle lipgloss.Style
}

// stylesForTheme returns styles for the given theme.
// If noColor is true, all color styling is disabled.
func stylesForTheme(theme string, noColor bool) Styles {
	if noColor {
		return plainStyles()
	}
	if theme == "light" {
		return lightStyles()
	}
	return darkStyles()
}

// darkStyles returns the dark theme styles.
func darkStyles() Styles {
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
		ViewedBadge: lipgloss.NewStyle().
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

// lightStyles returns the light theme styles.
func lightStyles() Styles {
	return Styles{
		LeftPane: lipgloss.NewStyle().
			Padding(0, 1),
		RightPane: lipgloss.NewStyle().
			Padding(0, 1),
		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("33")),
		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("250")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("130")),
		SelectedStep: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")),
		NormalStep: lipgloss.NewStyle().
			Foreground(lipgloss.Color("236")),
		StepBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("130")),
		ViewedBadge: lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")),
		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		StatusKey: lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true),
		CommentBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("33")),
		HelpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
	}
}

// plainStyles returns styles with no colors (for --no-color).
func plainStyles() Styles {
	return Styles{
		LeftPane: lipgloss.NewStyle().
			Padding(0, 1),
		RightPane: lipgloss.NewStyle().
			Padding(0, 1),
		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()),
		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()),
		Title:        lipgloss.NewStyle().Bold(true),
		SelectedStep: lipgloss.NewStyle().Bold(true),
		NormalStep:   lipgloss.NewStyle(),
		StepBadge:    lipgloss.NewStyle(),
		ViewedBadge: lipgloss.NewStyle(),
		StatusBar:    lipgloss.NewStyle(),
		StatusKey:    lipgloss.NewStyle().Bold(true),
		CommentBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()),
		HelpStyle: lipgloss.NewStyle(),
	}
}

// DefaultStyles returns the default dark theme styles.
func DefaultStyles() Styles {
	return darkStyles()
}
