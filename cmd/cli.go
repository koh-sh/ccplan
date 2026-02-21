package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// CLI is the top-level command structure for ccplan.
type CLI struct {
	Review  ReviewCmd  `cmd:"" help:"Review a plan file in TUI"`
	Locate  LocateCmd  `cmd:"" help:"Locate plan file path from transcript"`
	Hook    HookCmd    `cmd:"" help:"Run as Claude Code PostToolUse hook"`
	Version VersionCmd `cmd:"" help:"Show version"`
}

// HookCmd is the hook subcommand.
type HookCmd struct {
	Spawner string `enum:"wezterm,tmux,auto" default:"auto" help:"Force specific multiplexer (wezterm|tmux|auto)"`
	Theme   string `enum:"dark,light" default:"dark" help:"Color theme (dark|light)"`
}

// ReviewCmd is the review subcommand.
type ReviewCmd struct {
	PlanFile    string `arg:"" help:"Path to the Markdown plan file"`
	Output      string `enum:"clipboard,stdout,file" default:"clipboard" help:"Output method (clipboard|stdout|file)"`
	OutputPath  string `help:"File path for file output" type:"path"`
	Theme       string `enum:"dark,light" default:"dark" help:"Color theme (dark|light)"`
	TrackViewed bool   `help:"Persist viewed state to sidecar file for change detection across sessions"`
}

// LocateCmd is the locate subcommand.
type LocateCmd struct {
	Transcript string `help:"Path to transcript JSONL file" type:"existingfile"`
	CWD        string `help:"Working directory for resolving relative plansDirectory" default:"." type:"existingdir"`
	Stdin      bool   `help:"Read hook JSON input from stdin"`
	All        bool   `help:"Output all plan files found in transcript"`
}

// VersionCmd is the version subcommand.
type VersionCmd struct {
	Version string `kong:"hidden,env='version'"`
}

// Run executes the version subcommand.
func (v *VersionCmd) Run(vars kong.Vars) error {
	fmt.Println("ccplan version " + vars["version"])
	return nil
}
