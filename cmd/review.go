package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/koh-sh/ccplan/internal/plan"
	"github.com/koh-sh/ccplan/internal/tui"
)

// writeReviewOutput writes the review output using the specified method.
func writeReviewOutput(output, mode, outputPath string) error {
	switch mode {
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
		if outputPath == "" {
			return fmt.Errorf("--output-path is required with --output file")
		}
		if _, err := os.Stat(outputPath); errors.Is(err, os.ErrNotExist) {
			// Output file was deleted (possibly due to hook timeout). Fall back to clipboard.
			if err := clipboard.WriteAll(output); err != nil {
				fmt.Fprintf(os.Stderr, "Output file %s was deleted (possibly due to hook timeout). Failed to copy to clipboard: %v\n", outputPath, err)
				fmt.Print(output)
			} else {
				fmt.Fprintf(os.Stderr, "Output file %s was deleted (possibly due to hook timeout). Review copied to clipboard.\n", outputPath)
			}
		} else if err := os.WriteFile(outputPath, []byte(output), 0o644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		} else {
			fmt.Fprintf(os.Stderr, "Review written to %s\n", outputPath)
		}
	}
	return nil
}

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
	app := tui.NewApp(p, tui.AppOptions{
		Theme:       r.Theme,
		FilePath:    r.PlanFile,
		TrackViewed: r.TrackViewed,
	})
	prog := tea.NewProgram(app, tea.WithAltScreen())
	finalModel, err := prog.Run()
	if err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	app, ok := finalModel.(*tui.App)
	if !ok {
		return fmt.Errorf("unexpected model type: %T", finalModel)
	}

	// Save viewed state if tracking is enabled
	if r.TrackViewed {
		if vs := app.ViewedState(); vs != nil {
			statePath := plan.StatePath(r.PlanFile)
			if err := plan.SaveViewedState(statePath, vs); err != nil {
				fmt.Fprintf(os.Stderr, "ccplan: warning: failed to save viewed state: %v\n", err)
			}
		}
	}

	result := app.Result()

	// Output review if submitted
	if result.Status == plan.StatusSubmitted && result.Review != nil {
		output := plan.FormatReview(result.Review, p, r.PlanFile)
		if output == "" {
			return nil
		}

		if err := writeReviewOutput(output, r.Output, r.OutputPath); err != nil {
			return err
		}
	} else if result.Status == plan.StatusApproved {
		fmt.Fprintln(os.Stderr, "Plan approved.")
	}

	return nil
}
