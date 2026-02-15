package locate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePlansDirDefault(t *testing.T) {
	// Use a temp dir with no settings files
	tmpDir := t.TempDir()
	result := ResolvePlansDir(tmpDir)

	want := defaultPlansDir()
	if result != want {
		t.Errorf("ResolvePlansDir() = %q, want default %q", result, want)
	}
}

func TestResolvePlansDirFromLocalSettings(t *testing.T) {
	tmpDir := t.TempDir()
	settingsDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write settings.local.json
	localSettings := filepath.Join(settingsDir, "settings.local.json")
	if err := os.WriteFile(localSettings, []byte(`{"plansDirectory": "/custom/plans"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	result := ResolvePlansDir(tmpDir)
	if result != "/custom/plans" {
		t.Errorf("ResolvePlansDir() = %q, want %q", result, "/custom/plans")
	}
}

func TestResolvePlansDirFromSettings(t *testing.T) {
	tmpDir := t.TempDir()
	settingsDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settings := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settings, []byte(`{"plansDirectory": "/project/plans"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	result := ResolvePlansDir(tmpDir)
	if result != "/project/plans" {
		t.Errorf("ResolvePlansDir() = %q, want %q", result, "/project/plans")
	}
}

func TestResolvePlansDirRelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	settingsDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	settings := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settings, []byte(`{"plansDirectory": "plans"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	result := ResolvePlansDir(tmpDir)
	want := filepath.Join(tmpDir, "plans")
	if result != want {
		t.Errorf("ResolvePlansDir() = %q, want %q", result, want)
	}
}

func TestResolvePlansDirPriority(t *testing.T) {
	tmpDir := t.TempDir()
	settingsDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Both local and regular settings exist; local should win
	localSettings := filepath.Join(settingsDir, "settings.local.json")
	if err := os.WriteFile(localSettings, []byte(`{"plansDirectory": "/local/plans"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	settings := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settings, []byte(`{"plansDirectory": "/regular/plans"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	result := ResolvePlansDir(tmpDir)
	if result != "/local/plans" {
		t.Errorf("ResolvePlansDir() = %q, want %q (local should take priority)", result, "/local/plans")
	}
}

func TestResolvePlansDirBrokenJSON(t *testing.T) {
	tmpDir := t.TempDir()
	settingsDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(settingsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Broken settings.local.json should be skipped
	localSettings := filepath.Join(settingsDir, "settings.local.json")
	if err := os.WriteFile(localSettings, []byte(`{broken`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Valid settings.json should be used as fallback
	settings := filepath.Join(settingsDir, "settings.json")
	if err := os.WriteFile(settings, []byte(`{"plansDirectory": "/fallback/plans"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	result := ResolvePlansDir(tmpDir)
	if result != "/fallback/plans" {
		t.Errorf("ResolvePlansDir() = %q, want %q (should fallback past broken file)", result, "/fallback/plans")
	}
}
