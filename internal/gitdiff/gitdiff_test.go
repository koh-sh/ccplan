package gitdiff

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// git runs a git command in dir, failing the test on error.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// initRepo creates a repo with one commit containing the given files.
func initRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q")
	for name, content := range files {
		writeFile(t, dir, name, content)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "--no-verify", "-m", "chore: init")
	return dir
}

func TestOpen(t *testing.T) {
	tests := []struct {
		name    string
		dir     func(t *testing.T) string
		wantErr bool
	}{
		{name: "git repo", dir: func(t *testing.T) string { return initRepo(t, map[string]string{"a.md": "# A\n"}) }},
		{name: "plain directory", dir: func(t *testing.T) string { return t.TempDir() }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Open(tt.dir(t))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Open() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyRef(t *testing.T) {
	dir := initRepo(t, map[string]string{"a.md": "# A\n"})
	repo, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		ref     string
		wantErr bool
	}{
		{name: "HEAD", ref: "HEAD"},
		{name: "unknown ref", ref: "no-such-ref", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.VerifyRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyRef(%q) error = %v, wantErr %v", tt.ref, err, tt.wantErr)
			}
		})
	}
}

func TestChangedMarkdownFiles(t *testing.T) {
	dir := initRepo(t, map[string]string{
		"unchanged.md":     "# Same\n",
		"modified.md":      "# Old\n",
		"deleted.md":       "# Gone\n",
		"docs/nested.md":   "# Nested\n",
		"code.go":          "package x\n",
		"sub/inner.md":     "# Inner\n",
		"sub/untouched.md": "# Untouched\n",
	})
	writeFile(t, dir, "modified.md", "# New\n")
	writeFile(t, dir, "docs/nested.md", "# Nested\nmore\n")
	writeFile(t, dir, "untracked.md", "# Fresh\n")
	writeFile(t, dir, "code.go", "package y\n")
	writeFile(t, dir, "sub/inner.md", "# Inner changed\n")
	if err := os.Remove(filepath.Join(dir, "deleted.md")); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		dir  string
		want []string
	}{
		{
			name: "from repo root",
			dir:  dir,
			want: []string{"docs/nested.md", "modified.md", "sub/inner.md", "untracked.md"},
		},
		{
			name: "from subdirectory lists only files below it",
			dir:  filepath.Join(dir, "sub"),
			want: []string{"inner.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := Open(tt.dir)
			if err != nil {
				t.Fatal(err)
			}
			got, err := repo.ChangedMarkdownFiles("HEAD")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("ChangedMarkdownFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilePatch(t *testing.T) {
	dir := initRepo(t, map[string]string{
		"unchanged.md":    "# Same\n",
		"modified.md":     "# Title\n\nold line\n",
		"notes[draft].md": "# Bracket\n",
		"notesd.md":       "# Decoy d\n",
		"notesr.md":       "# Decoy r\n",
	})
	writeFile(t, dir, "modified.md", "# Title\n\nnew line\n")
	writeFile(t, dir, "untracked.md", "# Fresh\nbody\n")
	// A glob-looking name must match itself only: "notes[draft].md" as a git
	// pathspec would otherwise match notesd.md and notesr.md too.
	writeFile(t, dir, "notes[draft].md", "# Bracket\nchanged\n")
	writeFile(t, dir, "notesd.md", "# Decoy d\nchanged\n")
	writeFile(t, dir, "notesr.md", "# Decoy r\nchanged\n")
	repo, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		path         string
		wantEmpty    bool
		wantContains []string
		wantMissing  []string
		wantErr      bool
	}{
		{
			name:         "modified file starts at hunk and has +/- lines",
			path:         "modified.md",
			wantContains: []string{"@@ -1,3 +1,3 @@", "-old line", "+new line"},
		},
		{
			name:      "unchanged file yields empty patch",
			path:      "unchanged.md",
			wantEmpty: true,
		},
		{
			name:         "untracked file is all added",
			path:         "untracked.md",
			wantContains: []string{"@@ -0,0 +1,2 @@", "+# Fresh", "+body"},
		},
		{
			name:         "glob characters in the file name are matched literally",
			path:         "notes[draft].md",
			wantContains: []string{"+changed"},
			wantMissing:  []string{"Decoy", "diff --git"},
		},
		{
			name:    "missing file errors",
			path:    "missing.md",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FilePatch("HEAD", tt.path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.wantEmpty {
				if got != "" {
					t.Fatalf("expected empty patch, got %q", got)
				}
				return
			}
			if !strings.HasPrefix(got, "@@") {
				t.Errorf("patch should start at a hunk header, got %q", got)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("patch missing %q:\n%s", want, got)
				}
			}
			for _, missing := range tt.wantMissing {
				if strings.Contains(got, missing) {
					t.Errorf("patch should not contain %q:\n%s", missing, got)
				}
			}
		})
	}
}
