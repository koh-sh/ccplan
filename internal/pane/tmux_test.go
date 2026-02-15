package pane

import (
	"testing"
)

func TestTmuxSpawnerAvailable(t *testing.T) {
	s := &TmuxSpawner{}
	if s.Available() {
		t.Error("TmuxSpawner should not be available (not implemented)")
	}
}

func TestTmuxSpawnerName(t *testing.T) {
	s := &TmuxSpawner{}
	if s.Name() != "tmux" {
		t.Errorf("name = %s, want tmux", s.Name())
	}
}

func TestTmuxSpawnerSpawnError(t *testing.T) {
	s := &TmuxSpawner{}
	err := s.SpawnAndWait("echo", []string{"hello"})
	if err == nil {
		t.Error("SpawnAndWait should return error")
	}
}
