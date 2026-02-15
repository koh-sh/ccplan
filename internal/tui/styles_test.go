package tui

import (
	"testing"
)

func TestStylesForTheme(t *testing.T) {
	tests := []struct {
		theme string
	}{
		{"dark"},
		{"light"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.theme, func(t *testing.T) {
			s := stylesForTheme(tt.theme)
			// Just verify it returns a valid Styles struct (not zero value)
			if s.Title.GetBold() != true {
				t.Error("Title style should be bold")
			}
		})
	}
}

func TestDefaultStyles(t *testing.T) {
	s := DefaultStyles()
	if s.Title.GetBold() != true {
		t.Error("DefaultStyles Title should be bold")
	}
}
