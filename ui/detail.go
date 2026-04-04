package ui

import (
	"fmt"
	"strings"

	cron "github.com/gavasc/gronma/cron"
	"github.com/gavasc/gronma/monitor"
)

// DetailParams contains all data needed to render the detail screen.
type DetailParams struct {
	Entry   cron.CronEntry
	Records []monitor.RunRecord // all records, most recent first
	Width   int
	Height  int
}

// RenderDetail renders the detail view for a single cron entry.
func RenderDetail(p DetailParams) string {
	if p.Width < 10 || p.Height < 5 {
		return "terminal too small"
	}

	statusBarHeight := 1
	contentHeight := p.Height - statusBarHeight

	var lines []string

	// Header
	lines = append(lines, TitleStyle.Render("Cron Entry Detail"))
	lines = append(lines, strings.Repeat("─", p.Width-2))
	lines = append(lines, "")

	// Entry details block
	inner := p.Width - 4
	lines = append(lines, BoldStyle.Render("Command:  ")+truncate(p.Entry.Command, inner-10))

	schedLabel := SectionTitleStyle.Render("Schedule: ")
	lines = append(lines, schedLabel+p.Entry.Schedule.Value)

	nextRun := cron.FormatNextRun(p.Entry.Schedule)
	lines = append(lines, SectionTitleStyle.Render("Next run: ")+nextRun)

	statusStr := SuccessStyle.Render("enabled")
	if !p.Entry.Enabled {
		statusStr = DisabledStyle.Render("disabled")
	}
	lines = append(lines, SectionTitleStyle.Render("Status:   ")+statusStr)

	sourceStr := "user crontab"
	if p.Entry.Source == cron.SourceSystemFile {
		sourceStr = p.Entry.SourceFile
	}
	if p.Entry.User != "" {
		sourceStr += "  user=" + p.Entry.User
	}
	lines = append(lines, SectionTitleStyle.Render("Source:   ")+DimStyle.Render(truncate(sourceStr, inner-10)))

	lines = append(lines, "")
	lines = append(lines, strings.Repeat("─", p.Width-2))
	lines = append(lines, SectionTitleStyle.Render("Run History"))
	lines = append(lines, "")

	// Table header
	lines = append(lines, HeaderStyle.Render(
		fmt.Sprintf("  %-19s  %-6s  %-8s  %s", "Time", "Status", "Duration", "Log"),
	))
	lines = append(lines, DimStyle.Render("  "+strings.Repeat("─", min(p.Width-4, 80))))

	// Show up to 20 records
	limit := 20
	if len(p.Records) < limit {
		limit = len(p.Records)
	}

	if limit == 0 {
		lines = append(lines, DimStyle.Render("  no runs recorded"))
	} else {
		for i := 0; i < limit; i++ {
			r := p.Records[i]

			timeStr := r.StartedAt.Format("2006-01-02 15:04:05")

			var statusStr string
			if r.IsRunning() {
				statusStr = RunningStyle.Render("...")
			} else if r.IsSuccess() {
				statusStr = SuccessStyle.Render("OK ")
			} else {
				statusStr = ErrorStyle.Render("ERR")
			}

			dur := r.FormatDuration()

			firstLog := ""
			if len(r.LogLines) > 0 {
				firstLog = truncate(r.LogLines[0], p.Width-50)
			}

			line := fmt.Sprintf("  %-19s  %s     %-8s  %s",
				timeStr, statusStr, dur, DimStyle.Render(firstLog))
			lines = append(lines, line)
		}
	}

	// Pad to content height
	for len(lines) < contentHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines[:min(len(lines), contentHeight)], "\n")
	statusBar := renderDetailStatusBar()

	return content + "\n" + statusBar
}

func renderDetailStatusBar() string {
	return "  " +
		KeyHintStyle.Render("b/q/Esc") + StatusBarStyle.Render(":back  ") +
		KeyHintStyle.Render("j/k") + StatusBarStyle.Render(":scroll  ") +
		KeyHintStyle.Render("e") + StatusBarStyle.Render(":edit  ") +
		KeyHintStyle.Render("d") + StatusBarStyle.Render(":delete  ") +
		KeyHintStyle.Render("Space") + StatusBarStyle.Render(":toggle  ") +
		KeyHintStyle.Render("Ctrl-C") + StatusBarStyle.Render(":quit")
}
