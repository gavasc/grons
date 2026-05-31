package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	cron "github.com/gavasc/grons/cron"
	"github.com/gavasc/grons/monitor"
)

// ListParams contains all data needed to render the list screen.
type ListParams struct {
	Entries   []cron.CronEntry
	History   monitor.RunHistory
	Selected  int
	Width     int
	Height    int
	ErrorMsg  string
	ShowError bool
}

// RenderList renders the two-panel list view.
func RenderList(p ListParams) string {
	if p.Width < 10 || p.Height < 5 {
		return "terminal too small"
	}

	statusBarHeight := 1
	contentHeight := p.Height - statusBarHeight

	leftWidth := p.Width * 55 / 100
	rightWidth := p.Width - leftWidth - 1 // -1 for separator gap

	left := renderEntriesPanel(p, leftWidth, contentHeight)
	right := renderPreviewPanel(p, rightWidth, contentHeight)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	statusBar := renderListStatusBar(p)

	return lipgloss.JoinVertical(lipgloss.Left, panels, statusBar)
}

func renderEntriesPanel(p ListParams, width, height int) string {
	// Build header
	colSt := "St"
	colSched := "Schedule"
	colCmd := "Command"

	stW := 3
	schedW := 14
	cmdW := width - stW - schedW - 6 // paddings/borders
	if cmdW < 8 {
		cmdW = 8
	}

	header := fmt.Sprintf(" %-*s %-*s %-*s", stW, colSt, schedW, colSched, cmdW, colCmd)
	header = HeaderStyle.Render(header)

	var rows []string
	rows = append(rows, header)

	for i, e := range p.Entries {
		stIcon := SuccessStyle.Render("●")
		if !e.Enabled {
			stIcon = DisabledStyle.Render("○")
		}

		sched := truncate(e.Schedule.Value, schedW)
		cmd := e.Command
		if e.Source == cron.SourceSystemFile {
			cmd = cmd + " " + SystemBadgeStyle.Render("[sys]")
		}
		cmdDisplay := truncate(stripANSI(cmd), cmdW)
		if e.Source == cron.SourceSystemFile {
			cmdDisplay = truncate(e.Command, cmdW-6) + " " + SystemBadgeStyle.Render("[sys]")
		}

		if !e.Enabled {
			sched = DisabledStyle.Render(sched)
			cmdDisplay = DisabledStyle.Render(truncate(e.Command, cmdW))
			if e.Source == cron.SourceSystemFile {
				cmdDisplay = DisabledStyle.Render(truncate(e.Command, cmdW-6)) + " " + SystemBadgeStyle.Render("[sys]")
			}
		}

		row := fmt.Sprintf(" %s %-*s %s", stIcon, schedW, sched, cmdDisplay)

		if i == p.Selected {
			icon := "●"
			if !e.Enabled {
				icon = "○"
			}
			row = SelectedRowStyle.Render(fmt.Sprintf(" %s %-*s %s", icon, schedW, truncate(e.Schedule.Value, schedW), truncate(e.Command, cmdW)))
		}

		rows = append(rows, row)
	}

	// Pad rows to fill height
	innerHeight := height - 2 // border top/bottom
	for len(rows) < innerHeight {
		rows = append(rows, "")
	}
	if len(rows) > innerHeight {
		rows = rows[:innerHeight]
	}

	content := strings.Join(rows, "\n")
	return PanelStyle.Width(width).Height(height).Render(content)
}

func renderPreviewPanel(p ListParams, width, height int) string {
	if len(p.Entries) == 0 {
		return PanelStyle.Width(width).Height(height).Render(DimStyle.Render("No entries"))
	}

	if p.Selected < 0 || p.Selected >= len(p.Entries) {
		return PanelStyle.Width(width).Height(height).Render("")
	}

	e := p.Entries[p.Selected]
	inner := width - 2 // border

	var lines []string

	// Entry details
	lines = append(lines, BoldStyle.Render(truncate(e.Command, inner)))
	lines = append(lines, "")

	schedLabel := SectionTitleStyle.Render("Schedule: ")
	lines = append(lines, schedLabel+e.Schedule.Value)

	nextRun := cron.FormatNextRun(e.Schedule)
	lines = append(lines, SectionTitleStyle.Render("Next run: ")+nextRun)

	statusStr := SuccessStyle.Render("enabled")
	if !e.Enabled {
		statusStr = DisabledStyle.Render("disabled")
	}
	lines = append(lines, SectionTitleStyle.Render("Status:   ")+statusStr)

	sourceStr := "user crontab"
	if e.Source == cron.SourceSystemFile {
		sourceStr = e.SourceFile
	}
	lines = append(lines, SectionTitleStyle.Render("Source:   ")+DimStyle.Render(truncate(sourceStr, inner-10)))

	lines = append(lines, "")
	lines = append(lines, SectionTitleStyle.Render("Recent runs:"))

	recent := p.History.Recent(e.ID, 10)
	if len(recent) == 0 {
		lines = append(lines, DimStyle.Render("  no runs recorded"))
	} else {
		for _, r := range recent {
			icon := "✓"
			style := SuccessStyle
			if r.IsRunning() {
				icon = "…"
				style = RunningStyle
			} else if !r.IsSuccess() {
				icon = "✗"
				style = ErrorStyle
			}

			timeStr := r.StartedAt.Format("01-02 15:04")
			dur := r.FormatDuration()
			line := fmt.Sprintf("  %s %s  %s", style.Render(icon), timeStr, DimStyle.Render(dur))
			lines = append(lines, line)
		}
	}

	// Pad to fill
	innerHeight := height - 2
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}
	content := strings.Join(lines[:min(len(lines), innerHeight)], "\n")
	return PanelStyle.Width(width).Height(height).Render(content)
}

func renderListStatusBar(p ListParams) string {
	hints := KeyHintStyle.Render("j/k") + StatusBarStyle.Render(":move  ") +
		KeyHintStyle.Render("Enter") + StatusBarStyle.Render(":detail  ") +
		KeyHintStyle.Render("a") + StatusBarStyle.Render(":add  ") +
		KeyHintStyle.Render("e") + StatusBarStyle.Render(":edit  ") +
		KeyHintStyle.Render("d") + StatusBarStyle.Render(":delete  ") +
		KeyHintStyle.Render("Space") + StatusBarStyle.Render(":toggle  ") +
		KeyHintStyle.Render("q") + StatusBarStyle.Render(":quit")

	if p.ShowError && p.ErrorMsg != "" {
		return ErrorFlashStyle.Render("  " + p.ErrorMsg)
	}
	return "  " + hints
}

// truncate truncates s to at most n runes.
func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 3 {
		return string(runes[:n])
	}
	return string(runes[:n-1]) + "…"
}

// stripANSI removes ANSI escape codes (rough approximation).
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
