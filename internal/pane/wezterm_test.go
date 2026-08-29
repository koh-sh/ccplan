package pane

import (
	"context"
	"fmt"
	"testing"
)

func TestWezTermSpawnerName(t *testing.T) {
	w := &WezTermSpawner{}
	if w.Name() != "wezterm" {
		t.Errorf("name = %s, want wezterm", w.Name())
	}
}

func TestWezTermSpawnerAvailable(t *testing.T) {
	w := &WezTermSpawner{}
	// Just verify it doesn't panic; result depends on environment
	_ = w.Available()
}

// mockRunner implements cmdRunner for testing.
type mockRunner struct {
	calls   []mockCall
	callIdx int
}

type mockCall struct {
	out []byte
	err error
}

func (m *mockRunner) Output(name string, args ...string) ([]byte, error) {
	if m.callIdx >= len(m.calls) {
		return nil, fmt.Errorf("unexpected call #%d to Output(%s)", m.callIdx, name)
	}
	c := m.calls[m.callIdx]
	m.callIdx++
	return c.out, c.err
}

// newMockSpawner returns a WezTermSpawner whose wezterm invocations return
// the given results in order.
func newMockSpawner(calls ...mockCall) *WezTermSpawner {
	return &WezTermSpawner{runner: &mockRunner{calls: calls}}
}

func TestSplitDirection(t *testing.T) {
	tests := []struct {
		name      string
		paneEnv   string // WEZTERM_PANE
		list      mockCall
		wantDir   string
		wantPct   string
		wantCalls int // wezterm invocations expected
	}{
		{
			name:      "wide pane splits right",
			paneEnv:   "0",
			list:      mockCall{out: []byte(`[{"pane_id":0,"size":{"rows":40,"cols":120,"pixel_width":1920,"pixel_height":1080}}]`)},
			wantDir:   "--right",
			wantPct:   "50",
			wantCalls: 1,
		},
		{
			name:      "tall pane splits bottom",
			paneEnv:   "0",
			list:      mockCall{out: []byte(`[{"pane_id":0,"size":{"rows":80,"cols":60,"pixel_width":800,"pixel_height":1200}}]`)},
			wantDir:   "--bottom",
			wantPct:   "80",
			wantCalls: 1,
		},
		{
			name:      "pixel dimensions unavailable falls back to bottom",
			paneEnv:   "0",
			list:      mockCall{out: []byte(`[{"pane_id":0,"size":{"rows":40,"cols":200,"pixel_width":0,"pixel_height":0}}]`)},
			wantDir:   "--bottom",
			wantPct:   "80",
			wantCalls: 1,
		},
		{
			name:      "WEZTERM_PANE unset falls back to bottom without querying",
			paneEnv:   "",
			wantDir:   "--bottom",
			wantPct:   "80",
			wantCalls: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WEZTERM_PANE", tt.paneEnv)
			runner := &mockRunner{calls: []mockCall{tt.list}}
			w := &WezTermSpawner{runner: runner}
			dir, pct := w.splitDirection()
			if dir != tt.wantDir || pct != tt.wantPct {
				t.Errorf("splitDirection() = %s %s, want %s %s", dir, pct, tt.wantDir, tt.wantPct)
			}
			if runner.callIdx != tt.wantCalls {
				t.Errorf("wezterm calls = %d, want %d", runner.callIdx, tt.wantCalls)
			}
		})
	}
}

func TestCurrentPaneSize(t *testing.T) {
	tests := []struct {
		name     string
		paneEnv  string // WEZTERM_PANE
		list     mockCall
		wantSize *paneSize // nil = expect error
	}{
		{
			name:    "WEZTERM_PANE unset",
			paneEnv: "",
		},
		{
			name:    "command error",
			paneEnv: "0",
			list:    mockCall{err: fmt.Errorf("command not found")},
		},
		{
			name:    "invalid JSON",
			paneEnv: "0",
			list:    mockCall{out: []byte(`not json`)},
		},
		{
			name:    "pane not found",
			paneEnv: "99",
			list:    mockCall{out: []byte(`[{"pane_id":0,"size":{"rows":40,"cols":120}}]`)},
		},
		{
			name:     "success",
			paneEnv:  "5",
			list:     mockCall{out: []byte(`[{"pane_id":5,"size":{"rows":40,"cols":120,"pixel_width":1920,"pixel_height":1080}}]`)},
			wantSize: &paneSize{Rows: 40, Cols: 120, PixelWidth: 1920, PixelHeight: 1080},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WEZTERM_PANE", tt.paneEnv)
			w := newMockSpawner(tt.list)
			size, err := w.currentPaneSize()
			if tt.wantSize == nil {
				if err == nil {
					t.Fatalf("currentPaneSize() = %+v, want error", size)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if *size != *tt.wantSize {
				t.Errorf("size = %+v, want %+v", *size, *tt.wantSize)
			}
		})
	}
}

func TestPaneExists(t *testing.T) {
	tests := []struct {
		name   string
		list   mockCall
		paneID string
		want   bool
	}{
		{
			name:   "pane listed",
			list:   mockCall{out: []byte(`[{"pane_id":42,"size":{"rows":40,"cols":120}}]`)},
			paneID: "42",
			want:   true,
		},
		{
			name:   "pane not listed",
			list:   mockCall{out: []byte(`[{"pane_id":1,"size":{"rows":40,"cols":120}}]`)},
			paneID: "99",
			want:   false,
		},
		{
			name:   "command error reports missing",
			list:   mockCall{err: fmt.Errorf("command failed")},
			paneID: "0",
			want:   false,
		},
		{
			name:   "invalid JSON reports missing",
			list:   mockCall{out: []byte(`invalid`)},
			paneID: "0",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newMockSpawner(tt.list)
			if got := w.paneExists(tt.paneID); got != tt.want {
				t.Errorf("paneExists(%s) = %v, want %v", tt.paneID, got, tt.want)
			}
		})
	}
}

func TestSpawnAndWait(t *testing.T) {
	tests := []struct {
		name    string
		calls   []mockCall
		wantErr bool
	}{
		{
			name: "pane closes after first poll",
			calls: []mockCall{
				// split-pane returns the new pane ID
				{out: []byte("42\n")},
				// first poll: pane gone
				{out: []byte(`[]`)},
			},
		},
		{
			name:    "split-pane failure",
			calls:   []mockCall{{err: fmt.Errorf("split failed")}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WEZTERM_PANE unset so splitDirection falls back without calling the runner.
			t.Setenv("WEZTERM_PANE", "")
			w := newMockSpawner(tt.calls...)
			err := w.SpawnAndWait(context.Background(), "commd", []string{"review", "plan.md"})
			if (err != nil) != tt.wantErr {
				t.Errorf("SpawnAndWait() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
