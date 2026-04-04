package main

import (
	"github.com/google/uuid"

	"github.com/gavasc/gronma/monitor"
	"github.com/gavasc/gronma/ui"
)

func (m Model) View() string {
	switch m.screen {
	case ScreenList:
		return ui.RenderList(m.listParams())
	case ScreenDetail:
		return ui.RenderDetail(m.detailParams())
	case ScreenEditor:
		return ui.RenderEditor(m.editorParams())
	}
	return ""
}

func (m Model) listParams() ui.ListParams {
	return ui.ListParams{
		Entries:   m.entries,
		History:   m.history,
		Selected:  m.selected,
		Width:     m.width,
		Height:    m.height,
		ErrorMsg:  m.errorMsg,
		ShowError: m.errorMsg != "" && m.errorExp.After(timeNow()),
	}
}

func (m Model) detailParams() ui.DetailParams {
	var entry interface{ ID() uuid.UUID }
	_ = entry

	if m.selected < 0 || m.selected >= len(m.entries) {
		return ui.DetailParams{
			Width:  m.width,
			Height: m.height,
		}
	}

	e := m.entries[m.selected]
	records := m.history.Get(e.ID)

	return ui.DetailParams{
		Entry:   e,
		Records: ensureRunRecords(records),
		Width:   m.width,
		Height:  m.height,
	}
}

func (m Model) editorParams() ui.EditorParams {
	var pickerView string
	if m.editor.picker != nil {
		pickerView = m.editor.picker.View()
	}

	return ui.EditorParams{
		EditingID:    m.editor.editingID,
		Focus:        m.editor.focus,
		ScheduleView: m.editor.scheduleInput.View(),
		CommandView:  m.editor.commandInput.View(),
		NextRun:      m.editor.nextRun,
		ScheduleErr:  m.editor.scheduleErr,
		ScheduleDesc: m.editor.scheduleDesc,
		Width:        m.width,
		Height:       m.height,
		PickerView:   pickerView,
	}
}

// ensureRunRecords returns records or nil slice (never nil map).
func ensureRunRecords(records []monitor.RunRecord) []monitor.RunRecord {
	if records == nil {
		return []monitor.RunRecord{}
	}
	return records
}
