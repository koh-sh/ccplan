package pane

import (
	"testing"
)

func TestByName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{"wezterm", "wezterm", "*pane.WezTermSpawner"},
		{"tmux returns direct", "tmux", "*pane.DirectSpawner"},
		{"auto", "auto", ""},   // type depends on environment
		{"empty", "", ""},      // type depends on environment
		{"unknown", "foo", ""}, // falls back to AutoDetect
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ByName(tt.input)
			if s == nil {
				t.Fatal("ByName returned nil")
			}
			if tt.wantType != "" {
				got := typeName(s)
				if got != tt.wantType {
					t.Errorf("ByName(%q) type = %s, want %s", tt.input, got, tt.wantType)
				}
			}
		})
	}
}

func TestAutoDetect(t *testing.T) {
	s := AutoDetect()
	if s == nil {
		t.Fatal("AutoDetect returned nil")
	}
	// In CI/test environment without wezterm, should fall back to DirectSpawner
	if s.Name() != "wezterm" && s.Name() != "direct" {
		t.Errorf("AutoDetect name = %s, want wezterm or direct", s.Name())
	}
}

func typeName(s PaneSpawner) string {
	switch s.(type) {
	case *WezTermSpawner:
		return "*pane.WezTermSpawner"
	case *DirectSpawner:
		return "*pane.DirectSpawner"
	default:
		return "unknown"
	}
}
