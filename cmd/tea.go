package cmd

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/koh-sh/commd/internal/tui"
)

// runTea creates and runs a Bubble Tea program with alt screen and optional extra options.
// Alt screen mode is set via the View.AltScreen field in each model's View() method.
func runTea(model tea.Model, extraOpts []tea.ProgramOption) (tea.Model, error) {
	// Reset Line Feed/New Line Mode (LNM) so bare LF moves cursor down without
	// resetting the column. The renderer's cursor-down-and-back diff updates
	// rely on this; some terminal emulators default LNM to on, which makes
	// those updates land at the wrong column. VT100-compliant terminals already
	// have LNM off, so this is a no-op there.
	fmt.Print("\x1b[20l")
	return tea.NewProgram(model, extraOpts...).Run()
}

// runReviewApp runs the review TUI for one file and returns its result.
func runReviewApp(app *tui.App, teaOpts []tea.ProgramOption) (tui.AppResult, error) {
	finalModel, err := runTea(app, teaOpts)
	if err != nil {
		return tui.AppResult{}, fmt.Errorf("running TUI: %w", err)
	}
	reviewApp, ok := finalModel.(*tui.App)
	if !ok {
		return tui.AppResult{}, fmt.Errorf("unexpected model type: %T", finalModel)
	}
	return reviewApp.Result(), nil
}

// pickFiles shows the multi-select file picker and returns the chosen paths.
// Returns nil without error when the user cancels or selects nothing.
func pickFiles(paths []string, teaOpts []tea.ProgramOption) ([]string, error) {
	picker := tui.NewFilePicker(paths)
	finalModel, err := runTea(picker, teaOpts)
	if err != nil {
		return nil, fmt.Errorf("running file picker: %w", err)
	}
	fp, ok := finalModel.(*tui.FilePicker)
	if !ok {
		return nil, fmt.Errorf("unexpected model type: %T", finalModel)
	}
	result := fp.Result()
	if result.Cancelled {
		return nil, nil
	}
	return result.SelectedFiles, nil
}
