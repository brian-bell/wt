package model_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/model"
	"github.com/brian-bell/wt/scanner"
)

func testRepos() []scanner.Repo {
	return []scanner.Repo{
		{Path: "/dev/alpha", DisplayName: "alpha"},
		{Path: "/dev/bravo", DisplayName: "bravo"},
		{Path: "/dev/charlie", DisplayName: "charlie"},
	}
}

// update sends a message and returns the concrete Model.
func update(m model.Model, msg tea.Msg) (model.Model, tea.Cmd) {
	tm, cmd := m.Update(msg)
	return tm.(model.Model), cmd
}

func TestModel_InitialSelection(t *testing.T) {
	m := model.New(testRepos())
	if m.Selected() != 0 {
		t.Errorf("expected initial selected 0, got %d", m.Selected())
	}
}

func TestModel_DownArrowMovesSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.Selected() != 1 {
		t.Errorf("expected selected 1, got %d", m.Selected())
	}
}

func TestModel_UpArrowMovesSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.Selected() != 0 {
		t.Errorf("expected selected 0, got %d", m.Selected())
	}
}

func TestModel_DownClampsAtBottom(t *testing.T) {
	m := model.New(testRepos())
	for i := 0; i < 10; i++ {
		m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	}
	if m.Selected() != 2 {
		t.Errorf("expected selected 2 (last), got %d", m.Selected())
	}
}

func TestModel_UpClampsAtTop(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.Selected() != 0 {
		t.Errorf("expected selected 0, got %d", m.Selected())
	}
}

func TestModel_QuitReturnsQuitCmd(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestModel_CtrlCReturnsQuitCmd(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestModel_EscReturnsQuitCmd(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestModel_WindowSizeUpdates(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if m.Width() != 120 || m.Height() != 40 {
		t.Errorf("expected 120x40, got %dx%d", m.Width(), m.Height())
	}
}

func TestModel_EmptyReposNoPanic(t *testing.T) {
	m := model.New(nil)
	_ = m.View()
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
}

func TestModel_ModeSwitchOnKeyPress(t *testing.T) {
	m := model.New(testRepos())

	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2, got %d", m.Mode())
	}

	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if m.Mode() != 3 {
		t.Errorf("expected mode 3, got %d", m.Mode())
	}

	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1, got %d", m.Mode())
	}
}

func TestModel_TabCyclesMode(t *testing.T) {
	m := model.New(testRepos())
	// starts at mode 1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2, got %d", m.Mode())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.Mode() != 3 {
		t.Errorf("expected mode 3, got %d", m.Mode())
	}
	// wraps back to 1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1 after wrap, got %d", m.Mode())
	}
}

func TestModel_TabModeSwitchFiresFetch(t *testing.T) {
	m := model.New(testRepos())
	// tab from mode 1 to mode 2 — should fire fetch (stashes)
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Error("switching to mode 2 should fire stash fetch")
	}
	// tab from mode 2 to mode 3 — no fetch (branches placeholder)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	_, cmd = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Error("switching to mode 3 should not fire fetch")
	}
	// tab from mode 3 wraps to mode 1 — should fetch
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	_, cmd = update(m, tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Error("switching to mode 1 should fire fetch")
	}
}

func TestModel_ModeSwitchPreservesSelection(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown}) // select bravo
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.Selected() != 1 {
		t.Errorf("expected selection preserved at 1, got %d", m.Selected())
	}
}

func TestModel_WorktreeResultUpdatesState(t *testing.T) {
	m := model.New(testRepos())
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: "main", Dirty: false},
	}
	m, _ = update(m, model.WorktreeResultMsg{RepoPath: "/dev/alpha", Worktrees: wts})
	if len(m.Worktrees()) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(m.Worktrees()))
	}
	if m.Worktrees()[0].Branch != "main" {
		t.Errorf("expected branch 'main', got %q", m.Worktrees()[0].Branch)
	}
}

func TestModel_StaleWorktreeResultDiscarded(t *testing.T) {
	m := model.New(testRepos())
	// Move selection to bravo (index 1)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})

	// Send result for alpha (index 0) — stale
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: "main"},
	}
	m, _ = update(m, model.WorktreeResultMsg{RepoPath: "/dev/alpha", Worktrees: wts})
	if len(m.Worktrees()) != 0 {
		t.Errorf("expected stale result discarded, got %d worktrees", len(m.Worktrees()))
	}
}

func TestModel_NavKeyFiresFetchCmd(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyDown})
	if cmd == nil {
		t.Fatal("expected fetchWorktrees cmd on down arrow, got nil")
	}
}

func TestModel_InitFiresFetchCmd(t *testing.T) {
	m := model.New(testRepos())
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected fetchWorktrees cmd from Init, got nil")
	}
}

