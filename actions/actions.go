package actions

import (
	"fmt"
	"os/exec"
)

// RemoveWorktree runs `git worktree remove` for the given worktree path.
func RemoveWorktree(repoPath, worktreePath string) error {
	return exec.Command("git", "-C", repoPath, "worktree", "remove", worktreePath).Run()
}

// ForceRemoveWorktree runs `git worktree remove --force`.
func ForceRemoveWorktree(repoPath, worktreePath string) error {
	return exec.Command("git", "-C", repoPath, "worktree", "remove", "--force", worktreePath).Run()
}

// DeleteBranch runs `git branch -d`.
func DeleteBranch(repoPath, name string) error {
	return exec.Command("git", "-C", repoPath, "branch", "-d", name).Run()
}

// ForceDeleteBranch runs `git branch -D`.
func ForceDeleteBranch(repoPath, name string) error {
	return exec.Command("git", "-C", repoPath, "branch", "-D", name).Run()
}

// DropStash runs `git stash drop stash@{N}`.
func DropStash(repoPath string, index int) error {
	ref := fmt.Sprintf("stash@{%d}", index)
	return exec.Command("git", "-C", repoPath, "stash", "drop", ref).Run()
}

// OpenTerminal opens a new Terminal window at the given path.
func OpenTerminal(path string) error {
	return exec.Command("open", "-a", "Terminal", path).Run()
}

// OpenVSCode opens VSCode at the given path.
func OpenVSCode(path string) error {
	return exec.Command("code", path).Run()
}
