package gitquery_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brian-bell/wt/gitquery"
)

// realPath resolves symlinks (macOS /var → /private/var).
func realPath(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

// initRepo creates a git repo in dir with one commit on "main".
func initRepo(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "git", "init", "-b", "main")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
}

func TestListWorktrees_SingleMainWorktree(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if wts[0].Branch != "main" {
		t.Errorf("expected branch 'main', got %q", wts[0].Branch)
	}
	if wts[0].Path != dir {
		t.Errorf("expected path %q, got %q", dir, wts[0].Path)
	}
	if wts[0].IsBare {
		t.Error("expected IsBare=false")
	}
}

func TestListWorktrees_MultipleWorktrees(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	featurePath := filepath.Join(dir, "..", "feature-wt")
	run(t, dir, "git", "worktree", "add", "-b", "feature/auth", featurePath)

	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}

	branches := map[string]bool{}
	for _, wt := range wts {
		branches[wt.Branch] = true
	}
	if !branches["main"] {
		t.Error("expected branch 'main'")
	}
	if !branches["feature/auth"] {
		t.Error("expected branch 'feature/auth'")
	}
}

func TestListWorktrees_DirtyWorktree(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	// Create an uncommitted file → dirty
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0o644); err != nil {
		t.Fatal(err)
	}

	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if !wts[0].Dirty {
		t.Error("expected Dirty=true for worktree with uncommitted file")
	}
}

func TestListWorktrees_CleanWorktree(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wts[0].Dirty {
		t.Error("expected Dirty=false for clean worktree")
	}
}

func TestListWorktrees_AheadBehind(t *testing.T) {
	// Create a bare "remote" and clone it
	tmp := realPath(t, t.TempDir())
	bare := filepath.Join(tmp, "remote.git")
	clone := filepath.Join(tmp, "clone")

	run(t, tmp, "git", "init", "--bare", "-b", "main", bare)
	run(t, tmp, "git", "clone", bare, clone)
	run(t, clone, "git", "config", "user.email", "test@test.com")
	run(t, clone, "git", "config", "user.name", "Test")

	// Create initial commit and push
	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "first")
	run(t, clone, "git", "push")

	// Make 2 local commits (ahead by 2)
	for i := range 2 {
		if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte(string(rune('b'+i))), 0o644); err != nil {
			t.Fatal(err)
		}
		run(t, clone, "git", "add", ".")
		run(t, clone, "git", "commit", "-m", "local commit")
	}

	wts, err := gitquery.ListWorktrees(clone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if !wts[0].HasUpstream {
		t.Error("expected HasUpstream=true with remote tracking")
	}
	if wts[0].Ahead != 2 {
		t.Errorf("expected Ahead=2, got %d", wts[0].Ahead)
	}
	if wts[0].Behind != 0 {
		t.Errorf("expected Behind=0, got %d", wts[0].Behind)
	}
}

func TestListWorktrees_UnpushedCommits(t *testing.T) {
	tmp := realPath(t, t.TempDir())
	bare := filepath.Join(tmp, "remote.git")
	clone := filepath.Join(tmp, "clone")

	run(t, tmp, "git", "init", "--bare", "-b", "main", bare)
	run(t, tmp, "git", "clone", bare, clone)
	run(t, clone, "git", "config", "user.email", "test@test.com")
	run(t, clone, "git", "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "first")
	run(t, clone, "git", "push")

	// Make a local commit with a known message
	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "Fix login bug")

	wts, err := gitquery.ListWorktrees(clone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts[0].Unpushed) != 1 {
		t.Fatalf("expected 1 unpushed commit, got %d", len(wts[0].Unpushed))
	}
	if !strings.Contains(wts[0].Unpushed[0], "Fix login bug") {
		t.Errorf("expected unpushed to contain 'Fix login bug', got %q", wts[0].Unpushed[0])
	}
}

