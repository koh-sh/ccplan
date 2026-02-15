package pane

import (
	"context"
	"testing"
)

func TestDirectSpawnerAvailable(t *testing.T) {
	d := &DirectSpawner{}
	if !d.Available() {
		t.Error("DirectSpawner should always be available")
	}
}

func TestDirectSpawnerName(t *testing.T) {
	d := &DirectSpawner{}
	if d.Name() != "direct" {
		t.Errorf("name = %s, want direct", d.Name())
	}
}

func TestDirectSpawnerSpawnSuccess(t *testing.T) {
	d := &DirectSpawner{}
	err := d.SpawnAndWait(context.Background(), "true", nil)
	if err != nil {
		t.Errorf("SpawnAndWait(true) error = %v", err)
	}
}

func TestDirectSpawnerSpawnFailure(t *testing.T) {
	d := &DirectSpawner{}
	err := d.SpawnAndWait(context.Background(), "nonexistent-command-that-does-not-exist", nil)
	if err == nil {
		t.Error("SpawnAndWait with nonexistent command should return error")
	}
}
