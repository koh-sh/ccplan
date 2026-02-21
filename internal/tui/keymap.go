package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the TUI.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Expand      key.Binding
	Collapse    key.Binding
	Toggle      key.Binding
	Tab         key.Binding
	Comment     key.Binding
	CommentList key.Binding
	Viewed      key.Binding
	Search      key.Binding
	Submit      key.Binding
	Quit        key.Binding
	Help        key.Binding
	Save        key.Binding
	Cancel      key.Binding

	// Right pane specific
	ScrollToStart key.Binding
	ScrollToEnd   key.Binding

	// Pane resizing
	PaneGrow   key.Binding
	PaneShrink key.Binding

	// Comment list specific
	Edit   key.Binding
	Delete key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("j/k", "navigate"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("", ""),
		),
		Expand: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("l/h", "expand/collapse"),
		),
		Collapse: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("", ""),
		),
		Toggle: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "toggle"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),
		Comment: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "comment"),
		),
		CommentList: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("C", "manage comments"),
		),
		Viewed: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "viewed"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Submit: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "submit"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save comment"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		ScrollToStart: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "scroll to start"),
		),
		ScrollToEnd: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "scroll to end"),
		),
		PaneGrow: key.NewBinding(
			key.WithKeys(">"),
			key.WithHelp(">/<", "resize pane"),
		),
		PaneShrink: key.NewBinding(
			key.WithKeys("<"),
			key.WithHelp("", ""),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
	}
}
