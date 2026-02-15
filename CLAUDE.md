# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ccplan is a Go CLI tool for reviewing Claude Code-generated plans in an interactive TUI. It parses Markdown plan files, displays them in a 2-pane interface, and supports inline commenting with feedback integration via Claude Code hooks.

## Build & Test Commands

```bash
go build -o ccplan .                    # Build binary
go test ./...                           # Run all tests
go test -v ./internal/plan              # Run tests for a specific package
go test -run TestParsePreamble ./internal/plan  # Run a single test
```

## Architecture

### Entry Point & CLI

`main.go` → `cmd/cli.go`: Kong struct-based CLI with 4 subcommands (review, locate, hook, version).

### Package Layout

- **`internal/plan/`** — Core domain. Markdown parsing via goldmark AST (not regex, to avoid `#` in code blocks being misinterpreted as headings). Data models (`Plan`, `Step`, `ReviewComment`), review output formatting.
- **`internal/tui/`** — Bubble Tea TUI. 2-pane layout: `StepList` (left) + `DetailPane` (right). Mode-based state machine: `ModeNormal` → `ModeComment` → `ModeConfirm` → `ModeHelp` → `ModeSearch`.
- **`internal/locate/`** — Plan file discovery from Claude Code transcript JSONL files. `plansDirectory` resolution chain: `.claude/settings.local.json` → `.claude/settings.json` → `~/.claude/settings.json` → `~/.claude/plans/`.
- **`internal/hook/`** — PostToolUse hook orchestration. Parses stdin JSON from Claude Code, validates `permission_mode == "plan"`, spawns review in a pane, returns exit code 0 (continue) or 2 (feedback).
- **`internal/pane/`** — Terminal multiplexer abstraction. `PaneSpawner` interface with WezTerm, tmux (stub), and Direct (fallback) implementations. `AutoDetect()` tries WezTerm → tmux → Direct.

### Key Data Flow

**Review**: `cmd/review.go` → `plan.Parse(markdown)` → `tui.NewApp(plan)` → Bubble Tea loop → `plan.FormatReview(comments)` → clipboard/file/stdout

**Hook**: Claude Code (PostToolUse: Write|Edit) → stdin JSON → `hook.Run()` → `pane.SpawnAndWait(ccplan review ...)` → temp file IPC → exit 0 or 2

## Key Design Decisions

- **goldmark AST walk** for Markdown parsing instead of regex to correctly handle `#` inside code fences
- **PostToolUse (Write|Edit) hook** instead of Stop hook, because Stop doesn't fire during plan mode
- **Temp files for IPC** between hook process and review subprocess (which runs in a separate pane)
- **Hook always exits 0 on error** — never breaks the Claude Code workflow; uses `PLAN_REVIEW_SKIP=1` env var to disable

## Testing Patterns

- Golden file tests using `testdata/` directories (e.g., `internal/plan/testdata/basic.md`, `code-block-hash.md`)
- Test helper `readTestdata(t, name)` for loading fixtures
- TUI tests validate render bounds, scroll behavior, and cursor position