func TestListWorktrees_UnpushedReturnsFullMessage(t *testing.T) {
	tmp := realPath(t, t.TempDir())
	bare := filepath.Join(tmp, "remote.git")
	clone := filepath.Join(tmp, "clone")

	run(t, tmp, "git", "init", "--bare", "-b", "main", bare)
	run(t, tmp, "git", "clone", bare, clone)
	run(t, clone, "git", "config", "user.email", "test@test.com")
	run(t, clone, "git", "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", "first")
	run(t, clone, "git", "push")

	// Commit with a very long message (>60 chars) — data layer should not truncate
	longMsg := strings.Repeat("x", 80)
	if err := os.WriteFile(filepath.Join(clone, "f.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, clone, "git", "add", ".")
	run(t, clone, "git", "commit", "-m", longMsg)

	wts, err := gitquery.ListWorktrees(clone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wts[0].Unpushed) != 1 {
		t.Fatalf("expected 1 unpushed commit, got %d", len(wts[0].Unpushed))
	}
	if !strings.Contains(wts[0].Unpushed[0], longMsg) {
		t.Errorf("expected full message preserved, got %q", wts[0].Unpushed[0])
	}
}

func TestListWorktrees_NoUpstream(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	// No remote set up → no upstream
	wts, err := gitquery.ListWorktrees(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wts[0].HasUpstream {
		t.Error("expected HasUpstream=false with no remote")
	}
	if wts[0].Ahead != 0 {
		t.Errorf("expected Ahead=0 with no upstream, got %d", wts[0].Ahead)
	}
	if wts[0].Behind != 0 {
		t.Errorf("expected Behind=0 with no upstream, got %d", wts[0].Behind)
	}
	if len(wts[0].Unpushed) != 0 {
		t.Errorf("expected no unpushed with no upstream, got %d", len(wts[0].Unpushed))
	}
}

func TestListWorktrees_InvalidPath(t *testing.T) {
	_, err := gitquery.ListWorktrees("/no/such/path")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

// --- Stash tests ---

func TestListStashes_ReturnsStashes(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	// Modify file and stash
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "stash", "push", "-m", "my stash")

	stashes, err := gitquery.ListStashes(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stashes) != 1 {
		t.Fatalf("expected 1 stash, got %d", len(stashes))
	}
	if stashes[0].Index != 0 {
		t.Errorf("expected Index 0, got %d", stashes[0].Index)
	}
	if !strings.Contains(stashes[0].Message, "my stash") {
		t.Errorf("expected message containing 'my stash', got %q", stashes[0].Message)
	}
	if stashes[0].Date == "" {
		t.Error("expected non-empty Date")
	}
}

func TestListStashes_EmptyForNoStashes(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	stashes, err := gitquery.ListStashes(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stashes) != 0 {
		t.Errorf("expected 0 stashes, got %d", len(stashes))
	}
}

func TestListStashes_MultipleStashesInOrder(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	// First stash
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "stash", "push", "-m", "older stash")

	// Second stash
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "stash", "push", "-m", "newer stash")

	stashes, err := gitquery.ListStashes(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stashes) != 2 {
		t.Fatalf("expected 2 stashes, got %d", len(stashes))
	}
	// stash@{0} is most recent
	if stashes[0].Index != 0 {
		t.Errorf("expected first stash Index 0, got %d", stashes[0].Index)
	}
	if !strings.Contains(stashes[0].Message, "newer stash") {
		t.Errorf("expected first stash to be 'newer stash', got %q", stashes[0].Message)
	}
	if stashes[1].Index != 1 {
		t.Errorf("expected second stash Index 1, got %d", stashes[1].Index)
	}
	if !strings.Contains(stashes[1].Message, "older stash") {
		t.Errorf("expected second stash to be 'older stash', got %q", stashes[1].Message)
	}
}

func TestListStashes_InvalidPath(t *testing.T) {
	_, err := gitquery.ListStashes("/no/such/path")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestStashDiff_ReturnsDiff(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "git", "stash", "push", "-m", "diff test")

	diff, err := gitquery.StashDiff(dir, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff == "" {
		t.Fatal("expected non-empty diff")
	}
	if !strings.Contains(diff, "f.txt") {
		t.Error("diff should contain filename 'f.txt'")
	}
	if !strings.Contains(diff, "---") || !strings.Contains(diff, "+++") {
		t.Error("diff should contain diff markers")
	}
}

func TestStashDiff_InvalidIndex(t *testing.T) {
	dir := realPath(t, t.TempDir())
	initRepo(t, dir)

	_, err := gitquery.StashDiff(dir, 99)
	if err == nil {
		t.Fatal("expected error for invalid stash index, got nil")
	}
}
