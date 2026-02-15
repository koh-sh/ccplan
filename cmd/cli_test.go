package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

func TestVersionCmdRun(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	v := &VersionCmd{}
	err := v.Run(kong.Vars{"version": "1.2.3"})

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if output != "ccplan version 1.2.3\n" {
		t.Errorf("output = %q, want %q", output, "ccplan version 1.2.3\n")
	}
}

func TestLocateCmdRunNoArgs(t *testing.T) {
	l := &LocateCmd{}
	err := l.Run()
	if err == nil {
		t.Fatal("expected error when no transcript or stdin specified")
	}
	if !strings.Contains(err.Error(), "--transcript or --stdin is required") {
		t.Errorf("error = %q, want to contain '--transcript or --stdin is required'", err.Error())
	}
}

func TestLocateCmdRunWithTranscript(t *testing.T) {
	// Create a valid transcript JSONL with a plan file reference
	tmpDir := t.TempDir()
	plansDir := filepath.Join(tmpDir, ".claude", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create plan file
	planFile := filepath.Join(plansDir, "test-plan.md")
	if err := os.WriteFile(planFile, []byte("# Plan"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create settings
	settingsDir := filepath.Join(tmpDir, ".claude")
	settingsJSON := `{"plansDirectory":"` + plansDir + `"}`
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.local.json"), []byte(settingsJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create transcript in the correct format:
	// {"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Write","input":{"file_path":"..."}}]}}
	transcriptFile := filepath.Join(tmpDir, "transcript.jsonl")
	transcriptLine := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Write","input":{"file_path":"` + planFile + `"}}]}}`
	if err := os.WriteFile(transcriptFile, []byte(transcriptLine+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	l := &LocateCmd{
		Transcript: transcriptFile,
		CWD:        tmpDir,
	}
	err := l.Run()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if output == "" {
		t.Error("expected plan file path in output")
	}
}

func TestLocateCmdRunNoPlanFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty transcript
	transcriptFile := filepath.Join(tmpDir, "transcript.jsonl")
	if err := os.WriteFile(transcriptFile, []byte(`{"type":"other"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	l := &LocateCmd{
		Transcript: transcriptFile,
		CWD:        tmpDir,
	}
	err := l.Run()
	if err == nil {
		t.Fatal("expected error when no plan found")
	}
	if !strings.Contains(err.Error(), "no plan") {
		t.Errorf("error = %q, want to contain 'no plan'", err.Error())
	}
}

func TestReviewCmdRunFileNotFound(t *testing.T) {
	r := &ReviewCmd{
		PlanFile: "/nonexistent/path/plan.md",
	}
	err := r.Run()
	if err == nil {
		t.Fatal("expected error for nonexistent plan file")
	}
	if !strings.Contains(err.Error(), "reading plan file") {
		t.Errorf("error = %q, want to contain 'reading plan file'", err.Error())
	}
}

func TestLocateCmdRunStdinParseError(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	_, _ = w.WriteString("not valid json")
	w.Close()
	os.Stdin = r

	l := &LocateCmd{Stdin: true}
	err := l.Run()
	os.Stdin = oldStdin

	if err == nil {
		t.Fatal("expected error for invalid stdin JSON")
	}
	if !strings.Contains(err.Error(), "parsing") {
		t.Errorf("error = %q, want to contain 'parsing'", err.Error())
	}
}

func TestLocateCmdRunLocateError(t *testing.T) {
	tmpDir := t.TempDir()

	l := &LocateCmd{
		Transcript: "/nonexistent/transcript.jsonl",
		CWD:        tmpDir,
	}
	err := l.Run()
	if err == nil {
		t.Fatal("expected error for nonexistent transcript")
	}
	if !strings.Contains(err.Error(), "locating plan file") {
		t.Errorf("error = %q, want to contain 'locating plan file'", err.Error())
	}
}

func TestReviewCmdRunNoTerminal(t *testing.T) {
	tmpDir := t.TempDir()
	planFile := filepath.Join(tmpDir, "plan.md")
	if err := os.WriteFile(planFile, []byte("# Plan\n\n## Step 1\n\nContent.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := &ReviewCmd{
		PlanFile: planFile,
		Output:   "stdout",
	}
	// Should fail at prog.Run() because no terminal is available
	err := r.Run()
	if err == nil {
		t.Error("expected error when running TUI without terminal")
	}
}

func TestLocateCmdRunStdinMode(t *testing.T) {
	tmpDir := t.TempDir()
	plansDir := filepath.Join(tmpDir, ".claude", "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatal(err)
	}

	planFile := filepath.Join(plansDir, "plan.md")
	if err := os.WriteFile(planFile, []byte("# Plan"), 0o644); err != nil {
		t.Fatal(err)
	}

	settingsDir := filepath.Join(tmpDir, ".claude")
	settingsJSON := `{"plansDirectory":"` + plansDir + `"}`
	if err := os.WriteFile(filepath.Join(settingsDir, "settings.local.json"), []byte(settingsJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	transcriptFile := filepath.Join(tmpDir, "transcript.jsonl")
	transcriptLine := `{"type":"assistant","message":{"role":"assistant","content":[{"type":"tool_use","name":"Write","input":{"file_path":"` + planFile + `"}}]}}`
	if err := os.WriteFile(transcriptFile, []byte(transcriptLine+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create stdin with hook input JSON
	hookInput := `{"session_id":"test","transcript_path":"` + transcriptFile + `","cwd":"` + tmpDir + `"}`
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	_, _ = w.WriteString(hookInput)
	w.Close()
	os.Stdin = r

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	l := &LocateCmd{
		Stdin: true,
	}
	err := l.Run()

	wOut.Close()
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(rOut); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	if output == "" {
		t.Error("expected plan file path in output")
	}
}
