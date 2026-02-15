package cmd

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/koh-sh/ccplan/internal/plan"
	"github.com/koh-sh/ccplan/internal/tui"
)

// Run executes the review subcommand.
func (r *ReviewCmd) Run() error {
	// Read plan file
	source, err := os.ReadFile(r.PlanFile)
	if err != nil {
		return fmt.Errorf("reading plan file: %w", err)
	}

	// Parse plan
	p, err := plan.Parse(source)
	if err != nil {
		return fmt.Errorf("parsing plan file: %w", err)
	}

	// Create and run TUI
	app := tui.NewApp(p)
	prog := tea.NewProgram(app, tea.WithAltScreen())
	finalModel, err := prog.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	result := finalModel.(*tui.App).Result()

	// Write status file if requested
	if r.StatusPath != "" {
		if err := os.WriteFile(r.StatusPath, []byte(string(result.Status)), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to write status file: %v\n", err)
		}
	}

	// Output review if submitted
	if result.Status == plan.StatusSubmitted && result.Review != nil {
		output := plan.FormatReview(result.Review, p)
		if output == "" {
			return nil
		}

		switch r.Output {
		case "clipboard":
			if err := clipboard.WriteAll(output); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to copy to clipboard: %v\n", err)
				fmt.Fprintf(os.Stderr, "Use --output stdout or --output file instead.\n")
				// Still print to stdout as fallback
				fmt.Print(output)
			} else {
				fmt.Fprintln(os.Stderr, "Review copied to clipboard.")
			}
		case "stdout":
			fmt.Print(output)
		case "file":
			if r.OutputPath == "" {
				return fmt.Errorf("--output-path is required with --output file")
			}
			if err := os.WriteFile(r.OutputPath, []byte(output), 0o644); err != nil {
				return fmt.Errorf("writing output file: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Review written to %s\n", r.OutputPath)
		}
	} else if result.Status == plan.StatusApproved {
		fmt.Fprintln(os.Stderr, "Plan approved.")
	}

	return nil
}
