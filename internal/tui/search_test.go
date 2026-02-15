package tui

import (
	"testing"
)

func TestSearchBarOpenClose(t *testing.T) {
	sb := NewSearchBar()

	if sb.IsActive() {
		t.Error("should not be active initially")
	}

	sb.Open()
	if !sb.IsActive() {
		t.Error("should be active after Open")
	}

	sb.Close()
	if sb.IsActive() {
		t.Error("should not be active after Close")
	}
}

func TestSearchBarQuery(t *testing.T) {
	sb := NewSearchBar()
	sb.Open()

	if sb.Query() != "" {
		t.Errorf("query should be empty after Open, got %q", sb.Query())
	}
}
