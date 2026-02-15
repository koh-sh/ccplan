package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// CLI is the top-level command structure for ccplan.
type CLI struct {
	Review  ReviewCmd  `cmd:"" help:"Review a plan file in TUI"`
	Locate  LocateCmd  `cmd:"" help:"Locate plan file path from transcript"`
	Version VersionCmd `cmd:"" help:"Show version"`
}

// ReviewCmd is the review subcommand.
type ReviewCmd struct {
	PlanFile   string `arg:"" help:"Path to the Markdown plan file"`
	Output     string `enum:"clipboard,stdout,file" default:"clipboard" help:"Output method"`
	OutputPath string `help:"File path for file output" type:"path"`
	StatusPath string `help:"File path to write exit status" type:"path"`
	Theme      string `enum:"dark,light" default:"dark" help:"Color theme"`
	NoColor    bool   `help:"Disable colors"`
}

// LocateCmd is the locate subcommand.
type LocateCmd struct {
	Transcript string `help:"Path to transcript JSONL file" type:"existingfile"`
	CWD        string `help:"Working directory for resolving relative plansDirectory" default:"." type:"existingdir"`
	Stdin      bool   `help:"Read hook JSON input from stdin"`
	All        bool   `help:"Output all plan files found in transcript"`
	Quiet      bool   `help:"Exit code only (0=found, 1=not found), no output"`
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
