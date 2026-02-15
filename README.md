# ccplan

A CLI tool for reviewing Claude Code plan files in an interactive TUI.
Add inline review comments to each step and send feedback back to Claude Code.

## Install

```bash
go install github.com/koh-sh/ccplan@latest
```

## Subcommands

### `ccplan review`

Display a plan file in a 2-pane TUI and add review comments to each step.

```bash
ccplan review path/to/plan.md

# Output review to a file
ccplan review --output file --output-path ./review.md plan.md

# Output to stdout
ccplan review --output stdout plan.md
```

| Flag | Description |
|------|-------------|
| `--output` | Output method: `clipboard` (default), `stdout`, `file` |
| `--output-path` | File path for `--output file` |
| `--status-path` | Write exit status to file (`submitted` / `approved` / `cancelled`) |
| `--theme` | Color theme: `dark` (default), `light` |
| `--no-color` | Disable colors |

### `ccplan locate`

Locate plan file paths from a Claude Code transcript JSONL. This command is primarily used internally by `ccplan hook` to resolve plan file paths during hook execution.

```bash
ccplan locate --transcript ~/.claude/projects/.../session.jsonl

# List all plan files found in a transcript
ccplan locate --transcript session.jsonl --all
```

### `ccplan hook`

Run as a Claude Code PostToolUse (Write|Edit) hook. Detects writes to plan files and launches the review TUI to enable a feedback loop.

```bash
# Called automatically by Claude Code hook (no manual invocation needed)
ccplan hook
```

| Flag | Description |
|------|-------------|
| `--spawner` | Terminal multiplexer: `auto` (default), `wezterm`, `tmux` |
| `--theme` | Color theme: `dark` (default), `light` |
| `--no-color` | Disable colors |

## Claude Code Hook Setup

Add the following to `.claude/settings.json`:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "ccplan hook",
            "timeout": 600
          }
        ]
      }
    ]
  }
}
```

The hook only activates in plan mode and launches the review TUI when a file under `plansDirectory` is written.

- **submitted** (exit 2): Sends review comments to Claude via stderr, prompting plan revision
- **approved / cancelled** (exit 0): Continues normally

Set `PLAN_REVIEW_SKIP=1` to temporarily disable the hook.

## TUI Key Bindings

### Normal Mode

| Key | Action |
|-----|--------|
| `j` / `k` / `↑` / `↓` | Navigate steps |
| `gg` / `G` | Jump to first / last step |
| `l` / `h` / `→` / `←` | Expand / collapse |
| `Enter` | Toggle expand/collapse |
| `Tab` | Switch focus between panes |
| `c` | Add/edit comment |
| `d` | Toggle delete mark |
| `a` | Toggle approve mark |
| `s` | Submit review and exit |
| `q` / `Ctrl+C` | Quit |
| `?` | Show help |

### Comment Mode

| Key | Action |
|-----|--------|
| `Ctrl+S` | Save comment |
| `Esc` | Cancel |

## Review Output Format

The review output generated on submit:

```markdown
# Plan Review

## S1.1: JWT verification [modify]
Switch to HS256. Load the key from an environment variable.

## S2: Update routing [delete]
Not needed; the existing implementation covers this.

## S3: Add tests [approve]
Looks good as is.
```

Action types: `modify` (default), `delete`, `approve`, `insert-after`, `insert-before`
