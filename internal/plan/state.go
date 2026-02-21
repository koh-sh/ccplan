package plan

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
)

// ViewedState tracks which steps have been viewed and their content hashes.
type ViewedState struct {
	Steps map[string]string `json:"steps"` // title -> content hash
}

// NewViewedState creates an empty ViewedState.
func NewViewedState() *ViewedState {
	return &ViewedState{Steps: make(map[string]string)}
}

// StatePath returns the sidecar file path for persisting viewed state.
func StatePath(planFile string) string {
	return planFile + ".reviewed.json"
}

// ContentHash computes a truncated SHA-256 hash of a step's title and body.
func ContentHash(s *Step) string {
	h := sha256.Sum256([]byte(s.Title + "\x00" + s.Body))
	return fmt.Sprintf("%x", h[:8])
}

// LoadViewedState reads a viewed state file. Returns an empty state on any error.
func LoadViewedState(path string) *ViewedState {
	data, err := os.ReadFile(path)
	if err != nil {
		return NewViewedState()
	}
	var state ViewedState
	if err := json.Unmarshal(data, &state); err != nil {
		return NewViewedState()
	}
	if state.Steps == nil {
		state.Steps = make(map[string]string)
	}
	return &state
}

// SaveViewedState writes the viewed state to a JSON file.
func SaveViewedState(path string, state *ViewedState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling viewed state: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("writing viewed state: %w", err)
	}
	return nil
}

// IsStepViewed returns true if the step's title is tracked and its content hash matches.
func (vs *ViewedState) IsStepViewed(s *Step) bool {
	hash, ok := vs.Steps[s.Title]
	if !ok {
		return false
	}
	return hash == ContentHash(s)
}

// MarkViewed records a step as viewed with its current content hash.
func (vs *ViewedState) MarkViewed(s *Step) {
	vs.Steps[s.Title] = ContentHash(s)
}

// UnmarkViewed removes a step's viewed status.
func (vs *ViewedState) UnmarkViewed(s *Step) {
	delete(vs.Steps, s.Title)
}
