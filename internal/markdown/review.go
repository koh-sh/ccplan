package markdown

import (
	"fmt"
	"strings"
)

// FormatReview formats a ReviewResult as a Markdown string.
// Section-level comments are grouped under section headings.
// Line-level comments are listed separately by line number.
func FormatReview(result *ReviewResult, d *Document, filePath string) string {
	if len(result.Comments) == 0 {
		return ""
	}

	target := filePath
	if target == "" {
		target = "the file"
	}

	// Separate section-level and line-level comments.
	type group struct {
		title    string
		comments []ReviewComment
	}
	var sectionOrder []string
	sectionGroups := make(map[string]*group)
	var lineComments []ReviewComment

	for _, c := range result.Comments {
		if c.StartLine > 0 {
			lineComments = append(lineComments, c)
			continue
		}
		g, ok := sectionGroups[c.SectionID]
		if !ok {
			var title string
			if c.SectionID == OverviewSectionID {
				title = "Overview"
			} else {
				section := d.FindSection(c.SectionID)
				title = c.SectionID
				if section != nil {
					title = fmt.Sprintf("%s: %s", c.SectionID, section.Title)
				}
			}
			g = &group{title: title}
			sectionGroups[c.SectionID] = g
			sectionOrder = append(sectionOrder, c.SectionID)
		}
		g.comments = append(g.comments, c)
	}

	var sb strings.Builder
	sb.WriteString("# Review\n\n")
	fmt.Fprintf(&sb, "Please review and address the following comments on: %s\n", target)

	// Section-level comments grouped by section
	for _, id := range sectionOrder {
		g := sectionGroups[id]
		fmt.Fprintf(&sb, "\n## %s\n", g.title)
		for _, c := range g.comments {
			fmt.Fprintf(&sb, "[%s] %s\n", c.FormatLabel(), c.Body)
		}
	}

	// Line-level comments with inline code decoration, separated by a divider
	if len(lineComments) > 0 {
		sb.WriteString("\n---\n")
		for _, c := range lineComments {
			fmt.Fprintf(&sb, "\n`%s` [%s] %s\n", c.FormatLineRef(), c.FormatLabel(), c.Body)
		}
	}

	return sb.String()
}
