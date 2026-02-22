# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ccplan is a Go CLI tool for reviewing Claude Code-generated plans in an interactive TUI. It parses Markdown plan files, displays them in a 2-pane interface, and supports inline commenting with feedback integration via Claude Code hooks.

## Absolute Rule

**`make ci` must always pass.** Before finishing any code change, run `make ci` and confirm all steps succeed. This is non-negotiable — do not leave the codebase in a state where `make ci` fails.

`make ci` runs: `fmt` → `fix` → `lint` → `build` → `cov` (test with coverage). If any step fails, fix it before considering the task complete.

## Build & Test Commands

Dev tools (golangci-lint, tparse, gofumpt, octocov) are managed by mise (`.mise.toml`). Run `mise install` to set up the toolchain.

```bash
make ci                                 # Run full CI pipeline (MUST pass)
make build                              # Build binary
make test                               # Run all tests with tparse
make lint                               # Run golangci-lint with --fix
make fmt                                # Format with gofumpt
make fix                                # Run go fix (modernize)
make tidy                               # Run go mod tidy -v
go test -v ./internal/plan              # Run tests for a specific package
go test -run TestParsePreamble ./internal/plan  # Run a single test
```

Linter config: `.golangci.yml` (enabled: asciicheck, gocritic, misspell, nolintlint, predeclared, unconvert; formatters: gci, gofumpt).

## Architecture

### Entry Point & CLI

`main.go` → `cmd/cli.go`: Kong struct-based CLI with 4 subcommands (review, locate, hook, version).

### Package Layout

- **`internal/plan/`** — Core domain. Markdown parsing via goldmark AST (not regex, to avoid `#` in code blocks being misinterpreted as headings). Data models (`Plan`, `Step`, `ReviewComment`), review output formatting.
- **`internal/tui/`** — Bubble Tea TUI. 2-pane layout: `StepList` (left) + `DetailPane` (right). Mode-based state machine: `ModeNormal` → `ModeComment` → `ModeCommentList` → `ModeConfirm` → `ModeHelp` → `ModeSearch`.
- **`internal/locate/`** — Plan file discovery from Claude Code transcript JSONL files. `plansDirectory` resolution chain: `.claude/settings.local.json` → `.claude/settings.json` → `~/.claude/settings.json` → `~/.claude/plans/`.
- **`internal/hook/`** — PostToolUse hook orchestration. Parses stdin JSON from Claude Code, validates `permission_mode == "plan"`, spawns review in a pane, returns exit code 0 (continue) or 2 (feedback).
- **`internal/pane/`** — Terminal multiplexer abstraction. `PaneSpawner` interface with WezTerm and Direct (fallback) implementations. tmux is stub only (`ByName("tmux")` returns DirectSpawner). `AutoDetect()` tries WezTerm → Direct. WezTerm spawner uses pixel dimensions for split direction (right 50% or bottom 80%).

### Key Data Flow

**Review**: `cmd/review.go` → `plan.Parse(markdown)` → `tui.NewApp(plan)` → Bubble Tea loop → `plan.FormatReview(comments)` → clipboard/file/stdout

**Hook**: Claude Code (PostToolUse: Write|Edit) → stdin JSON → `hook.Run()` → `pane.SpawnAndWait(ccplan review ...)` → temp file IPC → exit 0 or 2

## Development Rules

- Use `go doc` to look up library APIs and usage (more accurate than web searches)
- Do not use Python for complex scripting. Use shell scripts instead
- Do not use `/tmp`. Keep temporary files within the project directory

## Implementation Guidelines

### Code

- Follow Go idioms and best practices
- Keep design style consistent across subcommands (Kong struct tags, error handling patterns, etc.)
- Follow the DRY principle and eliminate duplicate code
- Keep package and function responsibilities single-purpose. Do not mix multiple responsibilities
- Keep functions unexported (lowercase) unless they are referenced externally
- Use function and variable names that clearly represent their responsibilities
- Add comments for code or logic that is not immediately obvious

### Tests

- Write all tests in table-driven test format
- Do not write meaningless tests that only exist to inflate coverage

### Documentation

- Keep the README clear and concise for users
- Update the README when code specifications change

## Key Design Decisions

- **goldmark AST walk** for Markdown parsing instead of regex to correctly handle `#` inside code fences
- **Mermaid → ASCII rendering** via `mermaid-ascii` library; fenced `\`\`\`mermaid` blocks are converted to ASCII art in the detail pane, with fallback to raw source on error
- **Content-hash based viewed tracking** (`state.go`): steps are tracked by title → SHA-256 hash of title+body; if content changes, the viewed mark auto-clears
- **PostToolUse (Write|Edit) hook** instead of Stop hook, because Stop doesn't fire during plan mode
- **Temp files for IPC** between hook process and review subprocess (which runs in a separate pane)
- **Hook always exits 0 on error** — never breaks the Claude Code workflow; uses `PLAN_REVIEW_SKIP=1` env var to disable

## Testing Patterns

- Golden file tests using `testdata/` directories (e.g., `internal/plan/testdata/basic.md`, `code-block-hash.md`)
- Test helper `readTestdata(t, name)` for loading fixtures
- TUI tests validate render bounds, scroll behavior, and cursor position
