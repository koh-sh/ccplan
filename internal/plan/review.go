package plan

import (
	"fmt"
	"strings"
)

// FormatReview formats a ReviewResult as a Markdown string.
// Format: ## {StepID}: {StepTitle} [{ActionType}]
func FormatReview(result *ReviewResult, p *Plan, filePath string) string {
	if len(result.Comments) == 0 {
		return ""
	}

	target := filePath
	if target == "" {
		target = "the file"
	}

	var sb strings.Builder
	sb.WriteString("# Plan Review\n\n")
	sb.WriteString(fmt.Sprintf("Please review and address the following comments on: %s\n", target))

	for _, c := range result.Comments {
		step := p.FindStep(c.StepID)
		title := c.StepID
		if step != nil {
			title = fmt.Sprintf("%s: %s", c.StepID, step.Title)
		}

		sb.WriteString(fmt.Sprintf("\n## %s [%s]\n", title, c.Action))
		if c.Body != "" {
			sb.WriteString(c.Body + "\n")
		}
	}

	return sb.String()
}
