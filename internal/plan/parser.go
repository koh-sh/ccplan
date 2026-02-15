package plan

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Parse parses a Markdown source into a Plan structure.
// It uses goldmark to build an AST and walks headings to create steps.
func Parse(source []byte) (*Plan, error) {
	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	plan := &Plan{}

	type headingInfo struct {
		level int
		title string
		start int // byte offset of the heading line start (including # markers)
		end   int // byte offset past the heading line end (including newline)
	}

	var headings []headingInfo

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		title := extractHeadingText(heading, source)
		start := findHeadingStart(heading, source)
		end := findHeadingEnd(heading, source)

		headings = append(headings, headingInfo{
			level: heading.Level,
			title: title,
			start: start,
			end:   end,
		})

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking AST: %w", err)
	}

	if len(headings) == 0 {
		plan.Preamble = strings.TrimSpace(string(source))
		return plan, nil
	}

	// Extract preamble (text before first heading)
	if headings[0].start > 0 {
		plan.Preamble = strings.TrimSpace(string(source[:headings[0].start]))
	}

	var flatSteps []flatStep

	for i, h := range headings {
		bodyStart := h.end
		var bodyEnd int
		if i+1 < len(headings) {
			bodyEnd = headings[i+1].start
		} else {
			bodyEnd = len(source)
		}
		body := ""
		if bodyStart < bodyEnd {
			body = strings.TrimSpace(string(source[bodyStart:bodyEnd]))
		}

		if h.level == 1 && plan.Title == "" {
			plan.Title = h.title
			if body != "" {
				if plan.Preamble != "" {
					plan.Preamble = plan.Preamble + "\n\n" + body
				} else {
					plan.Preamble = body
				}
			}
			continue
		}

		flatSteps = append(flatSteps, flatStep{
			level: h.level,
			title: h.title,
			body:  body,
		})
	}

	plan.Steps = buildHierarchy(flatSteps)

	return plan, nil
}

// findHeadingStart finds the byte offset where the heading line starts in source.
// goldmark Lines() positions may point past the # markers for ATX headings,
// so we always search backwards to find the true line start.
func findHeadingStart(heading *ast.Heading, source []byte) int {
	pos := -1

	// Try Lines() first
	if heading.Lines().Len() > 0 {
		pos = heading.Lines().At(0).Start
	}

	// Fall back to child text nodes
	if pos < 0 {
		pos = findFirstTextPos(heading)
	}

	if pos < 0 {
		return 0
	}

	// Search backwards to find the actual line start (before # markers)
	for pos > 0 && source[pos-1] != '\n' {
		pos--
	}
	return pos
}

// findHeadingEnd finds the byte offset past the heading line end (including newline).
func findHeadingEnd(heading *ast.Heading, source []byte) int {
	pos := -1

	// Try Lines() first
	if heading.Lines().Len() > 0 {
		last := heading.Lines().At(heading.Lines().Len() - 1)
		pos = last.Stop
	}

	// Fall back to child text nodes
	if pos < 0 {
		pos = findLastTextPos(heading)
	}

	if pos < 0 {
		return 0
	}

	// Advance to end of line and include newline
	for pos < len(source) && source[pos] != '\n' {
		pos++
	}
	if pos < len(source) {
		pos++ // include the newline
	}
	return pos
}

// findFirstTextPos returns the Start position of the first Text segment in the node.
func findFirstTextPos(n ast.Node) int {
	pos := -1
	_ = ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := child.(*ast.Text); ok {
			pos = t.Segment.Start
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return pos
}

// findLastTextPos returns the Stop position of the last Text segment in the node.
func findLastTextPos(n ast.Node) int {
	pos := -1
	_ = ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := child.(*ast.Text); ok {
			pos = t.Segment.Stop
		}
		return ast.WalkContinue, nil
	})
	return pos
}

// buildHierarchy converts a flat list of steps into a tree based on heading levels.
func buildHierarchy(flatSteps []flatStep) []*Step {
	if len(flatSteps) == 0 {
		return nil
	}

	var topLevel []*Step
	var parentStack []*Step

	for _, fs := range flatSteps {
		step := &Step{
			Level: fs.level,
			Title: fs.title,
			Body:  fs.body,
		}

		// Pop from stack until we find a parent with a lower level
		for len(parentStack) > 0 && parentStack[len(parentStack)-1].Level >= fs.level {
			parentStack = parentStack[:len(parentStack)-1]
		}

		if len(parentStack) > 0 {
			parent := parentStack[len(parentStack)-1]
			step.Parent = parent
			parent.Children = append(parent.Children, step)
		} else {
			topLevel = append(topLevel, step)
		}

		parentStack = append(parentStack, step)
	}

	assignIDs(topLevel, "")
	return topLevel
}

// assignIDs assigns hierarchical IDs to steps (S1, S1.1, S1.2, S2, ...).
func assignIDs(steps []*Step, prefix string) {
	for i, s := range steps {
		if prefix == "" {
			s.ID = fmt.Sprintf("S%d", i+1)
		} else {
			s.ID = fmt.Sprintf("%s.%d", prefix, i+1)
		}
		assignIDs(s.Children, s.ID)
	}
}

// extractHeadingText extracts the plain text content of a heading node.
func extractHeadingText(heading *ast.Heading, source []byte) string {
	var sb strings.Builder
	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		extractNodeText(child, source, &sb)
	}
	return sb.String()
}

// extractNodeText recursively extracts text from an AST node.
func extractNodeText(n ast.Node, source []byte, sb *strings.Builder) {
	if t, ok := n.(*ast.Text); ok {
		sb.Write(t.Segment.Value(source))
		if t.HardLineBreak() || t.SoftLineBreak() {
			sb.WriteByte(' ')
		}
	}
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		extractNodeText(child, source, sb)
	}
}

type flatStep struct {
	level int
	title string
	body  string
}
