package cclocate

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fixturePlansDir is the plansDirectory the testdata transcripts refer to.
const fixturePlansDir = "/tmp/test-plans"

// writeTranscript writes content as a transcript file in a temp dir and
// returns its path.
func writeTranscript(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFindPlanFilesInTranscript(t *testing.T) {
	tmpPlans := filepath.Join(t.TempDir(), "plans")
	fixture := func(name string) string { return filepath.Join("testdata", name) }
	plan := func(name string) string { return filepath.Join(fixturePlansDir, name) }

	tests := []struct {
		name       string
		transcript string
		plansDir   string
		all        bool
		want       []string
		wantErr    string
	}{
		{
			name:       "latest plan",
			transcript: fixture("transcript-with-plan.jsonl"),
			plansDir:   fixturePlansDir,
			want:       []string{plan("jaunty-petting-nebula.md")},
		},
		{
			name:       "no plan written",
			transcript: fixture("transcript-no-plan.jsonl"),
			plansDir:   fixturePlansDir,
		},
		{
			name:       "multiple plans returns latest only",
			transcript: fixture("transcript-multiple-plans.jsonl"),
			plansDir:   fixturePlansDir,
			want:       []string{plan("plan-b.md")},
		},
		{
			name:       "multiple plans with all returns newest first",
			transcript: fixture("transcript-multiple-plans.jsonl"),
			plansDir:   fixturePlansDir,
			all:        true,
			want:       []string{plan("plan-b.md"), plan("plan-a.md")},
		},
		{
			name:       "malformed lines are skipped",
			transcript: fixture("transcript-malformed.jsonl"),
			plansDir:   fixturePlansDir,
			all:        true,
			want:       []string{plan("valid.md")},
		},
		{
			name:       "plans under another directory are ignored",
			transcript: fixture("transcript-with-plan.jsonl"),
			plansDir:   "/some/other/dir",
		},
		{
			name:       "nonexistent transcript",
			transcript: filepath.Join(t.TempDir(), "missing.jsonl"),
			plansDir:   fixturePlansDir,
			wantErr:    "no such file",
		},
		{
			name: "empty lines are skipped",
			transcript: writeTranscript(t, "\n"+
				`{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Write","input":{"file_path":"`+filepath.Join(tmpPlans, "plan.md")+`"}}]}}`+"\n"+
				"\n\n"),
			plansDir: tmpPlans,
			all:      true,
			want:     []string{filepath.Join(tmpPlans, "plan.md")},
		},
		{
			name:       "invalid content block is skipped",
			transcript: writeTranscript(t, `{"type":"assistant","message":{"role":"assistant","content":["not a valid block"]}}`+"\n"),
			plansDir:   tmpPlans,
			all:        true,
		},
		{
			name:       "invalid tool input is skipped",
			transcript: writeTranscript(t, `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Write","input":"not-an-object"}]}}`+"\n"),
			plansDir:   tmpPlans,
			all:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findPlanFilesInTranscript(tt.transcript, tt.plansDir, tt.all)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("findPlanFilesInTranscript() error: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("paths = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLocatePlanFile(t *testing.T) {
	// Create a temp plan file so validation passes
	tmpDir := t.TempDir()
	plansDir := filepath.Join(tmpDir, ".claude", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatal(err)
	}
	planFile := filepath.Join(plansDir, "test-plan.md")
	if err := os.WriteFile(planFile, []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a settings file pointing to plansDir
	settingsDir := filepath.Join(tmpDir, ".claude")
	settingsFile := filepath.Join(settingsDir, "settings.json")
	settingsContent := `{"plansDirectory": "` + plansDir + `"}`
	if err := os.WriteFile(settingsFile, []byte(settingsContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a transcript that references the plan file
	transcriptFile := filepath.Join(tmpDir, "transcript.jsonl")
	transcriptContent := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","id":"toolu_001","name":"Write","input":{"file_path":"` + planFile + `","content":"# Test"}}]},"sessionId":"test","timestamp":"2025-01-01T00:00:00Z"}` + "\n"
	if err := os.WriteFile(transcriptFile, []byte(transcriptContent), 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := LocatePlanFile(Options{
		TranscriptPath: transcriptFile,
		CWD:            tmpDir,
	})
	if err != nil {
		t.Fatalf("LocatePlanFile() error: %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("len(paths) = %d, want 1", len(paths))
	}
	if paths[0] != filepath.Clean(planFile) {
		t.Errorf("paths[0] = %q, want %q", paths[0], filepath.Clean(planFile))
	}
}
