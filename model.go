package main

import (
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	cronpkg "github.com/gavasc/grons/cron"
	"github.com/gavasc/grons/monitor"
)

// Screen represents which screen is currently shown.
type Screen int

const (
	ScreenList   Screen = iota
	ScreenDetail Screen = iota
	ScreenEditor Screen = iota
)

// Message types.
type tickMsg time.Time
type monitorTickMsg struct{}                              // timer fired — trigger a fetch
type monitorRefreshMsg struct {
	records []monitor.RunRecord
	err     error
}
type entriesLoadedMsg struct{ entries []cronpkg.CronEntry }
type errMsg struct{ err error }
type savedMsg struct{}

// EditorState holds the state for the editor screen.
type EditorState struct {
	editingID     *uuid.UUID
	focus         int // 0=schedule, 1=command
	scheduleInput textinput.Model
	commandInput  textinput.Model
	nextRun       string
	scheduleErr   string
	scheduleDesc  string
	picker        *filepicker.Model
}

// Model is the root application state.
type Model struct {
	screen   Screen
	width    int
	height   int
	entries  []cronpkg.CronEntry
	history  monitor.RunHistory
	selected int
	detail   *viewport.Model
	editor   EditorState
	errorMsg string
	errorExp time.Time
	detailErr string
}

func newModel() Model {
	schedInput := textinput.New()
	schedInput.Placeholder = "* * * * *  or  @daily"
	schedInput.Focus()
	schedInput.CharLimit = 100

	cmdInput := textinput.New()
	cmdInput.Placeholder = "/path/to/command --args"
	cmdInput.CharLimit = 500

	vp := viewport.New(80, 20)

	return Model{
		screen:  ScreenList,
		history: monitor.NewRunHistory(),
		detail:  &vp,
		editor: EditorState{
			scheduleInput: schedInput,
			commandInput:  cmdInput,
		},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadEntriesCmd(),
		tickCmd(),
		monitorTickCmd(),
	)
}

// loadEntriesCmd loads user and system crontab entries.
func loadEntriesCmd() tea.Cmd {
	return func() tea.Msg {
		userEntries, err := cronpkg.LoadUserCrontab()
		if err != nil {
			return errMsg{err: err}
		}
		sysEntries, err := cronpkg.LoadSystemCrontabs()
		if err != nil {
			// System crontab errors are non-fatal
			sysEntries = nil
		}
		all := append(userEntries, sysEntries...)
		return entriesLoadedMsg{entries: all}
	}
}

// saveCmd writes the user crontab and then triggers a reload.
func saveCmd(entries []cronpkg.CronEntry) tea.Cmd {
	return func() tea.Msg {
		// Only write user entries
		var userEntries []cronpkg.CronEntry
		for _, e := range entries {
			if e.Source == cronpkg.SourceUserCrontab {
				userEntries = append(userEntries, e)
			}
		}
		if err := cronpkg.WriteUserCrontab(userEntries); err != nil {
			return errMsg{err: err}
		}
		return savedMsg{}
	}
}

// monitorCmd fetches run records from journalctl.
func monitorCmd(entries []cronpkg.CronEntry) tea.Cmd {
	return func() tea.Msg {
		records, err := monitor.FetchRunRecords(entries)
		return monitorRefreshMsg{records: records, err: err}
	}
}

// tickCmd fires every 500ms.
func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// monitorTickCmd fires after 30s to trigger a journal refresh.
func monitorTickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return monitorTickMsg{}
	})
}
