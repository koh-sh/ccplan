package markdown

import (
	"strings"
	"testing"
)

func TestFormatReview(t *testing.T) {
	tests := []struct {
		name         string
		doc          *Document
		result       *ReviewResult
		filePath     string
		wantEmpty    bool
		wantContains []string
		wantCounts   map[string]int
	}{
		{
			name: "basic two sections",
			doc: &Document{
				Title: "Test Plan",
				Sections: []*Section{
					{ID: "S1", Title: "First Step", Level: 2},
					{ID: "S2", Title: "Second Step", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionSuggestion, Body: "Change the algorithm."},
					{SectionID: "S2", Action: ActionSuggestion, Body: "Not needed."},
				},
			},
			filePath: "/path/to/plan.md",
			wantContains: []string{
				"## S1: First Step\n",
				"[suggestion] Change the algorithm.",
				"## S2: Second Step\n",
				"/path/to/plan.md",
			},
		},
		{
			name:      "empty comments returns empty string",
			doc:       &Document{Title: "Test"},
			result:    &ReviewResult{},
			wantEmpty: true,
		},
		{
			name: "single section with body",
			doc: &Document{
				Sections: []*Section{
					{ID: "S1", Title: "Step One", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionSuggestion, Body: "Looks good but needs refactoring."},
				},
			},
			filePath: "test.md",
			wantContains: []string{
				"[suggestion] Looks good but needs refactoring.",
			},
		},
		{
			name: "grouped comments under same section",
			doc: &Document{
				Sections: []*Section{
					{ID: "S1", Title: "JWT verification", Level: 2},
					{ID: "S3", Title: "Add tests", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionSuggestion, Body: "Switch to HS256."},
					{SectionID: "S1", Action: ActionIssue, Body: "Not needed."},
					{SectionID: "S3", Action: ActionQuestion, Body: "Coverage target?"},
				},
			},
			filePath: "/path/to/plan.md",
			wantContains: []string{
				"[suggestion] Switch to HS256.",
				"[issue] Not needed.",
				"## S3: Add tests\n",
				"[question] Coverage target?",
			},
			wantCounts: map[string]int{
				"## S1: JWT verification": 1,
			},
		},
		{
			name: "comment with decoration",
			doc: &Document{
				Sections: []*Section{
					{ID: "S1", Title: "Step One", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionSuggestion, Decoration: DecorationNonBlocking, Body: "Use a cache."},
				},
			},
			filePath: "plan.md",
			wantContains: []string{
				"[suggestion (non-blocking)] Use a cache.",
			},
		},
		{
			name: "comment without decoration (zero value)",
			doc: &Document{
				Sections: []*Section{
					{ID: "S1", Title: "Step One", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionSuggestion, Body: "plain comment"},
				},
			},
			filePath: "plan.md",
			wantContains: []string{
				"[suggestion] plain comment",
			},
		},
		{
			name: "overview comment uses Overview heading",
			doc: &Document{
				Title: "Test Plan",
				Sections: []*Section{
					{ID: "S1", Title: "First Step", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: OverviewSectionID, Action: ActionNote, Body: "Overall looks good."},
					{SectionID: "S1", Action: ActionSuggestion, Body: "Change the algorithm."},
				},
			},
			filePath: "plan.md",
			wantContains: []string{
				"## Overview\n",
				"[note] Overall looks good.",
				"## S1: First Step\n",
				"[suggestion] Change the algorithm.",
			},
		},
		{
			name: "line-level comment single line",
			doc: &Document{
				Sections: []*Section{
					{ID: "S1", Title: "Step One", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionSuggestion, Body: "Fix this line.", StartLine: 10},
				},
			},
			filePath: "plan.md",
			wantContains: []string{
				"---\n",
				"`L10` [suggestion] Fix this line.",
			},
		},
		{
			name: "line-level comment range",
			doc: &Document{
				Sections: []*Section{
					{ID: "S1", Title: "Step One", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionIssue, Decoration: DecorationBlocking, Body: "Refactor this block.", StartLine: 10, EndLine: 15},
				},
			},
			filePath: "plan.md",
			wantContains: []string{
				"`L10-L15` [issue (blocking)] Refactor this block.",
			},
		},
		{
			name: "mixed section and line comments",
			doc: &Document{
				Sections: []*Section{
					{ID: "S1", Title: "Step One", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionNote, Body: "Section comment."},
					{SectionID: "S1", Action: ActionSuggestion, Body: "Line comment.", StartLine: 5},
				},
			},
			filePath: "plan.md",
			wantContains: []string{
				"## S1: Step One\n",
				"[note] Section comment.",
				"---\n",
				"`L5` [suggestion] Line comment.",
			},
		},
		{
			name: "empty filePath uses fallback text",
			doc: &Document{
				Sections: []*Section{
					{ID: "S1", Title: "Step One", Level: 2},
				},
			},
			result: &ReviewResult{
				Comments: []ReviewComment{
					{SectionID: "S1", Action: ActionSuggestion, Body: "comment"},
				},
			},
			filePath: "",
			wantContains: []string{
				"the file",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := FormatReview(tt.result, tt.doc, tt.filePath)

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("expected empty output, got: %q", output)
				}
				return
			}

			for _, s := range tt.wantContains {
				if !strings.Contains(output, s) {
					t.Errorf("output missing %q, got:\n%s", s, output)
				}
			}

			for s, wantN := range tt.wantCounts {
				if gotN := strings.Count(output, s); gotN != wantN {
					t.Errorf("expected %q to appear %d time(s), got %d in:\n%s", s, wantN, gotN, output)
				}
			}
		})
	}
}
