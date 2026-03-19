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

| Key | Action |
|-----|--------|
| `↑`/`k` | Move repo selection up |
| `↓`/`j` | Move repo selection down |
| `tab` | Cycle modes: worktrees → stashes → branches |
| `←`/`h` | Previous item in right pane (wrapping) |
| `→`/`l` | Next item in right pane (wrapping) |
| `1`/`2`/`3` | Jump to worktrees / stashes / branches mode |
| `enter` | View stash diff (stashes mode) |
| `esc`/`q` | Close overlay or quit |

### Worktree view

The right pane shows each worktree's:

- Branch name with status indicator: `✔` clean (green), `●` dirty (yellow), `●` no upstream (red)
- Ahead/behind counts relative to upstream (`+2/-1`)
- Unpushed commit messages (up to 5, with overflow count)

### Stashes view

Browse stashes for the selected repo. Use `←`/`→` to select a stash, `enter` to view its diff in a full-screen overlay.

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
