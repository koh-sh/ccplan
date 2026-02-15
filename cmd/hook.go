package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/koh-sh/ccplan/internal/hook"
	"github.com/koh-sh/ccplan/internal/pane"
)

// Run executes the hook subcommand.
func (h *HookCmd) Run() error {
	os.Exit(h.runExit(os.Stdin))
	return nil // unreachable
}

// runExit executes the hook logic and returns the exit code.
func (h *HookCmd) runExit(r io.Reader) int {
	input, err := hook.ParseInput(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccplan hook: failed to parse input: %v\n", err)
		return 0
	}

	spawner := pane.ByName(h.Spawner)

	exitCode, err := hook.Run(input, hook.RunConfig{
		Spawner: spawner,
		Theme:   h.Theme,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccplan hook: %v\n", err)
		return 0
	}

	return exitCode
}
