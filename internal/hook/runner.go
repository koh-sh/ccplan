package hook

import (
	"fmt"
	"os"

	"github.com/koh-sh/ccplan/internal/locate"
	"github.com/koh-sh/ccplan/internal/pane"
)

// RunConfig holds configuration for the hook runner.
type RunConfig struct {
	Spawner pane.PaneSpawner
	Theme   string
	NoColor bool
}

// Run executes the hook orchestration flow.
// Returns exitCode: 0 = continue normally, 2 = feedback to Claude.
func Run(input *Input, cfg RunConfig) (int, error) {
	// Early returns
	if input.PermissionMode != "plan" {
		return 0, nil
	}
	if input.StopHookActive {
		return 0, nil
	}
	if os.Getenv("PLAN_REVIEW_SKIP") == "1" {
		return 0, nil
	}

	// Locate plan file
	paths, err := locate.LocatePlanFile(locate.Options{
		TranscriptPath: input.TranscriptPath,
		CWD:            input.CWD,
	})
	if err != nil || len(paths) == 0 {
		return 0, nil
	}
	planFile := paths[0]

	// Prepare temp files for IPC with review subprocess
	reviewFile, err := os.CreateTemp("", "ccplan-review-*.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccplan: failed to create temp review file: %v\n", err)
		return 0, nil
	}
	reviewPath := reviewFile.Name()
	reviewFile.Close()
	defer os.Remove(reviewPath)

	statusFile, err := os.CreateTemp("", "ccplan-status-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccplan: failed to create temp status file: %v\n", err)
		return 0, nil
	}
	statusPath := statusFile.Name()
	statusFile.Close()
	defer os.Remove(statusPath)

	// Resolve ccplan binary path
	executable, err := os.Executable()
	if err != nil {
		executable = "ccplan"
	}

	// Build review args
	args := []string{
		"review",
		"--output", "file",
		"--output-path", reviewPath,
		"--status-path", statusPath,
		"--theme", cfg.Theme,
	}
	if cfg.NoColor {
		args = append(args, "--no-color")
	}
	args = append(args, planFile)

	// Spawn review in pane
	spawner := cfg.Spawner
	err = spawner.SpawnAndWait(executable, args)
	if err != nil {
		// Fallback to direct if not already direct
		if spawner.Name() != "direct" {
			fmt.Fprintf(os.Stderr, "ccplan: %s spawn failed, falling back to direct: %v\n", spawner.Name(), err)
			direct := &pane.DirectSpawner{}
			err = direct.SpawnAndWait(executable, args)
		}
		if err != nil {
			return 0, nil
		}
	}

	// Read results
	statusBytes, err := os.ReadFile(statusPath)
	if err != nil {
		return 0, nil
	}
	status := string(statusBytes)

	if status == "submitted" {
		reviewBytes, err := os.ReadFile(reviewPath)
		if err != nil {
			return 0, nil
		}
		review := string(reviewBytes)
		if review != "" {
			fmt.Fprint(os.Stderr, review)
			return 2, nil
		}
	}

	return 0, nil
}
