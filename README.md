# wt

A terminal UI for managing git worktrees across repositories.

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)

## Install

```bash
git clone https://github.com/brian-bell/wt.git
cd wt
make build
```

The binary is built to `bin/wt`.

## Usage

```bash
# Run with default root (~/dev)
./bin/wt

# Run with a custom root
WORKTREE_ROOT=~/projects ./bin/wt
```

### Keys

The UI has two panes: repos on the left, content on the right. `tab` switches focus between them. The active pane is highlighted with a blue border.

**Left pane (repos)**

| Key | Action |
|-----|--------|
| `↑`/`k` | Select previous repo |
| `↓`/`j` | Select next repo |
| `tab` | Switch focus to right pane |
| `q`/`esc` | Quit |

**Right pane (content)**

| Key | Action |
|-----|--------|
| `↑`/`k` | Move selection up |
| `↓`/`j` | Move selection down |
| `1`/`2` | Switch to branches / stashes mode |
| `←`/`h`/`→`/`l` | Switch mode |
| `enter` | View diff (dirty branch or stash) |
| `d` | Delete branch/worktree or drop stash (with confirmation) |
| `t` | Open terminal at worktree path |
| `c` | Open VSCode at worktree path |
| `r` | Refresh current mode data |
| `tab` | Switch focus to left pane |
| `q`/`esc` | Close overlay or quit |

The right pane header shows the active mode (`[1] branches` or `[2] stashes`). Press `1` or `2` to switch.

### Branches view (mode 1)

The right pane shows all local branches alphabetically with stacking indicators:

- `✔` green: even with upstream, clean working tree
- `●` yellow: ahead/behind upstream — shows `+N/-N` counts
- `●` red: dirty worktree — shows `N files +X/-Y` (lines added/deleted)
- `●` purple: no upstream or upstream gone

Worktree branches are annotated with `[root]` (repo root) or `[<path>]` (additional worktrees). Multi-checkout branches expand to one row per worktree. Detached worktrees appear as `(detached)` rows with their path annotation. Branches ahead of upstream show up to 5 unpushed commit messages, with overflow count. When a branch is dirty and is a worktree, `enter` opens a full-screen diff overlay. `t`/`c` open a terminal or VSCode at the worktree path. `d` removes the worktree (or deletes the branch for non-worktree branches), with a force-retry prompt on failure.

### Stashes view (mode 2)

Browse stashes for the selected repo. Use `↑`/`↓` to select a stash, `enter` to view its diff in a full-screen overlay, `d` to drop the selected stash (with confirmation).

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `WORKTREE_ROOT` | `~/dev` | Root directory to scan for git repos (up to 2 levels deep) |

## Development

```bash
make build   # Build binary to bin/wt
make test    # Run all tests
make run     # Build and run
make tidy    # go mod tidy
make clean   # Remove bin/
```

## Requirements

- Go 1.26+
- Git 2.15+ (worktree support)
