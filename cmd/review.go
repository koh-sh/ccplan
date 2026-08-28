package cmd

import (
	"cmp"
	"errors"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/koh-sh/commd/internal/diff"
	"github.com/koh-sh/commd/internal/gitdiff"
	"github.com/koh-sh/commd/internal/markdown"
	"github.com/koh-sh/commd/internal/tui"
)

// defaultBase is the git ref --diff compares against unless --base is given.
const defaultBase = "HEAD"

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

// Validate requires --output-path when --output=file, exactly one <file>
// unless --diff is set, and rejects --base and --track-viewed combinations
// that only make sense in one of the two modes.
func (r *ReviewCmd) Validate() error {
	if r.Output == "file" && r.OutputPath == "" {
		return fmt.Errorf("--output-path is required with --output file")
	}
	if r.Diff {
		if r.TrackViewed {
			return fmt.Errorf("--track-viewed cannot be combined with --diff")
		}
		return nil
	}
	if r.Base != "" {
		return fmt.Errorf("--base requires --diff")
	}
	if len(r.Files) != 1 {
		return fmt.Errorf("expected exactly one <file> (use --diff to review changed files)")
	}
	return nil
}

// Run executes the review subcommand.
func (r *ReviewCmd) Run() error {
	if r.Diff {
		return r.runDiff()
	}
	return r.runFile(r.Files[0])
}

// runFile reviews a single file in the rendered view.
func (r *ReviewCmd) runFile(path string) error {
	// Read file
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	// Parse document
	p, err := markdown.Parse(source)
	if err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}

	// Create and run TUI
	app := tui.NewApp(p, tui.AppOptions{
		Theme:       r.Theme,
		FilePath:    path,
		TrackViewed: r.TrackViewed,
	})
	result, err := runReviewApp(app, r.teaOpts)
	if err != nil {
		return err
	}

	// Save viewed state if tracking is enabled
	if r.TrackViewed {
		if vs := app.ViewedState(); vs != nil {
			statePath := markdown.StatePath(path)
			if err := markdown.SaveViewedState(statePath, vs); err != nil {
				fmt.Fprintf(os.Stderr, "commd: warning: failed to save viewed state: %v\n", err)
			}
		}
	}

	// Output review if submitted
	if result.Status == markdown.StatusSubmitted && result.Review != nil {
		output := markdown.FormatReview(result.Review, p, path)
		if output == "" {
			return nil
		}

		if err := writeReviewOutput(output, r.Output, r.OutputPath); err != nil {
			return err
		}
	} else if result.Status == markdown.StatusApproved {
		fmt.Fprintln(os.Stderr, "Approved.")
	}

	return nil
}

// runDiff reviews local git changes to Markdown files in the diff view.
// Without explicit files, changed .md files are offered in a picker. Each
// selected file is reviewed in turn and the comments are combined into one
// output.
func (r *ReviewCmd) runDiff() error {
	base := cmp.Or(r.Base, defaultBase)
	repo, err := gitdiff.Open(".")
	if err != nil {
		return err
	}
	if err := repo.VerifyRef(base); err != nil {
		return err
	}

	paths := r.Files
	if len(paths) == 0 {
		changed, err := repo.ChangedMarkdownFiles(base)
		if err != nil {
			return err
		}
		if len(changed) == 0 {
			fmt.Fprintf(os.Stderr, "No changed Markdown files vs %s.\n", base)
			return nil
		}
		paths, err = pickFiles(changed, r.teaOpts)
		if err != nil || len(paths) == 0 {
			return err
		}
	}

	multi := len(paths) > 1
	var reviews []markdown.FileReview
	for _, path := range paths {
		patch, err := repo.FilePatch(base, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			continue
		}
		if patch == "" {
			fmt.Fprintf(os.Stderr, "No changes in %s vs %s.\n", path, base)
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: reading %s: %v\n", path, err)
			continue
		}
		doc, err := markdown.Parse(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: parsing %s: %v\n", path, err)
			continue
		}

		app := tui.NewApp(doc, tui.AppOptions{
			Theme:     r.Theme,
			FilePath:  path,
			MultiFile: multi,
			Diff:      tui.NewDiffData(diff.ParsePatch(patch)),
		})
		result, err := runReviewApp(app, r.teaOpts)
		if err != nil {
			return err
		}
		// Submitted or Approved = done with this file, Cancelled = skipped
		if result.Status == markdown.StatusCancelled {
			continue
		}
		reviews = append(reviews, markdown.FileReview{Path: path, Doc: doc, Review: result.Review})
	}

	if len(reviews) == 0 {
		return nil
	}

	var output string
	if len(reviews) == 1 {
		output = markdown.FormatReview(reviews[0].Review, reviews[0].Doc, reviews[0].Path)
	} else {
		output = markdown.FormatReviews(reviews)
	}
	if output == "" {
		fmt.Fprintln(os.Stderr, "Approved.")
		return nil
	}
	return writeReviewOutput(output, r.Output, r.OutputPath)
}