func TestModel_ModeSwitchToWorktreesFiresFetchCmd(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if cmd == nil {
		t.Fatal("expected fetchWorktrees cmd on switch to mode 1, got nil")
	}
}

func TestModel_Pressing1WhileInMode1NoFetch(t *testing.T) {
	m := model.New(testRepos())
	// Already in mode 1; pressing 1 should not fire a redundant fetch
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if cmd != nil {
		t.Error("pressing 1 while already in mode 1 should not fire fetch")
	}
}

func TestModel_DefaultModeIsWorktrees(t *testing.T) {
	m := model.New(testRepos())
	if m.Mode() != 1 {
		t.Errorf("expected default mode 1, got %d", m.Mode())
	}
}

func TestModel_ViewContainsExpectedContent(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()

	for _, name := range []string{"alpha", "bravo", "charlie"} {
		if !strings.Contains(view, name) {
			t.Errorf("view should contain repo name %q", name)
		}
	}
	if !strings.Contains(view, "q/esc: quit") {
		t.Error("view should contain quit keybinding")
	}
}

func TestModel_ViewShowsWorktreeData(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	wts := []gitquery.Worktree{
		{Path: "/dev/alpha", Branch: "main", Dirty: false, Ahead: 0},
		{Path: "/dev/alpha-feat", Branch: "feature/auth", Dirty: true, Ahead: 2, Behind: 1,
			Unpushed: []string{"abc1234 Fix login bug", "def5678 Add profile page"}},
	}
	m, _ = update(m, model.WorktreeResultMsg{RepoPath: "/dev/alpha", Worktrees: wts})

	view := m.View()

	if !strings.Contains(view, "main") {
		t.Error("view should contain branch 'main'")
	}
	if !strings.Contains(view, "feature/auth") {
		t.Error("view should contain branch 'feature/auth'")
	}
	if !strings.Contains(view, "Fix login bug") {
		t.Error("view should contain unpushed commit message")
	}
	if !strings.Contains(view, "+2/-1") {
		t.Error("view should contain ahead/behind counts")
	}
}

func TestModel_ViewMode2ShowsPlaceholder(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	view := m.View()
	if !strings.Contains(view, "nothing here yet") {
		t.Error("mode 2 should show placeholder")
	}
}

func TestModel_ViewStatusBarShowsIndicatorLegend(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})

	view := m.View()
	for _, legend := range []string{"✔ clean", "● dirty"} {
		if !strings.Contains(view, legend) {
			t.Errorf("status bar should contain legend %q", legend)
		}
	}
}

func TestModel_ViewStatusBarHighlightsActiveMode(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})

	// Default mode 1: worktrees should be bracketed, others not
	view := m.View()
	if !strings.Contains(view, "[1] worktrees") {
		t.Error("mode 1 active: status bar should contain '[1] worktrees'")
	}
	if strings.Contains(view, "[2]") {
		t.Error("mode 1 active: status bar should not bracket mode 2")
	}
	if strings.Contains(view, "[3]") {
		t.Error("mode 1 active: status bar should not bracket mode 3")
	}

	// Switch to mode 2
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	view = m.View()
	if !strings.Contains(view, "[2] stashes") {
		t.Error("mode 2 active: status bar should contain '[2] stashes'")
	}
	if strings.Contains(view, "[1]") {
		t.Error("mode 2 active: status bar should not bracket mode 1")
	}
}

// --- Stash model tests (Slices 7-20) ---

func testStashes() []gitquery.Stash {
	return []gitquery.Stash{
		{Index: 0, Date: "2026-03-18 10:00:00 -0700", Message: "WIP: feature A"},
		{Index: 1, Date: "2026-03-17 09:00:00 -0700", Message: "backup: old approach"},
		{Index: 2, Date: "2026-03-16 08:00:00 -0700", Message: "experiment"},
	}
}

func TestModel_SwitchToMode2FiresFetchStashes(t *testing.T) {
	m := model.New(testRepos())
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if cmd == nil {
		t.Fatal("expected fetchStashes cmd on switch to mode 2, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.StashResultMsg); !ok {
		t.Errorf("expected StashResultMsg, got %T", msg)
	}
}

func TestModel_StashResultUpdatesState(t *testing.T) {
	m := model.New(testRepos())
	stashes := testStashes()
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: stashes})
	if len(m.Stashes()) != 3 {
		t.Fatalf("expected 3 stashes, got %d", len(m.Stashes()))
	}
}

func TestModel_StaleStashResultDiscarded(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown}) // select bravo
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	if len(m.Stashes()) != 0 {
		t.Errorf("expected stale stash result discarded, got %d stashes", len(m.Stashes()))
	}
}

func TestModel_NavInMode2FiresFetchStashes(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyDown})
	if cmd == nil {
		t.Fatal("expected fetch cmd on nav in mode 2, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.StashResultMsg); !ok {
		t.Errorf("expected StashResultMsg, got %T", msg)
	}
}

