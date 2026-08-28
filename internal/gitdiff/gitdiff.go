// Package gitdiff reads local working-tree changes from git for the
// `commd review --diff` flow. It shells out to the git binary; paths are
// always relative to the directory the Repo was opened in.
package gitdiff

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/koh-sh/commd/internal/diff"
)

// markdownPathspec matches .md files at any depth (git globs cross '/').
const markdownPathspec = "*.md"

// Repo runs git commands inside a working tree directory.
type Repo struct {
	dir string
}

// Open verifies that dir is inside a git working tree.
func Open(dir string) (*Repo, error) {
	r := &Repo{dir: dir}
	if _, err := r.run("rev-parse", "--show-toplevel"); err != nil {
		return nil, fmt.Errorf("not a git repository: %s: %w", dir, err)
	}
	return r, nil
}

// VerifyRef checks that base resolves to a commit.
func (r *Repo) VerifyRef(base string) error {
	if _, err := r.run("rev-parse", "--verify", "--quiet", base+"^{commit}"); err != nil {
		return fmt.Errorf("unknown git ref %q: %w", base, err)
	}
	return nil
}

// literalPathspec makes git match path exactly instead of treating '*', '?'
// and '[...]' in the file name as glob patterns.
func literalPathspec(path string) string {
	return ":(literal)" + path
}

// ChangedMarkdownFiles lists .md files under the repo directory that differ
// from base in the working tree, plus untracked .md files. Deleted files are
// excluded. Paths are relative to the directory Open was called with.
func (r *Repo) ChangedMarkdownFiles(base string) ([]string, error) {
	tracked, err := r.runList("diff", "--name-only", "--relative", "-z", "--diff-filter=d", base, "--", markdownPathspec)
	if err != nil {
		return nil, err
	}
	untracked, err := r.untrackedFiles(markdownPathspec)
	if err != nil {
		return nil, err
	}
	files := slices.Concat(tracked, untracked)
	slices.Sort(files)
	return slices.Compact(files), nil
}

// FilePatch returns the unified diff of path against base, starting at the
// first hunk header so diff.ParsePatch can consume it. Untracked files yield
// an all-added patch. Returns "" when the file has no changes.
func (r *Repo) FilePatch(base, path string) (string, error) {
	// git reports nothing for an unknown pathspec, so check existence first
	// to give a clear error instead of "no changes".
	fullPath := filepath.Join(r.dir, path)
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}

	untracked, err := r.untrackedFiles(literalPathspec(path))
	if err != nil {
		return "", err
	}
	if len(untracked) > 0 {
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", path, err)
		}
		return diff.AddedFilePatch(content), nil
	}

	out, err := r.run("diff", "--no-color", "--no-ext-diff", "--unified=3", base, "--", literalPathspec(path))
	if err != nil {
		return "", err
	}
	return diff.StripHeader(string(out)), nil
}

// untrackedFiles lists untracked, non-ignored files matching pathspec.
func (r *Repo) untrackedFiles(pathspec string) ([]string, error) {
	return r.runList("ls-files", "--others", "--exclude-standard", "-z", "--", pathspec)
}

// runList runs a git command whose output is NUL-separated paths.
func (r *Repo) runList(args ...string) ([]string, error) {
	out, err := r.run(args...)
	if err != nil {
		return nil, err
	}
	var files []string
	for f := range bytes.SplitSeq(out, []byte{0}) {
		if len(f) > 0 {
			files = append(files, string(f))
		}
	}
	return files, nil
}

// run executes git with the given arguments in the repo directory.
func (r *Repo) run(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	// Avoid taking index locks for read-only queries.
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && msg != "" {
			return nil, fmt.Errorf("git %s: %s: %w", args[0], msg, err)
		}
		return nil, fmt.Errorf("git %s: %w", args[0], err)
	}
	return out, nil
}
