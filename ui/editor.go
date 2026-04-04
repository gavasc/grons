package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// EditorParams contains all data needed to render the editor screen.
type EditorParams struct {
	EditingID    *uuid.UUID // nil for new entry
	Focus        int        // 0=schedule, 1=command
	ScheduleView string     // rendered textinput view
	CommandView  string     // rendered textinput view
	NextRun      string
	ScheduleErr  string
	ScheduleDesc string
	Width        int
	Height       int
	PickerView   string // non-empty when file picker is open
}

// RenderEditor renders the editor popup.
func RenderEditor(p EditorParams) string {
	// If file picker is open, show it instead
	if p.PickerView != "" {
		return renderWithFilePicker(p)
	}

	title := "New Entry"
	if p.EditingID != nil {
		title = "Edit Entry"
	}

	popupWidth := p.Width * 60 / 100
	if popupWidth < 50 {
		popupWidth = 50
	}
	if popupWidth > 100 {
		popupWidth = 100
	}

	inner := popupWidth - 6 // border + padding

	var lines []string

	lines = append(lines, TitleStyle.Render(title))
	lines = append(lines, "")

	// Schedule field
	schedLabel := "Schedule"
	if p.Focus == 0 {
		schedLabel = FocusedInputStyle.Render("Schedule")
	} else {
		schedLabel = NormalInputStyle.Render("Schedule")
	}
	if p.ScheduleDesc != "" {
		desc := DimStyle.Render(p.ScheduleDesc)
		pad := inner - lipgloss.Width(schedLabel) - lipgloss.Width(desc)
		if pad > 0 {
			schedLabel = schedLabel + strings.Repeat(" ", pad) + desc
		}
	}
	lines = append(lines, schedLabel)
	lines = append(lines, p.ScheduleView)

	// Live schedule preview
	if p.ScheduleErr != "" {
		lines = append(lines, ErrorStyle.Render("  "+p.ScheduleErr))
	} else if p.NextRun != "" {
		lines = append(lines, DimStyle.Render("  next: "+p.NextRun))
	} else {
		lines = append(lines, "")
	}

	lines = append(lines, "")

	// Command field
	cmdLabel := "Command"
	if p.Focus == 1 {
		cmdLabel = FocusedInputStyle.Render("Command")
	} else {
		cmdLabel = NormalInputStyle.Render("Command")
	}
	lines = append(lines, cmdLabel)
	lines = append(lines, p.CommandView)

	lines = append(lines, "")
	lines = append(lines, strings.Repeat("─", inner))
	lines = append(lines, DimStyle.Render(
		fmt.Sprintf("%-*s", inner,
			"Tab:next  Ctrl-F:picker  Ctrl-S:save  Esc:cancel")))

	content := strings.Join(lines, "\n")
	popup := PopupStyle.Width(popupWidth).Render(content)

	// Center the popup on the screen
	return centerPopup(popup, p.Width, p.Height)
}

func renderWithFilePicker(p EditorParams) string {
	// Show file picker centered
	pickerWidth := p.Width * 70 / 100
	if pickerWidth < 60 {
		pickerWidth = 60
	}

	title := TitleStyle.Render("Select File")
	content := title + "\n\n" + p.PickerView + "\n\n" +
		DimStyle.Render("Enter:select  Esc:cancel")

	popup := PopupStyle.Width(pickerWidth).Render(content)
	return centerPopup(popup, p.Width, p.Height)
}

// centerPopup centers the popup string within the given terminal dimensions.
func centerPopup(popup string, width, height int) string {
	popupLines := strings.Split(popup, "\n")
	popupH := len(popupLines)
	popupW := 0
	for _, l := range popupLines {
		w := lipgloss.Width(l)
		if w > popupW {
			popupW = w
		}
	}

	padTop := (height - popupH) / 2
	if padTop < 0 {
		padTop = 0
	}
	padLeft := (width - popupW) / 2
	if padLeft < 0 {
		padLeft = 0
	}

	leftPad := strings.Repeat(" ", padLeft)
	var out []string
	for i := 0; i < padTop; i++ {
		out = append(out, "")
	}
	for _, l := range popupLines {
		out = append(out, leftPad+l)
	}

	return strings.Join(out, "\n")
}
