package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read testdata/%s: %v", name, err)
	}
	return data
}

func TestParseBasic(t *testing.T) {
	source := readTestdata(t, "basic.md")
	plan, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if plan.Title != "Plan: Authentication System" {
		t.Errorf("Title = %q, want %q", plan.Title, "Plan: Authentication System")
	}

	if plan.Preamble != "This plan implements a basic authentication system." {
		t.Errorf("Preamble = %q, want %q", plan.Preamble, "This plan implements a basic authentication system.")
	}

	if len(plan.Steps) != 3 {
		t.Fatalf("len(Steps) = %d, want 3", len(plan.Steps))
	}

	// Check top-level step IDs
	wantIDs := []string{"S1", "S2", "S3"}
	for i, want := range wantIDs {
		if plan.Steps[i].ID != want {
			t.Errorf("Steps[%d].ID = %q, want %q", i, plan.Steps[i].ID, want)
		}
	}

	// Check S1 children
	s1 := plan.Steps[0]
	if s1.Title != "Step 1: Auth Middleware" {
		t.Errorf("S1.Title = %q, want %q", s1.Title, "Step 1: Auth Middleware")
	}
	if len(s1.Children) != 2 {
		t.Fatalf("len(S1.Children) = %d, want 2", len(s1.Children))
	}
	if s1.Children[0].ID != "S1.1" {
		t.Errorf("S1.Children[0].ID = %q, want %q", s1.Children[0].ID, "S1.1")
	}
	if s1.Children[1].ID != "S1.2" {
		t.Errorf("S1.Children[1].ID = %q, want %q", s1.Children[1].ID, "S1.2")
	}

	// Check parent references
	if s1.Children[0].Parent != s1 {
		t.Error("S1.1.Parent should be S1")
	}

	// Check S2 children
	s2 := plan.Steps[1]
	if len(s2.Children) != 2 {
		t.Fatalf("len(S2.Children) = %d, want 2", len(s2.Children))
	}
	if s2.Children[0].ID != "S2.1" {
		t.Errorf("S2.Children[0].ID = %q, want %q", s2.Children[0].ID, "S2.1")
	}

	// Check S3 has no children
	s3 := plan.Steps[2]
	if len(s3.Children) != 0 {
		t.Errorf("len(S3.Children) = %d, want 0", len(s3.Children))
	}

	// Check AllSteps count: 3 top-level + 2 children of S1 + 2 children of S2 = 7
	allSteps := plan.AllSteps()
	if len(allSteps) != 7 {
		t.Errorf("len(AllSteps) = %d, want 7", len(allSteps))
	}
}

func TestParseNoHeadings(t *testing.T) {
	source := readTestdata(t, "no-headings.md")
	plan, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if plan.Title != "" {
		t.Errorf("Title = %q, want empty", plan.Title)
	}

	if plan.Preamble == "" {
		t.Error("Preamble should not be empty")
	}

	if len(plan.Steps) != 0 {
		t.Errorf("len(Steps) = %d, want 0", len(plan.Steps))
	}
}

func TestParseDeepNesting(t *testing.T) {
	source := readTestdata(t, "deep-nesting.md")
	plan, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if plan.Title != "Deep Nesting Plan" {
		t.Errorf("Title = %q, want %q", plan.Title, "Deep Nesting Plan")
	}

	if len(plan.Steps) != 2 {
		t.Fatalf("len(Steps) = %d, want 2", len(plan.Steps))
	}

	// First top-level step should have children
	s1 := plan.Steps[0]
	if s1.ID != "S1" {
		t.Errorf("S1.ID = %q, want %q", s1.ID, "S1")
	}
	if len(s1.Children) != 2 {
		t.Fatalf("len(S1.Children) = %d, want 2", len(s1.Children))
	}

	// S1.1 (Level 3) should have a child at level 4
	s11 := s1.Children[0]
	if s11.ID != "S1.1" {
		t.Errorf("S1.1.ID = %q, want %q", s11.ID, "S1.1")
	}
	if len(s11.Children) != 1 {
		t.Fatalf("len(S1.1.Children) = %d, want 1", len(s11.Children))
	}

	// S1.1.1 (Level 4) should have a child at level 5
	s111 := s11.Children[0]
	if s111.ID != "S1.1.1" {
		t.Errorf("S1.1.1.ID = %q, want %q", s111.ID, "S1.1.1")
	}
	if len(s111.Children) != 1 {
		t.Fatalf("len(S1.1.1.Children) = %d, want 1", len(s111.Children))
	}

	// Check total steps
	allSteps := plan.AllSteps()
	if len(allSteps) != 6 {
		t.Errorf("len(AllSteps) = %d, want 6", len(allSteps))
	}
}

