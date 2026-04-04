package main

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	cronpkg "github.com/gavasc/gronma/cron"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.detail != nil {
			m.detail.Width = msg.Width
			m.detail.Height = msg.Height - 4
		}
		return m, nil

	case tickMsg:
		// Check error TTL
		if m.errorMsg != "" && time.Now().After(m.errorExp) {
			m.errorMsg = ""
		}
		return m, tickCmd()

	case monitorTickMsg:
		return m, monitorCmd(m.entries)

	case monitorRefreshMsg:
		if len(msg.records) > 0 {
			m.history.AddAll(msg.records)
		}
		return m, monitorTickCmd()

	case entriesLoadedMsg:
		m.entries = msg.entries
		if m.selected >= len(m.entries) && len(m.entries) > 0 {
			m.selected = len(m.entries) - 1
		}
		// Trigger initial monitor fetch
		return m, monitorCmd(m.entries)

	case errMsg:
		m.errorMsg = msg.err.Error()
		m.errorExp = time.Now().Add(4 * time.Second)
		return m, nil

	case savedMsg:
		return m, loadEntriesCmd()

	case tea.KeyMsg:
		switch m.screen {
		case ScreenList:
			return m.updateList(msg)
		case ScreenDetail:
			return m.updateDetail(msg)
		case ScreenEditor:
			return m.updateEditor(msg)
		}

	default:
		// Forward all non-key messages to the filepicker when open so its
		// internal ReadDirMsg (and similar) are delivered correctly.
		if m.screen == ScreenEditor && m.editor.picker != nil {
			fpModel, cmd := m.editor.picker.Update(msg)
			m.editor.picker = &fpModel
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "q":
		return m, tea.Quit

	case "j", "down":
		if m.selected < len(m.entries)-1 {
			m.selected++
		}
		return m, nil

	case "k", "up":
		if m.selected > 0 {
			m.selected--
		}
		return m, nil

	case "enter":
		if len(m.entries) > 0 {
			m.screen = ScreenDetail
			// Rebuild viewport content
			if m.detail != nil {
				m.detail.GotoTop()
			}
		}
		return m, nil

	case "a":
		m.screen = ScreenEditor
		m.editor = newEditorState(nil, "", "")
		return m, textinput.Blink

	case "e":
		if len(m.entries) > 0 {
			e := m.entries[m.selected]
			if e.Source == cronpkg.SourceUserCrontab {
				id := e.ID
				m.screen = ScreenEditor
				m.editor = newEditorState(&id, e.Schedule.Value, e.Command)
				return m, textinput.Blink
			}
		}
		return m, nil

	case "d":
		if len(m.entries) > 0 {
			e := m.entries[m.selected]
			if e.Source == cronpkg.SourceUserCrontab {
				m.entries = cronpkg.DeleteEntry(m.entries, e.ID)
				if m.selected >= len(m.entries) && len(m.entries) > 0 {
					m.selected = len(m.entries) - 1
				}
				return m, saveCmd(m.entries)
			}
		}
		return m, nil

	case " ":
		if len(m.entries) > 0 {
			e := m.entries[m.selected]
			if e.Source == cronpkg.SourceUserCrontab {
				m.entries = cronpkg.ToggleEntry(m.entries, e.ID)
				return m, saveCmd(m.entries)
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "b", "q", "esc":
		m.screen = ScreenList
		return m, nil

	case "j", "down":
		if m.detail != nil {
			m.detail.LineDown(1)
		}
		return m, nil

	case "k", "up":
		if m.detail != nil {
			m.detail.LineUp(1)
		}
		return m, nil

	case "e":
		if len(m.entries) > 0 {
			e := m.entries[m.selected]
			if e.Source == cronpkg.SourceUserCrontab {
				id := e.ID
				m.screen = ScreenEditor
				m.editor = newEditorState(&id, e.Schedule.Value, e.Command)
				return m, textinput.Blink
			}
		}
		return m, nil

	case "d":
		if len(m.entries) > 0 {
			e := m.entries[m.selected]
			if e.Source == cronpkg.SourceUserCrontab {
				m.entries = cronpkg.DeleteEntry(m.entries, e.ID)
				m.screen = ScreenList
				if m.selected >= len(m.entries) && len(m.entries) > 0 {
					m.selected = len(m.entries) - 1
				}
				return m, saveCmd(m.entries)
			}
		}
		return m, nil

	case " ":
		if len(m.entries) > 0 {
			e := m.entries[m.selected]
			if e.Source == cronpkg.SourceUserCrontab {
				m.entries = cronpkg.ToggleEntry(m.entries, e.ID)
				return m, saveCmd(m.entries)
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) updateEditor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If file picker is open, delegate to it
	if m.editor.picker != nil {
		return m.updateEditorPicker(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.screen = ScreenList
		return m, nil

	case "tab":
		m.editor.focus = (m.editor.focus + 1) % 2
		return m.syncEditorFocus()

	case "shift+tab":
		m.editor.focus = (m.editor.focus + 1) % 2
		return m.syncEditorFocus()

	case "ctrl+s":
		return m.editorSave()

	case "ctrl+f":
		fp := filepicker.New()
		if home, err := os.UserHomeDir(); err == nil {
			fp.CurrentDirectory = home
		}
		fp.Height = m.height - 8
		if fp.Height < 5 {
			fp.Height = 5
		}
		m.editor.picker = &fp
		return m, fp.Init()

	default:
		return m.updateEditorInput(msg)
	}
}

func (m Model) updateEditorPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.editor.picker = nil
		return m, nil
	}

	fpModel, cmd := m.editor.picker.Update(msg)
	m.editor.picker = &fpModel

	if didSelect, path := fpModel.DidSelectFile(msg); didSelect {
		m.editor.commandInput.SetValue(path)
		m.editor.picker = nil
	}

	return m, cmd
}

func (m Model) syncEditorFocus() (tea.Model, tea.Cmd) {
	if m.editor.focus == 0 {
		m.editor.scheduleInput.Focus()
		m.editor.commandInput.Blur()
	} else {
		m.editor.scheduleInput.Blur()
		m.editor.commandInput.Focus()
	}
	return m, textinput.Blink
}

func (m Model) updateEditorInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if m.editor.focus == 0 {
		m.editor.scheduleInput, cmd = m.editor.scheduleInput.Update(msg)
		m.revalidateSchedule()
	} else {
		m.editor.commandInput, cmd = m.editor.commandInput.Update(msg)
	}
	return m, cmd
}

func (m *Model) revalidateSchedule() {
	val := m.editor.scheduleInput.Value()
	if val == "" {
		m.editor.scheduleErr = ""
		m.editor.nextRun = ""
		m.editor.scheduleDesc = ""
		return
	}
	if err := cronpkg.ValidateSchedule(val); err != nil {
		m.editor.scheduleErr = err.Error()
		m.editor.nextRun = ""
		m.editor.scheduleDesc = ""
	} else {
		m.editor.scheduleErr = ""
		sched := parseScheduleValue(val)
		m.editor.nextRun = cronpkg.FormatNextRun(sched)
		m.editor.scheduleDesc = cronpkg.DescribeSchedule(val)
	}
}

func (m Model) editorSave() (tea.Model, tea.Cmd) {
	schedVal := m.editor.scheduleInput.Value()
	cmdVal := m.editor.commandInput.Value()

	if schedVal == "" {
		m.errorMsg = "schedule cannot be empty"
		m.errorExp = time.Now().Add(4 * time.Second)
		return m, nil
	}
	if cmdVal == "" {
		m.errorMsg = "command cannot be empty"
		m.errorExp = time.Now().Add(4 * time.Second)
		return m, nil
	}

	if err := cronpkg.ValidateSchedule(schedVal); err != nil {
		m.editor.scheduleErr = err.Error()
		return m, nil
	}

	sched := parseScheduleValue(schedVal)

	if m.editor.editingID == nil {
		// Add new entry
		newEntry := cronpkg.AddEntry(m.entries, sched, cmdVal)
		m.entries = append(m.entries, newEntry)
	} else {
		m.entries = cronpkg.UpdateEntry(m.entries, *m.editor.editingID, sched, cmdVal)
	}

	m.screen = ScreenList
	return m, saveCmd(m.entries)
}

// newEditorState creates a fresh EditorState for the editor screen.
func newEditorState(editingID *uuid.UUID, schedVal, cmdVal string) EditorState {
	schedInput := textinput.New()
	schedInput.Placeholder = "* * * * *  or  @daily"
	schedInput.SetValue(schedVal)
	schedInput.Focus()
	schedInput.CharLimit = 100

	cmdInput := textinput.New()
	cmdInput.Placeholder = "/path/to/command --args"
	cmdInput.SetValue(cmdVal)
	cmdInput.CharLimit = 500

	es := EditorState{
		editingID:     editingID,
		focus:         0,
		scheduleInput: schedInput,
		commandInput:  cmdInput,
	}

	// Pre-validate if we have a value
	if schedVal != "" {
		if err := cronpkg.ValidateSchedule(schedVal); err == nil {
			sched := parseScheduleValue(schedVal)
			es.nextRun = cronpkg.FormatNextRun(sched)
			es.scheduleDesc = cronpkg.DescribeSchedule(schedVal)
		}
	}

	return es
}

// parseScheduleValue returns a CronSchedule from a raw string value.
func parseScheduleValue(val string) cronpkg.CronSchedule {
	if len(val) > 0 && val[0] == '@' {
		return cronpkg.CronSchedule{Kind: cronpkg.KindNamed, Value: val}
	}
	return cronpkg.CronSchedule{Kind: cronpkg.KindExpression, Value: val}
}

// updateDetailViewport passes non-key messages to the viewport.
func (m *Model) updateDetailViewport(msg tea.Msg) tea.Cmd {
	if m.detail == nil {
		return nil
	}
	vpModel, cmd := m.detail.Update(msg)
	m.detail = &vpModel
	return cmd
}
