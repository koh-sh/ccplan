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
			if l.Quiet {
				os.Exit(1)
			}
			return fmt.Errorf("parsing stdin: %w", err)
		}
		opts.TranscriptPath = input.TranscriptPath
		if input.CWD != "" {
			opts.CWD = input.CWD
		}
	}

	if opts.TranscriptPath == "" {
		if l.Quiet {
			os.Exit(1)
		}
		return fmt.Errorf("--transcript or --stdin is required")
	}

	paths, err := locate.LocatePlanFile(opts)
	if err != nil {
		if l.Quiet {
			os.Exit(1)
		}
		return fmt.Errorf("locating plan file: %w", err)
	}

	if len(paths) == 0 {
		if l.Quiet {
			os.Exit(1)
		}
		plansDir := locate.ResolvePlansDir(opts.CWD)
		return fmt.Errorf("no plan file found (plansDirectory: %s)", plansDir)
	}

	if !l.Quiet {
		for _, p := range paths {
			fmt.Println(p)
		}
	}

	return nil
}
