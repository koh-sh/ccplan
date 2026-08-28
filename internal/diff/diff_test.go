package diff

import (
	"strings"
	"testing"
)

func TestParsePatch(t *testing.T) {
	tests := []struct {
		name      string
		patch     string
		wantLines int
		wantTypes []LineType
	}{
		{
			name: "simple addition",
			patch: `@@ -1,3 +1,4 @@
 # Title
+New line
 ## Section
 Content`,
			wantLines: 4,
			wantTypes: []LineType{Context, Added, Context, Context},
		},
		{
			name: "mixed add and remove",
			patch: `@@ -1,3 +1,3 @@
 # Title
-Old line
+New line
 Content`,
			wantLines: 4,
			wantTypes: []LineType{Context, Removed, Added, Context},
		},
		{
			name: "multiple hunks",
			patch: `@@ -1,2 +1,2 @@
-Old title
+New title
 Content
@@ -10,2 +10,3 @@
 Section
+Added
 End`,
			wantLines: 6,
			wantTypes: []LineType{Removed, Added, Context, Context, Added, Context},
		},
		{
			name:      "no newline marker is skipped",
			patch:     "@@ -1,2 +1,2 @@\n-old\n\\ No newline at end of file\n+new\n Content",
			wantLines: 3,
			wantTypes: []LineType{Removed, Added, Context},
		},
		{
			name:      "empty patch",
			patch:     "",
			wantLines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := ParsePatch(tt.patch)

			if tt.patch == "" {
				if info != nil {
					t.Fatal("expected nil for empty patch")
				}
				return
			}

			if len(info.Lines) != tt.wantLines {
				t.Fatalf("got %d lines, want %d", len(info.Lines), tt.wantLines)
			}

			for i, wantType := range tt.wantTypes {
				if info.Lines[i].Type != wantType {
					t.Errorf("line[%d].Type = %c, want %c", i, info.Lines[i].Type, wantType)
				}
			}
		})
	}
}

func TestParsePatchLineNumbers(t *testing.T) {
	info := ParsePatch(`@@ -5,3 +10,3 @@
 context
-removed
+added`)

	if len(info.Lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(info.Lines))
	}

	// Context line: both old and new
	if info.Lines[0].NewLine != 10 || info.Lines[0].OldLine != 5 {
		t.Errorf("context: NewLine=%d OldLine=%d, want 10,5", info.Lines[0].NewLine, info.Lines[0].OldLine)
	}
	// Removed line: only old
	if info.Lines[1].NewLine != 0 || info.Lines[1].OldLine != 6 {
		t.Errorf("removed: NewLine=%d OldLine=%d, want 0,6", info.Lines[1].NewLine, info.Lines[1].OldLine)
	}
	// Added line: only new
	if info.Lines[2].NewLine != 11 || info.Lines[2].OldLine != 0 {
		t.Errorf("added: NewLine=%d OldLine=%d, want 11,0", info.Lines[2].NewLine, info.Lines[2].OldLine)
	}
}

func TestFormatDiffLines(t *testing.T) {
	info := ParsePatch(`@@ -1,2 +1,2 @@
-old
+new
 same`)

	lines := info.FormatDiffLines()
	want := []string{"- old", "+ new", "  same"}

	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d", len(lines), len(want))
	}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], w)
		}
	}
}

func TestParseHunkHeader(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantNew int
		wantOld int
	}{
		{"standard", "@@ -1,3 +1,4 @@", 1, 1},
		{"offset", "@@ -10,5 +15,8 @@", 15, 10},
		{"no count", "@@ -1 +1 @@", 1, 1},
		{"with context", "@@ -1,3 +1,4 @@ func main()", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNew, gotOld := parseHunkHeader(tt.line)
			if gotNew != tt.wantNew {
				t.Errorf("newStart = %d, want %d", gotNew, tt.wantNew)
			}
			if gotOld != tt.wantOld {
				t.Errorf("oldStart = %d, want %d", gotOld, tt.wantOld)
			}
		})
	}
}

