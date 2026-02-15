package cmd

import (
	"fmt"
	"os"

	"github.com/koh-sh/ccplan/internal/locate"
)

// Run executes the locate subcommand.
func (l *LocateCmd) Run() error {
	opts := locate.Options{
		TranscriptPath: l.Transcript,
		CWD:            l.CWD,
		All:            l.All,
	}

	// If --stdin, read hook input from stdin
	if l.Stdin {
		input, err := locate.ParseHookInput(os.Stdin)
		if err != nil {
			return fmt.Errorf("parsing stdin: %w", err)
		}
		opts.TranscriptPath = input.TranscriptPath
		if input.CWD != "" {
			opts.CWD = input.CWD
		}
	}

	if opts.TranscriptPath == "" {
		return fmt.Errorf("--transcript or --stdin is required")
	}

	paths, err := locate.LocatePlanFile(opts)
	if err != nil {
		return fmt.Errorf("locating plan file: %w", err)
	}

	if len(paths) == 0 {
		plansDir := locate.ResolvePlansDir(opts.CWD)
		return fmt.Errorf("no plan file found (plansDirectory: %s)", plansDir)
	}

	for _, p := range paths {
		fmt.Println(p)
	}

	return nil
}
