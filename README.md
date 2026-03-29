# wtui

A terminal UI for managing git worktrees across repositories.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)

## Install

```bash
git clone https://github.com/brian-bell/wtui.git
cd wtui
make build
```

The binary is built to `bin/wtui`.

## Usage

```bash
# Run with default root (~/dev)
./bin/wtui

# Run with a custom root
WORKTREE_ROOT=~/projects ./bin/wtui
```

### Keys

The UI has two panes: repos on the left, content on the right. `tab` switches focus between them. The active pane is highlighted with a blue border.

**Destructive mode:** The app starts in read-only mode — deletion keys are disabled. Press `D` (Shift+D) to toggle destructive mode on/off. When active, the right pane border turns red and delete/drop hints appear in red as a visual warning.

**Left pane (repos)**

| Key | Action |
|-----|--------|
| `↑`/`k` | Select previous repo |
| `↓`/`j` | Select next repo |
| `D` | Toggle destructive mode |
| `tab` | Switch focus to right pane |
| `q`/`esc` | Quit |

**Right pane (content)**

| Key | Action |
|-----|--------|
| `↑`/`k` | Move selection up |
| `↓`/`j` | Move selection down |
| `1`/`2`/`3`/`4` | Switch to worktrees / branches / stashes / history |
| `←`/`h`/`→`/`l` | Cycle through modes |
| `enter` | View diff (dirty branch, stash, or commit) |
| `d` | Delete branch/worktree or drop stash — requires destructive mode |
| `p` | Prune stale worktree — requires destructive mode (branches view) |
| `t` | Open terminal at worktree path |
| `c` | Open VSCode at worktree path |
| `y` | Copy commit hash to clipboard (history view) |
| `D` | Toggle destructive mode |
| `tab` | Switch focus to left pane |
| `q`/`esc` | Close overlay or quit |

The right pane header shows the active mode. Press `1`–`4` or use arrow keys to switch between worktrees, branches, stashes, and history.

### Worktrees view (mode 1)

The default view. Shows all worktree checkouts for the selected repo. The main (root) worktree always appears first with a blue `[root]` annotation.

Each row shows the branch name (or `(detached)` for detached HEAD), status indicators, and the worktree path:

- `✔` green: clean working tree
- `●` red: dirty — shows `N files +X/-Y` (lines added/deleted)
- `✗` red: stale — worktree directory no longer exists

### Branches view (mode 2)

Shows all local branches alphabetically with stacking indicators:

- `✔` green: even with upstream, clean working tree
- `●` yellow: ahead/behind upstream — shows `+N/-N` counts
- `●` red: dirty worktree — shows `N files +X/-Y` (lines added/deleted)
- `●` purple: no upstream or upstream gone

Worktree branches are annotated with `[root]` (repo root) or `[<path>]` (additional worktrees). Multi-checkout branches expand to one row per worktree. Detached worktrees appear as `(detached)` rows with their path annotation. Branches ahead of upstream show up to 5 unpushed commit messages, with overflow count. When a branch is dirty and is a worktree, `enter` opens a full-screen diff overlay. `t`/`c` open a terminal or VSCode at the worktree path. `d` removes the worktree (or deletes the branch for non-worktree branches), with a force-retry prompt on failure. `p` prunes stale worktree references. Deletion requires destructive mode to be enabled first (`D`).

### Stashes view (mode 3)

Browse stashes for the selected repo. Long stash messages wrap to two lines (date + message start, then indented continuation). Use `↑`/`↓` to select a stash, `enter` to view its diff in a full-screen overlay, `d` to drop the selected stash (with confirmation, requires destructive mode). The stash list scrolls when entries exceed the pane height.

### History view (mode 4)

Browse recent commits (up to 50) for the selected repo. Each row shows the commit hash, author, relative date, and subject. Use `enter` to view the full commit diff, `y` to copy the commit hash to clipboard, and `t`/`c` to open terminal or VSCode at the repo root.

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `WORKTREE_ROOT` | `~/dev` | Root directory to scan for git repos (up to 2 levels deep) |

## Development

```bash
make build   # Build binary to bin/wtui
make test    # Run all tests
make run     # Build and run
make tidy    # go mod tidy
make clean   # Remove bin/
```

## Requirements

- Go 1.26+
- Git 2.15+ (worktree support)
