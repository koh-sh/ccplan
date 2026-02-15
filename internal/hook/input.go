package hook

import (
	"encoding/json"
	"fmt"
	"io"
)

// Input represents the JSON input from a Claude Code Stop hook.
type Input struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
	PermissionMode string `json:"permission_mode"`
	StopHookActive bool   `json:"stop_hook_active"`
}

// ParseInput reads and parses hook JSON input from a reader.
func ParseInput(r io.Reader) (*Input, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("parsing hook input JSON: %w", err)
	}

	return &input, nil
}
