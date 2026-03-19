# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # Build binary to bin/wt
make test           # Run all tests
make run            # Build and run the TUI
go test ./scanner   # Run tests for a single package
gofmt -l .          # Check formatting (CI enforces zero output)
```

## Architecture

Go TUI for managing git worktrees across repositories. Uses **Bubble Tea** (MVC pattern) with **lipgloss** for styling.

**Data flow:** `main` reads `WORKTREE_ROOT` env var → `scanner.Scan()` discovers repos → `model.New(repos)` creates Bubble Tea model → `Init()` fires async `fetchWorktrees` → `ui.Render()` draws two-pane layout.

- **`cmd/wt/main.go`** — Entry point. Wires scanner output into the Bubble Tea program (alt-screen mode).
- **`scanner/`** — Discovers git repos under `WORKTREE_ROOT` (default `~/dev`), up to 2 levels deep. Excludes `*-worktrees` dirs. Detects both `.git` dirs and `.git` files (worktree markers). Returns repos sorted case-insensitively.
- **`gitquery/`** — Queries git data. `ListWorktrees(repoPath)` shells out to `git worktree list --porcelain`, then per non-bare worktree runs `git status --porcelain`, `git rev-list --left-right @{upstream}...HEAD`, and `git log --oneline @{upstream}..HEAD`. Sets `HasUpstream` based on whether the upstream query succeeds. `ListStashes(repoPath)` runs `git stash list --format=%gd%x00%ai%x00%s`. `StashDiff(repoPath, index)` runs `git stash show -p stash@{N}`. Only list-level failures are hard errors; per-item failures silently default to zero values.
- **`model/`** — Bubble Tea Model. Holds repo list, selection index, terminal dimensions, active mode (1=worktrees, 2=stashes, 3=branches), worktree/stash data, stash cursor, overlay state, and overlay diff content. `tab` cycles modes; `left/right` navigate stash selection (wrapping). Number keys `1/2/3` jump directly to modes. Nav keys fire async fetch commands (`fetchWorktrees` for mode 1, `fetchStashes` for mode 2). Result messages (`WorktreeResultMsg`, `StashResultMsg`, `StashDiffResultMsg`) update state with stale-result protection. Overlay intercepts all keys when open (esc/q close, up/down scroll).
- **`ui/`** — Stateless rendering. Two-pane layout: left pane (30 chars, repo list with scrolling viewport) + divider + right pane (mode-aware: mode 1 shows worktree details, mode 2 shows stash list with selection highlight, mode 3 shows placeholder). Full-screen diff overlay replaces two-pane when active. Context-aware status bar shows mode-specific keybindings. Worktree status indicators: green `✔` (clean), yellow `●` (dirty), red `●` (no upstream).

## CI

CI runs on push to `main` and PRs targeting `main`. Checks: `gofmt`, `make test`, `make build`.

## Testing

Tests use real temp directories with actual `.git` dirs/files — no mocks. Scanner tests create nested repo structures; model tests simulate key messages via Bubble Tea's `Update()`. Gitquery tests create real git repos with remotes, commits, and worktrees to verify dirty/clean, ahead/behind, and unpushed detection.
