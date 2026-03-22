package github

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// PRRef identifies a GitHub pull request.
type PRRef struct {
	Owner  string
	Repo   string
	Number int
}

// PRFile represents a Markdown file changed in a PR.
type PRFile struct {
	Path  string // file path in the repo
	Patch string // unified diff patch
}

// ParsePRURL parses a GitHub PR URL into its components.
// Accepts: https://github.com/{owner}/{repo}/pull/{number}
func ParsePRURL(rawURL string) (*PRRef, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid PR URL: %w", err)
	}

	if u.Scheme != "https" {
		return nil, fmt.Errorf("invalid PR URL: expected https scheme, got %q", u.Scheme)
	}
	if u.Host != "github.com" {
		return nil, fmt.Errorf("invalid PR URL: expected github.com host, got %q", u.Host)
	}

	// Path: /{owner}/{repo}/pull/{number}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 || parts[2] != "pull" {
		return nil, fmt.Errorf("invalid PR URL: expected https://github.com/{owner}/{repo}/pull/{number}")
	}

	number, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid PR URL: PR number %q is not a valid integer", parts[3])
	}

	return &PRRef{
		Owner:  parts[0],
		Repo:   parts[1],
		Number: number,
	}, nil
}
