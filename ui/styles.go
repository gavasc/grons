package ui

import "github.com/charmbracelet/lipgloss"

var (
	// SuccessStyle is used for successful run indicators.
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // bright green

	// ErrorStyle is used for failed run indicators.
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // bright red

	// RunningStyle is used for in-progress run indicators.
	RunningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // bright yellow

	// DisabledStyle is used for disabled entries.
	DisabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// SystemBadgeStyle is used for the [sys] badge on system entries.
	SystemBadgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

	// SelectedRowStyle highlights the selected table row.
	SelectedRowStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("33"))

	// HeaderStyle is used for column headers.
	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))

	// FocusedInputStyle is used for the focused editor input label.
	FocusedInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Underline(true)

	// NormalInputStyle is used for unfocused editor input labels.
	NormalInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	// SectionTitleStyle is used for section headers.
	SectionTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("240"))

	// ErrorFlashStyle is used for the error flash in the status bar.
	ErrorFlashStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)

	// BorderStyle is used for panels and popups.
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	// PanelStyle is used for the left/right panels.
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("237"))

	// PopupStyle is used for editor popups.
	PopupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("33")).
			Padding(1, 2)

	// StatusBarStyle is used for the bottom status bar.
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// KeyHintStyle styles keyboard shortcut hints.
	KeyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Bold(true)

	// DimStyle is used for less important text.
	DimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// BoldStyle is plain bold.
	BoldStyle = lipgloss.NewStyle().Bold(true)

	// TitleStyle is used for popup titles.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33"))
)
