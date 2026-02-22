package tui

import (
	"testing"
)

func TestSearchBar(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*SearchBar)
		wantActive bool
		wantQuery  string
	}{
		{
			name:       "not active initially",
			setup:      func(*SearchBar) {},
			wantActive: false,
		},
		{
			name:       "active after Open",
			setup:      func(sb *SearchBar) { sb.Open() },
			wantActive: true,
			wantQuery:  "",
		},
		{
			name:       "not active after Open then Close",
			setup:      func(sb *SearchBar) { sb.Open(); sb.Close() },
			wantActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewSearchBar()
			tt.setup(sb)

			if got := sb.IsActive(); got != tt.wantActive {
				t.Errorf("IsActive() = %v, want %v", got, tt.wantActive)
			}
			if tt.wantActive {
				if got := sb.Query(); got != tt.wantQuery {
					t.Errorf("Query() = %q, want %q", got, tt.wantQuery)
				}
			}
		})
	}
}
