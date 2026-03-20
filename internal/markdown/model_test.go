package markdown

import "testing"

func TestReviewCommentFormatLabel(t *testing.T) {
	tests := []struct {
		name    string
		comment ReviewComment
		want    string
	}{
		{
			name:    "no decoration",
			comment: ReviewComment{Action: ActionSuggestion},
			want:    "suggestion",
		},
		{
			name:    "zero value decoration",
			comment: ReviewComment{Action: ActionIssue, Decoration: DecorationNone},
			want:    "issue",
		},
		{
			name:    "non-blocking",
			comment: ReviewComment{Action: ActionSuggestion, Decoration: DecorationNonBlocking},
			want:    "suggestion (non-blocking)",
		},
		{
			name:    "blocking",
			comment: ReviewComment{Action: ActionIssue, Decoration: DecorationBlocking},
			want:    "issue (blocking)",
		},
		{
			name:    "if-minor",
			comment: ReviewComment{Action: ActionNitpick, Decoration: DecorationIfMinor},
			want:    "nitpick (if-minor)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.comment.FormatLabel()
			if got != tt.want {
				t.Errorf("FormatLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReviewCommentFormatLineRef(t *testing.T) {
	tests := []struct {
		name    string
		comment ReviewComment
		want    string
	}{
		{
			name:    "section-level (no line info)",
			comment: ReviewComment{StartLine: 0, EndLine: 0},
			want:    "",
		},
		{
			name:    "single line",
			comment: ReviewComment{StartLine: 10, EndLine: 0},
			want:    "L10",
		},
		{
			name:    "single line (EndLine equals StartLine)",
			comment: ReviewComment{StartLine: 5, EndLine: 5},
			want:    "L5",
		},
		{
			name:    "line range",
			comment: ReviewComment{StartLine: 10, EndLine: 15},
			want:    "L10-L15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.comment.FormatLineRef()
			if got != tt.want {
				t.Errorf("FormatLineRef() = %q, want %q", got, tt.want)
			}
		})
	}
}
