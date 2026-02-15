package hook

import (
	"strings"
	"testing"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    *Input
		wantErr bool
	}{
		{
			name: "valid plan mode input",
			json: `{
				"session_id": "eb5b0174-0555-4601-804e-672d68069c89",
				"transcript_path": "/home/user/.claude/projects/test/session.jsonl",
				"cwd": "/home/user/projects/myapp",
				"hook_event_name": "Stop",
				"permission_mode": "plan",
				"stop_hook_active": false
			}`,
			want: &Input{
				SessionID:      "eb5b0174-0555-4601-804e-672d68069c89",
				TranscriptPath: "/home/user/.claude/projects/test/session.jsonl",
				CWD:            "/home/user/projects/myapp",
				HookEventName:  "Stop",
				PermissionMode: "plan",
				StopHookActive: false,
			},
		},
		{
			name: "stop_hook_active true",
			json: `{
				"session_id": "test",
				"transcript_path": "/tmp/session.jsonl",
				"cwd": "/tmp",
				"hook_event_name": "Stop",
				"permission_mode": "plan",
				"stop_hook_active": true
			}`,
			want: &Input{
				SessionID:      "test",
				TranscriptPath: "/tmp/session.jsonl",
				CWD:            "/tmp",
				HookEventName:  "Stop",
				PermissionMode: "plan",
				StopHookActive: true,
			},
		},
		{
			name: "non-plan permission mode",
			json: `{
				"session_id": "test",
				"transcript_path": "/tmp/session.jsonl",
				"cwd": "/tmp",
				"hook_event_name": "Stop",
				"permission_mode": "default",
				"stop_hook_active": false
			}`,
			want: &Input{
				SessionID:      "test",
				TranscriptPath: "/tmp/session.jsonl",
				CWD:            "/tmp",
				HookEventName:  "Stop",
				PermissionMode: "default",
				StopHookActive: false,
			},
		},
		{
			name: "extra fields are ignored",
			json: `{
				"session_id": "test",
				"transcript_path": "/tmp/session.jsonl",
				"cwd": "/tmp",
				"hook_event_name": "Stop",
				"permission_mode": "plan",
				"stop_hook_active": false,
				"unknown_field": "value"
			}`,
			want: &Input{
				SessionID:      "test",
				TranscriptPath: "/tmp/session.jsonl",
				CWD:            "/tmp",
				HookEventName:  "Stop",
				PermissionMode: "plan",
				StopHookActive: false,
			},
		},
		{
			name:    "invalid JSON",
			json:    `{broken`,
			wantErr: true,
		},
		{
			name:    "empty input",
			json:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInput(strings.NewReader(tt.json))
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseInput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.SessionID != tt.want.SessionID {
				t.Errorf("SessionID = %q, want %q", got.SessionID, tt.want.SessionID)
			}
			if got.TranscriptPath != tt.want.TranscriptPath {
				t.Errorf("TranscriptPath = %q, want %q", got.TranscriptPath, tt.want.TranscriptPath)
			}
			if got.CWD != tt.want.CWD {
				t.Errorf("CWD = %q, want %q", got.CWD, tt.want.CWD)
			}
			if got.HookEventName != tt.want.HookEventName {
				t.Errorf("HookEventName = %q, want %q", got.HookEventName, tt.want.HookEventName)
			}
			if got.PermissionMode != tt.want.PermissionMode {
				t.Errorf("PermissionMode = %q, want %q", got.PermissionMode, tt.want.PermissionMode)
			}
			if got.StopHookActive != tt.want.StopHookActive {
				t.Errorf("StopHookActive = %v, want %v", got.StopHookActive, tt.want.StopHookActive)
			}
		})
	}
}
