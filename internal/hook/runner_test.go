package hook

import (
	"testing"
)

// mockSpawner implements pane.PaneSpawner for testing.
type mockSpawner struct {
	available   bool
	name        string
	spawnFunc   func(cmd string, args []string) error
	spawnCalled bool
}

func (m *mockSpawner) Available() bool { return m.available }
func (m *mockSpawner) Name() string    { return m.name }
func (m *mockSpawner) SpawnAndWait(cmd string, args []string) error {
	m.spawnCalled = true
	if m.spawnFunc != nil {
		return m.spawnFunc(cmd, args)
	}
	return nil
}

func TestRunSkipsNonPlanMode(t *testing.T) {
	mock := &mockSpawner{available: true, name: "mock"}
	input := &Input{PermissionMode: "default"}
	code, err := Run(input, RunConfig{Spawner: mock})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if mock.spawnCalled {
		t.Error("spawner should not be called for non-plan mode")
	}
}

func TestRunSkipsStopHookActive(t *testing.T) {
	mock := &mockSpawner{available: true, name: "mock"}
	input := &Input{PermissionMode: "plan", StopHookActive: true}
	code, err := Run(input, RunConfig{Spawner: mock})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if mock.spawnCalled {
		t.Error("spawner should not be called when stop_hook_active")
	}
}

func TestRunSkipsPlanReviewSkipEnv(t *testing.T) {
	t.Setenv("PLAN_REVIEW_SKIP", "1")
	mock := &mockSpawner{available: true, name: "mock"}
	input := &Input{PermissionMode: "plan", StopHookActive: false}
	code, err := Run(input, RunConfig{Spawner: mock})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if mock.spawnCalled {
		t.Error("spawner should not be called when PLAN_REVIEW_SKIP=1")
	}
}

func TestRunSkipsEmptyTranscriptPath(t *testing.T) {
	mock := &mockSpawner{available: true, name: "mock"}
	input := &Input{
		PermissionMode: "plan",
		StopHookActive: false,
		TranscriptPath: "",
		CWD:            ".",
	}
	code, err := Run(input, RunConfig{Spawner: mock})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if mock.spawnCalled {
		t.Error("spawner should not be called with empty transcript path")
	}
}
