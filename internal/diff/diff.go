// Package diff parses unified diff patches into line-level data used by the
// diff review view. It is independent of where the patch came from (GitHub
// API or a local git working tree).
package diff

import (
	"fmt"
	"strconv"
	"strings"
)

// LineType represents the type of a diff line.
type LineType byte

const (
	Context LineType = ' '
	Added   LineType = '+'
	Removed LineType = '-'
)

// Side constants for GitHub PR review comments.
const (
	SideRight = "RIGHT"
	SideLeft  = "LEFT"
)

// Line represents a single line in a unified diff.
type Line struct {
	Type    LineType
	Content string // line content without the +/-/space prefix
	NewLine int    // 1-based line number in the new file (0 for removed lines)
	OldLine int    // 1-based line number in the old file (0 for added lines)
}

// Info contains parsed diff data for a PR file.
type Info struct {
	Lines []Line
}

// ParsePatch parses a unified diff patch string into Info.
func ParsePatch(patch string) *Info {
	if patch == "" {
		return nil
	}

	info := &Info{}

	lines := strings.Split(patch, "\n")
	var newLine, oldLine int

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			newLine, oldLine = parseHunkHeader(line)
			continue
		}

		if len(line) == 0 {
			continue
		}

		prefix := line[0]
		content := ""
		if len(line) > 1 {
			content = line[1:]
		}

		switch prefix {
		case '+':
			info.Lines = append(info.Lines, Line{
				Type:    Added,
				Content: content,
				NewLine: newLine,
			})
			newLine++
		case '-':
			info.Lines = append(info.Lines, Line{
				Type:    Removed,
				Content: content,
				OldLine: oldLine,
			})
			oldLine++
		case ' ':
			info.Lines = append(info.Lines, Line{
				Type:    Context,
				Content: content,
				NewLine: newLine,
				OldLine: oldLine,
			})
			newLine++
			oldLine++
		case '\\':
			// "\ No newline at end of file" — skip
		}
	}

	return info
}

// FormatDiffLines returns display strings for the diff lines with +/-/space prefix.
// A space separator is inserted between the prefix and content for readability.
func (d *Info) FormatDiffLines() []string {
	result := make([]string, len(d.Lines))
	for i, dl := range d.Lines {
		result[i] = fmt.Sprintf("%c %s", dl.Type, dl.Content)
	}
	return result
}

// LineSideMap builds line number, side, and type maps from the parsed diff.
// lineMap maps display index → file line number, sideMap → "RIGHT"/"LEFT",
// typeMap → LineType byte ('+', '-', ' ').
func (d *Info) LineSideMap() (lineMap []int, sideMap []string, typeMap []byte) {
	lineMap = make([]int, len(d.Lines))
	sideMap = make([]string, len(d.Lines))
	typeMap = make([]byte, len(d.Lines))
	for i, dl := range d.Lines {
		typeMap[i] = byte(dl.Type)
		if dl.Type == Removed {
			lineMap[i] = dl.OldLine
			sideMap[i] = SideLeft
		} else {
			lineMap[i] = dl.NewLine
			sideMap[i] = SideRight
		}
	}
	return
}

// parseHunkHeader parses "@@ -old,count +new,count @@" and returns (newStart, oldStart).
func parseHunkHeader(line string) (int, int) {
	newStart := 1
	oldStart := 1

	// Parse -X part
	if _, after, ok := strings.Cut(line, "-"); ok {
		oldStart = parseNumberAt(after)
	}
	// Parse +N part
	if _, after, ok := strings.Cut(line, "+"); ok {
		newStart = parseNumberAt(after)
	}

	return newStart, oldStart
}

func parseNumberAt(s string) int {
	end := len(s)
	for i, c := range s {
		if c < '0' || c > '9' {
			end = i
			break
		}
	}
	n, err := strconv.Atoi(s[:end])
	if err != nil {
		return 1
	}
	return n
}
