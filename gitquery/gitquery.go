package gitquery

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Stash represents a single git stash entry.
type Stash struct {
	Index   int
	Date    string
	Message string
}

// ListStashes returns stash entries for the given repo path.
func ListStashes(repoPath string) ([]Stash, error) {
	out, err := exec.Command("git", "-C", repoPath, "stash", "list", "--format=%gd%x00%ai%x00%s").Output()
	if err != nil {
		return nil, fmt.Errorf("listing stashes: %w", err)
	}

	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}

	var stashes []Stash
	for _, line := range strings.Split(text, "\n") {
		parts := strings.SplitN(line, "\x00", 3)
		if len(parts) != 3 {
			continue
		}
		// parts[0] is like "stash@{0}"
		idxStr := strings.TrimPrefix(parts[0], "stash@{")
		idxStr = strings.TrimSuffix(idxStr, "}")
		idx, _ := strconv.Atoi(idxStr)
		stashes = append(stashes, Stash{
			Index:   idx,
			Date:    parts[1],
			Message: parts[2],
		})
	}
	return stashes, nil
}

// StashDiff returns the diff for a specific stash entry.
func StashDiff(repoPath string, index int) (string, error) {
	ref := fmt.Sprintf("stash@{%d}", index)
	out, err := exec.Command("git", "-C", repoPath, "stash", "show", "-p", ref).Output()
	if err != nil {
		return "", fmt.Errorf("stash diff for %s: %w", ref, err)
	}
	return string(out), nil
}

// Worktree represents a single git worktree with its status.
type Worktree struct {
	Path        string
	Branch      string
	IsBare      bool
	Dirty       bool
	HasUpstream bool
	Ahead       int
	Behind      int
	Unpushed    []string
}

// ListWorktrees returns worktree information for the given repo path.
func ListWorktrees(repoPath string) ([]Worktree, error) {
	out, err := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}

	var worktrees []Worktree
	for _, block := range splitWorktreeBlocks(string(out)) {
		wt := parseWorktreeBlock(block)
		if wt.Path == "" {
			continue
		}
		if !wt.IsBare {
			fillStatus(&wt)
		}
		worktrees = append(worktrees, wt)
	}
	return worktrees, nil
}

func splitWorktreeBlocks(output string) []string {
	var blocks []string
	var current []string
	for _, line := range strings.Split(strings.TrimRight(output, "\n"), "\n") {
		if line == "" {
			if len(current) > 0 {
				blocks = append(blocks, strings.Join(current, "\n"))
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		blocks = append(blocks, strings.Join(current, "\n"))
	}
	return blocks
}

func parseWorktreeBlock(block string) Worktree {
	var wt Worktree
	for _, line := range strings.Split(block, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			wt.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			wt.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case line == "bare":
			wt.IsBare = true
		case line == "detached":
			wt.Branch = "(detached)"
		}
	}
	return wt
}

func fillStatus(wt *Worktree) {
	// Dirty check
	out, err := exec.Command("git", "-C", wt.Path, "status", "--porcelain").Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		wt.Dirty = true
	}

	// Ahead/behind
	out, err = exec.Command("git", "-C", wt.Path, "rev-list", "--count", "--left-right", "@{upstream}...HEAD").Output()
	if err == nil {
		wt.HasUpstream = true
		parts := strings.Fields(strings.TrimSpace(string(out)))
		if len(parts) == 2 {
			wt.Behind, _ = strconv.Atoi(parts[0])
			wt.Ahead, _ = strconv.Atoi(parts[1])
		}
	}

	// Unpushed commit messages
	out, err = exec.Command("git", "-C", wt.Path, "log", "--oneline", "@{upstream}..HEAD").Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			wt.Unpushed = append(wt.Unpushed, line)
		}
	}
}