func TestLineSideMap(t *testing.T) {
	patch := "@@ -1,3 +1,3 @@\n context\n-removed\n+added\n same"
	info := ParsePatch(patch)
	if info == nil {
		t.Fatal("ParsePatch returned nil")
	}

	lineMap, sideMap, typeMap := info.LineSideMap()

	// context: old=1/new=1, removed: old=2, added: new=2, context: old=3/new=3
	wantLineMap := []int{1, 2, 2, 3}
	wantSideMap := []string{SideRight, SideLeft, SideRight, SideRight}
	wantTypeMap := []byte{' ', '-', '+', ' '}

	for i := range lineMap {
		if lineMap[i] != wantLineMap[i] {
			t.Errorf("lineMap[%d] = %d, want %d", i, lineMap[i], wantLineMap[i])
		}
		if sideMap[i] != wantSideMap[i] {
			t.Errorf("sideMap[%d] = %q, want %q", i, sideMap[i], wantSideMap[i])
		}
		if typeMap[i] != wantTypeMap[i] {
			t.Errorf("typeMap[%d] = %c, want %c", i, typeMap[i], wantTypeMap[i])
		}
	}
}

func TestStripHeader(t *testing.T) {
	gitPatch := strings.Join([]string{
		"diff --git a/doc.md b/doc.md",
		"index 1234567..89abcde 100644",
		"--- a/doc.md",
		"+++ b/doc.md",
		"@@ -1,2 +1,3 @@",
		" # Title",
		"+added",
		" body",
	}, "\n")

	tests := []struct {
		name  string
		patch string
		want  string
	}{
		{
			name:  "git diff headers are removed",
			patch: gitPatch,
			want:  "@@ -1,2 +1,3 @@\n # Title\n+added\n body",
		},
		{
			name:  "patch already starting at hunk passes through",
			patch: "@@ -1 +1 @@\n-a\n+b",
			want:  "@@ -1 +1 @@\n-a\n+b",
		},
		{
			name:  "no hunk header yields empty",
			patch: "diff --git a/x b/x\nindex 0..1",
			want:  "",
		},
		{
			name:  "@@ inside header line is not a hunk start",
			patch: "diff --git a/@@x b/@@x\n@@ -1 +1 @@\n-a\n+b",
			want:  "@@ -1 +1 @@\n-a\n+b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripHeader(tt.patch)
			if got != tt.want {
				t.Errorf("StripHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddedFilePatch(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantPatch string
		wantLines int
	}{
		{
			name:      "two lines with trailing newline",
			content:   "# Title\nbody\n",
			wantPatch: "@@ -0,0 +1,2 @@\n+# Title\n+body\n",
			wantLines: 2,
		},
		{
			name:      "no trailing newline",
			content:   "only",
			wantPatch: "@@ -0,0 +1,1 @@\n+only\n",
			wantLines: 1,
		},
		{
			name:      "empty content",
			content:   "",
			wantPatch: "",
			wantLines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddedFilePatch([]byte(tt.content))
			if got != tt.wantPatch {
				t.Fatalf("AddedFilePatch() = %q, want %q", got, tt.wantPatch)
			}
			info := ParsePatch(got)
			if tt.wantLines == 0 {
				if info != nil {
					t.Fatal("expected nil Info for empty patch")
				}
				return
			}
			if len(info.Lines) != tt.wantLines {
				t.Fatalf("got %d lines, want %d", len(info.Lines), tt.wantLines)
			}
			for i, l := range info.Lines {
				if l.Type != Added || l.NewLine != i+1 {
					t.Errorf("line[%d] = %+v, want Added at NewLine %d", i, l, i+1)
				}
			}
		})
	}
}

func TestParsePatchCRLF(t *testing.T) {
	info := ParsePatch("@@ -1,2 +1,2 @@\r\n ctx\r\n-old\r\n+new\r\n")
	if info == nil {
		t.Fatal("ParsePatch returned nil")
	}
	want := []string{"ctx", "old", "new"}
	if len(info.Lines) != len(want) {
		t.Fatalf("got %d lines, want %d", len(info.Lines), len(want))
	}
	for i, w := range want {
		if info.Lines[i].Content != w {
			t.Errorf("line[%d].Content = %q, want %q (carriage return must be stripped)", i, info.Lines[i].Content, w)
		}
	}
}
