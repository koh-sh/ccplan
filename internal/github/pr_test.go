package github

import (
	"testing"
)

func TestParsePRURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *PRRef
		wantErr bool
	}{
		{
			name: "valid URL",
			url:  "https://github.com/owner/repo/pull/123",
			want: &PRRef{Owner: "owner", Repo: "repo", Number: 123},
		},
		{
			name: "valid URL with trailing slash",
			url:  "https://github.com/owner/repo/pull/456/",
			want: &PRRef{Owner: "owner", Repo: "repo", Number: 456},
		},
		{
			name: "valid URL with org containing hyphens",
			url:  "https://github.com/my-org/my-repo/pull/1",
			want: &PRRef{Owner: "my-org", Repo: "my-repo", Number: 1},
		},
		{
			name:    "http scheme rejected",
			url:     "http://github.com/owner/repo/pull/123",
			wantErr: true,
		},
		{
			name:    "non-github host",
			url:     "https://gitlab.com/owner/repo/pull/123",
			wantErr: true,
		},
		{
			name:    "missing pull segment",
			url:     "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
		{
			name:    "missing PR number",
			url:     "https://github.com/owner/repo/pull",
			wantErr: true,
		},
		{
			name:    "non-numeric PR number",
			url:     "https://github.com/owner/repo/pull/abc",
			wantErr: true,
		},
		{
			name:    "too many path segments",
			url:     "https://github.com/owner/repo/pull/123/files",
			wantErr: true,
		},
		{
			name:    "too few path segments",
			url:     "https://github.com/owner/pull/123",
			wantErr: true,
		},
		{
			name:    "empty string",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePRURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParsePRURL(%q) = %+v, want error", tt.url, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePRURL(%q) error: %v", tt.url, err)
			}
			if got.Owner != tt.want.Owner || got.Repo != tt.want.Repo || got.Number != tt.want.Number {
				t.Errorf("ParsePRURL(%q) = %+v, want %+v", tt.url, got, tt.want)
			}
		})
	}
}
