package github

import (
	"fmt"
	"strconv"
	"strings"
)

// DiffLineType represents the type of a diff line.
type DiffLineType byte

const (
	DiffContext DiffLineType = ' '
	DiffAdded   DiffLineType = '+'
	DiffRemoved DiffLineType = '-'
)

// Side constants for GitHub PR review comments.
const (
	SideRight = "RIGHT"
	SideLeft  = "LEFT"
)

// DiffLine represents a single line in a unified diff.
type DiffLine struct {
	Type    DiffLineType
	Content string // line content without the +/-/space prefix
	NewLine int    // 1-based line number in the new file (0 for removed lines)
	OldLine int    // 1-based line number in the old file (0 for added lines)
}

// DiffInfo contains parsed diff data for a PR file.
type DiffInfo struct {
	Lines []DiffLine
}

// ParsePatch parses a unified diff patch string into DiffInfo.
func ParsePatch(patch string) *DiffInfo {
	if patch == "" {
		return nil
	}

	info := &DiffInfo{}

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
			info.Lines = append(info.Lines, DiffLine{
				Type:    DiffAdded,
				Content: content,
				NewLine: newLine,
			})
			newLine++
		case '-':
			info.Lines = append(info.Lines, DiffLine{
				Type:    DiffRemoved,
				Content: content,
				OldLine: oldLine,
			})
			oldLine++
		case ' ':
			info.Lines = append(info.Lines, DiffLine{
				Type:    DiffContext,
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
func (d *DiffInfo) FormatDiffLines() []string {
	result := make([]string, len(d.Lines))
	for i, dl := range d.Lines {
		result[i] = fmt.Sprintf("%c %s", dl.Type, dl.Content)
	}
	return result
}

// LineSideMap builds line number, side, and type maps from the parsed diff.
// lineMap maps display index → file line number, sideMap → "RIGHT"/"LEFT",
// typeMap → DiffLineType byte ('+', '-', ' ').
func (d *DiffInfo) LineSideMap() (lineMap []int, sideMap []string, typeMap []byte) {
	lineMap = make([]int, len(d.Lines))
	sideMap = make([]string, len(d.Lines))
	typeMap = make([]byte, len(d.Lines))
	for i, dl := range d.Lines {
		typeMap[i] = byte(dl.Type)
		if dl.Type == DiffRemoved {
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
