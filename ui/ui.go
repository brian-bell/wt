package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/brian-bell/wt/gitquery"
	"github.com/brian-bell/wt/scanner"
)

const LeftPaneWidth = 30

var (
	repoStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	selectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Reverse(true)
	placeholderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	statusStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dividerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	branchStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	dirtyStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	cleanStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	commitStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	activeModeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	inactiveModeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// RenderParams holds everything the renderer needs.
type RenderParams struct {
	Repos     []scanner.Repo
	Selected  int
	Width     int
	Height    int
	Mode      int
	Worktrees []gitquery.Worktree
}

// Render produces the full terminal view string.
func Render(p RenderParams) string {
	if p.Width == 0 {
		p.Width = 80
	}
	if p.Height == 0 {
		p.Height = 24
	}

	statusBar := RenderStatusBar(p.Width, p.Mode)
	contentHeight := p.Height - 1 // reserve 1 row for status bar

	// Build left pane
	leftLines := renderRepoList(p.Repos, p.Selected, contentHeight)

	// Build right pane
	rightWidth := p.Width - LeftPaneWidth - 1 // 1 for divider
	if rightWidth < 0 {
		rightWidth = 0
	}

	var rightLines []string
	if p.Mode == 1 && len(p.Worktrees) > 0 {
		rightLines = renderWorktreePane(p.Worktrees, rightWidth, contentHeight)
	} else {
		rightLines = renderPlaceholderPane(rightWidth, contentHeight)
	}

	// Build divider
	divider := make([]string, contentHeight)
	for i := range divider {
		divider[i] = dividerStyle.Render("│")
	}

	// Combine panes
	left := strings.Join(leftLines, "\n")
	mid := strings.Join(divider, "\n")
	right := strings.Join(rightLines, "\n")

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, mid, right)

	return content + "\n" + statusBar
}

// RenderStatusBar produces the bottom status bar.
func RenderStatusBar(width, mode int) string {
	modes := []struct {
		key  int
		name string
	}{
		{1, "worktrees"},
		{2, "stashes"},
		{3, "branches"},
	}

	var parts []string
	for _, m := range modes {
		if mode == m.key {
			parts = append(parts, activeModeStyle.Render(fmt.Sprintf("[%d] %s", m.key, m.name)))
		} else {
			parts = append(parts, inactiveModeStyle.Render(fmt.Sprintf(" %d %s", m.key, m.name)))
		}
	}

	text := "  " + strings.Join(parts, " ") + "  ✔ clean  ● dirty  ↑/↓ navigate  ←/→ mode  q/esc: quit"
	return statusStyle.Width(width).Render(text)
}

func renderRepoList(repos []scanner.Repo, selected, height int) []string {
	lines := make([]string, height)

	// Calculate viewport for scrolling
	start := 0
	if selected >= height {
		start = selected - height + 1
	}

	for i := 0; i < height; i++ {
		idx := start + i
		if idx < len(repos) {
			name := repos[idx].DisplayName
			if idx == selected {
				line := fmt.Sprintf(" > %s", name)
				lines[i] = selectedStyle.Width(LeftPaneWidth).Render(line)
			} else {
				line := fmt.Sprintf("   %s", name)
				lines[i] = repoStyle.Width(LeftPaneWidth).Render(line)
			}
		} else {
			lines[i] = strings.Repeat(" ", LeftPaneWidth)
		}
	}

	return lines
}

func renderWorktreePane(worktrees []gitquery.Worktree, width, height int) []string {
	var content []string

	for i, wt := range worktrees {
		if wt.IsBare {
			continue
		}

		// Branch line: "  main ✔" or "  feature/auth ● +2/-1"
		branch := branchStyle.Render(wt.Branch)
		var status string
		if wt.Dirty {
			status = dirtyStyle.Render(" ●")
		} else {
			status = cleanStyle.Render(" ✔")
		}

		var upstream string
		if wt.Ahead > 0 || wt.Behind > 0 {
			upstream = fmt.Sprintf(" +%d/-%d", wt.Ahead, wt.Behind)
		}

		line := "  " + branch + status
		if upstream != "" {
			line += " " + upstream
		}
		content = append(content, line)

		// Unpushed commits (max 5)
		maxShow := 5
		for j, msg := range wt.Unpushed {
			if j >= maxShow {
				remaining := len(wt.Unpushed) - maxShow
				content = append(content, "    "+commitStyle.Render(fmt.Sprintf("... and %d more", remaining)))
				break
			}
			content = append(content, "    "+commitStyle.Render(msg))
		}

		// Blank line between entries
		if i < len(worktrees)-1 {
			content = append(content, "")
		}
	}

	// Truncate lines to pane width
	for i, line := range content {
		content[i] = truncateToWidth(line, width)
	}

	// Pad to fill height
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		if i < len(content) {
			lines[i] = content[i]
		} else {
			lines[i] = ""
		}
	}
	return lines
}

// truncateToWidth trims a styled string to fit within maxWidth visible columns.
func truncateToWidth(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Strip ANSI, truncate runes, re-measure. Crude but correct for our use.
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > maxWidth {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}

func renderPlaceholderPane(width, height int) []string {
	lines := make([]string, height)

	placeholder := placeholderStyle.Render("nothing here yet")

	mid := height / 2
	for i := 0; i < height; i++ {
		if i == mid {
			pad := (width - lipgloss.Width(placeholder)) / 2
			if pad < 0 {
				pad = 0
			}
			lines[i] = strings.Repeat(" ", pad) + placeholder
		} else {
			lines[i] = ""
		}
	}

	return lines
}