func TestModel_RightArrowMovesStashCursor(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.StashSelected() != 1 {
		t.Errorf("expected StashSelected 1, got %d", m.StashSelected())
	}
}

func TestModel_RightArrowWrapsStashCursor(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	// Right through all 3 stashes: 0→1→2→0
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.StashSelected() != 2 {
		t.Errorf("expected StashSelected 2, got %d", m.StashSelected())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.StashSelected() != 0 {
		t.Errorf("expected StashSelected wrapped to 0, got %d", m.StashSelected())
	}
}

func TestModel_LeftArrowMovesStashCursorBack(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight}) // 0→1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight}) // 1→2
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyLeft})  // 2→1
	if m.StashSelected() != 1 {
		t.Errorf("expected StashSelected 1, got %d", m.StashSelected())
	}
}

func TestModel_LeftArrowWrapsStashCursorToBottom(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	// At 0, left wraps to last (2)
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.StashSelected() != 2 {
		t.Errorf("expected StashSelected wrapped to 2, got %d", m.StashSelected())
	}
}

func TestModel_EnterOpensOverlay(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.Overlay() != model.OverlayStashDiff {
		t.Errorf("expected OverlayStashDiff, got %d", m.Overlay())
	}
	if cmd == nil {
		t.Fatal("expected fetchStashDiff cmd, got nil")
	}
	msg := cmd()
	if _, ok := msg.(model.StashDiffResultMsg); !ok {
		t.Errorf("expected StashDiffResultMsg, got %T", msg)
	}
}

func TestModel_StashDiffResultStoresDiff(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, model.StashDiffResultMsg{RepoPath: "/dev/alpha", Index: 0, Diff: "diff --git a/f.txt"})
	if m.OverlayDiff() != "diff --git a/f.txt" {
		t.Errorf("expected diff stored, got %q", m.OverlayDiff())
	}
}

func TestModel_EscClosesOverlay(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	// Now close overlay with esc
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyEscape})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed, got %d", m.Overlay())
	}
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Error("esc with overlay open should not quit")
		}
	}
}

func TestModel_QClosesOverlayNotQuit(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	// Close with q
	m, cmd := update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if m.Overlay() != model.OverlayNone {
		t.Errorf("expected overlay closed, got %d", m.Overlay())
	}
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Error("q with overlay open should not quit")
		}
	}
}

func TestModel_OverlayScrollUpDown(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = update(m, model.StashDiffResultMsg{RepoPath: "/dev/alpha", Index: 0, Diff: "line1\nline2\nline3"})

	m, _ = update(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.OverlayScroll() != 1 {
		t.Errorf("expected scroll 1, got %d", m.OverlayScroll())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.OverlayScroll() != 0 {
		t.Errorf("expected scroll 0, got %d", m.OverlayScroll())
	}
	// Up at 0 stays 0
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyUp})
	if m.OverlayScroll() != 0 {
		t.Errorf("expected scroll clamped at 0, got %d", m.OverlayScroll())
	}
}

func TestModel_ModeKeysIgnoredInOverlay(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	// Press "1" — should not change mode
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.Mode() != 2 {
		t.Errorf("expected mode unchanged at 2, got %d", m.Mode())
	}
	if m.Overlay() != model.OverlayStashDiff {
		t.Errorf("expected overlay still open, got %d", m.Overlay())
	}
}

func TestModel_ViewMode2ShowsStashContent(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})

	view := m.View()
	if !strings.Contains(view, "WIP: feature A") {
		t.Error("view should contain stash message 'WIP: feature A'")
	}
	if !strings.Contains(view, "backup: old approach") {
		t.Error("view should contain stash message 'backup: old approach'")
	}
}

func TestModel_ViewOverlayShowsDiff(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = update(m, model.StashResultMsg{RepoPath: "/dev/alpha", Stashes: testStashes()})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyEnter})
	m, _ = update(m, model.StashDiffResultMsg{RepoPath: "/dev/alpha", Index: 0, Diff: "diff --git a/f.txt\n--- a/f.txt\n+++ b/f.txt"})

	view := m.View()
	if !strings.Contains(view, "diff --git") {
		t.Error("overlay should show diff content")
	}
	if !strings.Contains(view, "esc") {
		t.Error("overlay should show esc hint")
	}
}

func TestModel_StatusBarMode2ShowsStashKeys(t *testing.T) {
	m := model.New(testRepos())
	m, _ = update(m, tea.WindowSizeMsg{Width: 120, Height: 24})
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	view := m.View()
	if !strings.Contains(view, "←/→") {
		t.Error("mode 2 status bar should mention '←/→' for stash navigation")
	}
	if !strings.Contains(view, "enter") {
		t.Error("mode 2 status bar should mention 'enter'")
	}
}
