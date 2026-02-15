package pane

import (
	"testing"
)

func TestWezTermSpawnerName(t *testing.T) {
	w := &WezTermSpawner{}
	if w.Name() != "wezterm" {
		t.Errorf("name = %s, want wezterm", w.Name())
	}
}

func TestWezTermSpawnerAvailable(t *testing.T) {
	w := &WezTermSpawner{}
	// Just verify it doesn't panic; result depends on environment
	_ = w.Available()
}
