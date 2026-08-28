package cchook

import (
	"context"
	"fmt"
	"os"

	"github.com/koh-sh/commd/internal/cclocate"
	"github.com/koh-sh/commd/internal/pane"
)

// RunConfig holds configuration for the hook runner.
type RunConfig struct {
	Spawner pane.PaneSpawner
	Theme   string
}

// Run executes the hook orchestration flow.
// Returns exitCode: 0 = continue normally, 2 = feedback to Claude.
func Run(input *Input, cfg RunConfig) (int, error) {
	// Early returns
	if input.PermissionMode != permissionModePlan {
		return 0, nil
	}
	if os.Getenv("CC_PLAN_REVIEW_SKIP") == "1" {
		return 0, nil
	}

	planFile, ok := resolvePlanFile(input)
	if !ok {
		return 0, nil
	}

	// Check file exists
	if _, err := os.Stat(planFile); err != nil {
		return 0, nil
	}

	// Prepare temp file for IPC with review subprocess
	reviewFile, err := os.CreateTemp("", "commd-review-*.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "commd: failed to create temp review file: %v\n", err)
		return 0, nil
	}
	reviewPath := reviewFile.Name()
	reviewFile.Close()
	defer os.Remove(reviewPath)

	// Resolve commd binary path
	executable, err := os.Executable()
	if err != nil {
		executable = "commd"
	}

	// Build review args
	args := []string{
		"review",
		"--output", "file",
		"--output-path", reviewPath,
		"--theme", cfg.Theme,
		"--track-viewed",
		planFile,
	}

	// Spawn review in pane
	ctx := context.Background()
	spawner := cfg.Spawner
	err = spawner.SpawnAndWait(ctx, executable, args)
	if err != nil {
		// Fallback to direct if not already direct
		if spawner.Name() != pane.NameDirect {
			fmt.Fprintf(os.Stderr, "commd: %s spawn failed, falling back to direct: %v\n", spawner.Name(), err)
			direct := &pane.DirectSpawner{}
			err = direct.SpawnAndWait(ctx, executable, args)
		}
		if err != nil {
			return 0, nil
		}
	}

	// Read review result -- non-empty means submitted
	reviewBytes, err := os.ReadFile(reviewPath)
	if err != nil {
		return 0, nil
	}
	if review := string(reviewBytes); review != "" {
		fmt.Fprint(os.Stderr, review)
		return 2, nil
	}

	return 0, nil
}

// resolvePlanFile determines which plan file to review from the hook input.
//
// PreToolUse on ExitPlanMode: Claude Code injects planFilePath, which is
// authoritative, so no plansDirectory lookup is needed.
//
// PostToolUse on Write/Edit: the written file must be under plansDirectory to
// count as a plan.
//
// Returns false when the input does not identify a plan file.
func resolvePlanFile(input *Input) (string, bool) {
	if input.ToolInput == nil {
		return "", false
	}
	if input.HookEventName == eventPreToolUse && input.ToolName == toolExitPlanMode {
		return input.ToolInput.PlanFilePath, input.ToolInput.PlanFilePath != ""
	}
	planFile := input.ToolInput.FilePath
	if planFile == "" {
		return "", false
	}
	plansDir := cclocate.ResolvePlansDir(input.CWD)
	if !cclocate.IsUnderDir(planFile, plansDir) {
		return "", false
	}
	return planFile, true
}
