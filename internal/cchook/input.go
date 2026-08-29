package cchook

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/koh-sh/commd/internal/cclocate"
)

// permissionModePlan is the Claude Code permission mode that triggers plan review.
const permissionModePlan = "plan"

// Hook event and tool names that identify the two supported triggers:
// PreToolUse on ExitPlanMode and PostToolUse on Write/Edit.
const (
	eventPreToolUse  = "PreToolUse"
	eventPostToolUse = "PostToolUse"
	toolExitPlanMode = "ExitPlanMode"
)

// Input represents the JSON input from a Claude Code tool hook (PreToolUse
// or PostToolUse). It embeds cclocate.HookInput for the common fields
// (session_id, transcript_path, cwd).
type Input struct {
	cclocate.HookInput
	HookEventName  string     `json:"hook_event_name"`
	PermissionMode string     `json:"permission_mode"`
	ToolName       string     `json:"tool_name"`
	ToolInput      *ToolInput `json:"tool_input"`
}

// ToolInput holds the tool parameters commd cares about. FilePath comes from
// Write/Edit calls; PlanFilePath is injected by Claude Code into ExitPlanMode
// calls (the plan is already on disk when the tool is invoked).
type ToolInput struct {
	FilePath     string `json:"file_path"`
	PlanFilePath string `json:"planFilePath"`
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
