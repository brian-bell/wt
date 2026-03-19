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

func TestModel_RightArrowCyclesMode(t *testing.T) {
	m := model.New(testRepos())
	// starts at mode 1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2, got %d", m.Mode())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.Mode() != 3 {
		t.Errorf("expected mode 3, got %d", m.Mode())
	}
	// wraps back to 1
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRight})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1 after wrap, got %d", m.Mode())
	}
}

func TestModel_LeftArrowCyclesMode(t *testing.T) {
	m := model.New(testRepos())
	// starts at mode 1, left wraps to 3
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.Mode() != 3 {
		t.Errorf("expected mode 3, got %d", m.Mode())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.Mode() != 2 {
		t.Errorf("expected mode 2, got %d", m.Mode())
	}
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyLeft})
	if m.Mode() != 1 {
		t.Errorf("expected mode 1 after wrap, got %d", m.Mode())
	}
}

func TestModel_ArrowModeSwitchFiresFetchForWorktrees(t *testing.T) {
	m := model.New(testRepos())
	// right from mode 1 to mode 2 — no fetch (not worktrees mode)
	_, cmd := update(m, tea.KeyMsg{Type: tea.KeyRight})
	if cmd != nil {
		t.Error("switching to mode 2 should not fire fetch")
	}
	// right from mode 3 wraps to mode 1 — should fetch
	m, _ = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	_, cmd = update(m, tea.KeyMsg{Type: tea.KeyRight})
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