func TestParseCodeBlockHash(t *testing.T) {
	source := readTestdata(t, "code-block-hash.md")
	plan, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if plan.Title != "Plan with Code Blocks" {
		t.Errorf("Title = %q, want %q", plan.Title, "Plan with Code Blocks")
	}

	// Should only have 2 steps (code block # should not be parsed as headings)
	if len(plan.Steps) != 2 {
		t.Fatalf("len(Steps) = %d, want 2", len(plan.Steps))
	}

	if plan.Steps[0].Title != "Step 1: Configuration" {
		t.Errorf("Steps[0].Title = %q, want %q", plan.Steps[0].Title, "Step 1: Configuration")
	}
	if plan.Steps[1].Title != "Step 2: Implementation" {
		t.Errorf("Steps[1].Title = %q, want %q", plan.Steps[1].Title, "Step 2: Implementation")
	}

	// Body of step 2 should contain the code block
	if plan.Steps[1].Body == "" {
		t.Error("Steps[1].Body should not be empty")
	}
}

func TestParseSingleStep(t *testing.T) {
	source := readTestdata(t, "single-step.md")
	plan, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if plan.Title != "Simple Plan" {
		t.Errorf("Title = %q, want %q", plan.Title, "Simple Plan")
	}

	if len(plan.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(plan.Steps))
	}

	if plan.Steps[0].ID != "S1" {
		t.Errorf("Steps[0].ID = %q, want %q", plan.Steps[0].ID, "S1")
	}
	if plan.Steps[0].Title != "The Only Step" {
		t.Errorf("Steps[0].Title = %q, want %q", plan.Steps[0].Title, "The Only Step")
	}
}

func TestParseEmptySource(t *testing.T) {
	plan, err := Parse([]byte(""))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if plan.Title != "" {
		t.Errorf("Title = %q, want empty", plan.Title)
	}
	if len(plan.Steps) != 0 {
		t.Errorf("len(Steps) = %d, want 0", len(plan.Steps))
	}
}

func TestParseMultipleH1(t *testing.T) {
	source := []byte(`# First Title

Intro text.

# Second Title

This should be a step.

## Sub Step

Content here.
`)
	plan, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if plan.Title != "First Title" {
		t.Errorf("Title = %q, want %q", plan.Title, "First Title")
	}

	// Second H1 is top-level, H2 is its child
	if len(plan.Steps) != 1 {
		t.Fatalf("len(Steps) = %d, want 1", len(plan.Steps))
	}

	if plan.Steps[0].Title != "Second Title" {
		t.Errorf("Steps[0].Title = %q, want %q", plan.Steps[0].Title, "Second Title")
	}

	if len(plan.Steps[0].Children) != 1 {
		t.Fatalf("len(Steps[0].Children) = %d, want 1", len(plan.Steps[0].Children))
	}
	if plan.Steps[0].Children[0].Title != "Sub Step" {
		t.Errorf("Steps[0].Children[0].Title = %q, want %q", plan.Steps[0].Children[0].Title, "Sub Step")
	}
}

func TestParseLevelSkip(t *testing.T) {
	source := []byte(`# Plan

### Level 3 without Level 2

Content.

## Level 2

More content.
`)
	plan, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// Both should be top-level since H3 has no H2 parent
	if len(plan.Steps) != 2 {
		t.Fatalf("len(Steps) = %d, want 2", len(plan.Steps))
	}

	if plan.Steps[0].ID != "S1" {
		t.Errorf("Steps[0].ID = %q, want %q", plan.Steps[0].ID, "S1")
	}
	if plan.Steps[1].ID != "S2" {
		t.Errorf("Steps[1].ID = %q, want %q", plan.Steps[1].ID, "S2")
	}
}

func TestFindStep(t *testing.T) {
	source := readTestdata(t, "basic.md")
	plan, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	step := plan.FindStep("S1.2")
	if step == nil {
		t.Fatal("FindStep(S1.2) returned nil")
	}
	if step.Title != "1.2 Middleware Registration" {
		t.Errorf("FindStep(S1.2).Title = %q, want %q", step.Title, "1.2 Middleware Registration")
	}

	missing := plan.FindStep("S99")
	if missing != nil {
		t.Error("FindStep(S99) should return nil")
	}
}
