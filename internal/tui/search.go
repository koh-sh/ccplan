package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SearchBar wraps a textinput for step searching.
type SearchBar struct {
	input  textinput.Model
	active bool
}

// NewSearchBar creates a new SearchBar.
func NewSearchBar() *SearchBar {
	ti := textinput.New()
	ti.Prompt = "/"
	ti.CharLimit = 100

	return &SearchBar{
		input: ti,
	}
}

// Open activates the search bar.
func (s *SearchBar) Open() {
	s.active = true
	s.input.SetValue("")
	s.input.Focus()
}

// Close deactivates the search bar.
func (s *SearchBar) Close() {
	s.active = false
	s.input.Blur()
}

// IsActive returns whether the search bar is active.
func (s *SearchBar) IsActive() bool {
	return s.active
}

// Query returns the current search query.
func (s *SearchBar) Query() string {
	return s.input.Value()
}

// Update handles tea messages for the textinput.
func (s *SearchBar) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	return cmd
}

// View renders the search bar.
func (s *SearchBar) View() string {
	return s.input.View()
}
