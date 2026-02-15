package hook

import (
	"os"
	"path/filepath"
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

func TestRunSkipsPlanReviewSkipEnv(t *testing.T) {
	t.Setenv("PLAN_REVIEW_SKIP", "1")
	mock := &mockSpawner{available: true, name: "mock"}
	input := &Input{PermissionMode: "plan"}
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

func TestRunSkipsNilToolInput(t *testing.T) {
	mock := &mockSpawner{available: true, name: "mock"}
	input := &Input{
		PermissionMode: "plan",
		ToolInput:      nil,
	}
	code, err := Run(input, RunConfig{Spawner: mock})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if mock.spawnCalled {
		t.Error("spawner should not be called with nil tool_input")
	}
}

func TestRunSkipsEmptyFilePath(t *testing.T) {
	mock := &mockSpawner{available: true, name: "mock"}
	input := &Input{
		PermissionMode: "plan",
		ToolInput:      &ToolInput{FilePath: ""},
	}
	code, err := Run(input, RunConfig{Spawner: mock})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if mock.spawnCalled {
		t.Error("spawner should not be called with empty file_path")
	}
}

func TestRunSkipsNonPlanFile(t *testing.T) {
	// Create a temp file that exists but is NOT under plans directory
	tmpFile, err := os.CreateTemp("", "not-a-plan-*.go")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mock := &mockSpawner{available: true, name: "mock"}
	input := &Input{
		PermissionMode: "plan",
		CWD:            t.TempDir(),
		ToolInput:      &ToolInput{FilePath: tmpFile.Name()},
	}
	code, err := Run(input, RunConfig{Spawner: mock})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if mock.spawnCalled {
		t.Error("spawner should not be called for file outside plans directory")
	}
}

func TestIsUnderPlansDir(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		plansDir string
		want     bool
	}{
		{
			name:     "file under plans dir",
			filePath: "/home/user/.claude/plans/my-plan.md",
			plansDir: "/home/user/.claude/plans",
			want:     true,
		},
		{
			name:     "file not under plans dir",
			filePath: "/home/user/projects/src/main.go",
			plansDir: "/home/user/.claude/plans",
			want:     false,
		},
		{
			name:     "plans dir with trailing separator",
			filePath: "/home/user/.claude/plans/my-plan.md",
			plansDir: "/home/user/.claude/plans/",
			want:     true,
		},
		{
			name:     "similar prefix but different dir",
			filePath: "/home/user/.claude/plans-backup/my-plan.md",
			plansDir: "/home/user/.claude/plans",
			want:     false,
		},
		{
			name:     "empty plans dir",
			filePath: "/home/user/.claude/plans/my-plan.md",
			plansDir: "",
			want:     false,
		},
		{
			name:     "nested file under plans dir",
			filePath: filepath.Join("/home/user/.claude/plans", "subdir", "my-plan.md"),
			plansDir: "/home/user/.claude/plans",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnderPlansDir(tt.filePath, tt.plansDir)
			if got != tt.want {
				t.Errorf("isUnderPlansDir(%q, %q) = %v, want %v", tt.filePath, tt.plansDir, got, tt.want)
			}
		})
	}
}
