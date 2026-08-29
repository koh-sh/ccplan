package cclocate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePlansDir(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string]string // relative path from cwd → content
		homeFiles map[string]string // relative path from $HOME → content
		wantFn    func(cwd, home string) string
	}{
		{
			name: "default when no settings exist",
			wantFn: func(_, home string) string {
				return filepath.Join(home, ".claude", "plans")
			},
		},
		{
			name: "from settings.local.json",
			files: map[string]string{
				".claude/settings.local.json": `{"plansDirectory": "/custom/plans"}`,
			},
			wantFn: func(_, _ string) string { return "/custom/plans" },
		},
		{
			name: "from settings.json",
			files: map[string]string{
				".claude/settings.json": `{"plansDirectory": "/project/plans"}`,
			},
			wantFn: func(_, _ string) string { return "/project/plans" },
		},
		{
			name: "from ~/.claude/settings.json",
			homeFiles: map[string]string{
				".claude/settings.json": `{"plansDirectory": "/home/plans"}`,
			},
			wantFn: func(_, _ string) string { return "/home/plans" },
		},
		{
			name: "relative path resolved from cwd",
			files: map[string]string{
				".claude/settings.json": `{"plansDirectory": "plans"}`,
			},
			wantFn: func(cwd, _ string) string { return filepath.Join(cwd, "plans") },
		},
		{
			name: "local settings take priority over regular settings",
			files: map[string]string{
				".claude/settings.local.json": `{"plansDirectory": "/local/plans"}`,
				".claude/settings.json":       `{"plansDirectory": "/regular/plans"}`,
			},
			wantFn: func(_, _ string) string { return "/local/plans" },
		},
		{
			name: "project settings take priority over home settings",
			files: map[string]string{
				".claude/settings.json": `{"plansDirectory": "/project/plans"}`,
			},
			homeFiles: map[string]string{
				".claude/settings.json": `{"plansDirectory": "/home/plans"}`,
			},
			wantFn: func(_, _ string) string { return "/project/plans" },
		},
		{
			name: "broken local JSON falls back to valid settings",
			files: map[string]string{
				".claude/settings.local.json": `{broken`,
				".claude/settings.json":       `{"plansDirectory": "/fallback/plans"}`,
			},
			wantFn: func(_, _ string) string { return "/fallback/plans" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwd := t.TempDir()
			// Isolate the last step of the chain from the developer's real
			// ~/.claude/settings.json.
			home := t.TempDir()
			t.Setenv("HOME", home)

			writeFiles(t, cwd, tt.files)
			writeFiles(t, home, tt.homeFiles)

			got := ResolvePlansDir(cwd)
			want := tt.wantFn(cwd, home)
			if got != want {
				t.Errorf("ResolvePlansDir() = %q, want %q", got, want)
			}
		})
	}
}

// writeFiles creates each relative path under root with the given content.
func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for relPath, content := range files {
		absPath := filepath.Join(root, relPath)
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
