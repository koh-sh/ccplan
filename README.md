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
| `--theme` | Color theme: `dark` (default), `light` |

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
| `c` | Add comment |
| `C` | Manage comments (edit/delete) |
| `v` | Toggle viewed mark |
| `/` | Search steps |
| `s` | Submit review and exit |
| `q` / `Ctrl+C` | Quit |
| `?` | Show help |

### Comment Mode

| Key | Action |
|-----|--------|
| `Tab` | Cycle comment label |
| `Ctrl+S` | Save comment |
| `Esc` | Cancel |

### Comment List Mode

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate comments |
| `e` | Edit selected comment |
| `d` | Delete selected comment |
| `Esc` | Back to normal mode |

### Search Mode

| Key | Action |
|-----|--------|
| Type text | Incremental filter |
| `j` / `k` | Navigate results |
| `Enter` | Confirm search |
| `Esc` | Cancel search |

## Review Output Format

The review output generated on submit uses [Conventional Comments](https://conventionalcomments.org/) labels:

```markdown
# Plan Review

Please review and address the following comments on: /path/to/plan.md

## S1.1: JWT verification
[suggestion] Switch to HS256. Load the key from an environment variable.

## S2: Update routing
[issue] Not needed; the existing implementation covers this.

## S3: Add tests
[question] Is the coverage target 80% or 90%?
```

Labels: `suggestion` (default), `issue`, `question`, `nitpick`, `todo`, `thought`, `note`, `praise`, `chore`
