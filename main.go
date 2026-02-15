package main

import (
	"github.com/alecthomas/kong"
	"github.com/koh-sh/ccplan/cmd"
)

var version = "dev"

func main() {
	var cli cmd.CLI
	ctx := kong.Parse(&cli,
		kong.Name("ccplan"),
		kong.Description("Claude Code Plan CLI tool"),
		kong.UsageOnError(),
		kong.Vars{"version": version},
	)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
