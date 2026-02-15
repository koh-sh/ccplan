package cmd

import (
	"fmt"
	"os"

	"github.com/koh-sh/ccplan/internal/hook"
	"github.com/koh-sh/ccplan/internal/pane"
)

// Run executes the hook subcommand.
func (h *HookCmd) Run() error {
	input, err := hook.ParseInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccplan hook: failed to parse input: %v\n", err)
		os.Exit(0)
	}

	spawner := pane.ByName(h.Spawner)

	exitCode, err := hook.Run(input, hook.RunConfig{
		Spawner: spawner,
		Theme:   h.Theme,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccplan hook: %v\n", err)
		os.Exit(0)
	}

	os.Exit(exitCode)
	return nil // unreachable
}
