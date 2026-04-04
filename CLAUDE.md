# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...          # build
go run .                # run the TUI
go test ./...           # run all tests
go test ./cron/...      # run tests for a single package
go mod tidy             # sync dependencies
```

## What this is

**gronma** is a Go + Charm reimplementation of [cronma](../cronma) (a Rust/ratatui cron manager). It is a terminal UI for browsing, editing, and monitoring cron jobs. It reads user crontab (`crontab -l`) and system crontabs (`/etc/crontab`, `/etc/cron.d/*`), polls `journalctl -u crond` for run history, and lets the user add/edit/delete/toggle user crontab entries.

## Architecture

The app follows the **Bubble Tea Elm architecture**: a single `Model` struct with `Init`, `Update`, and `View` methods.

### Data flow

```
Init() → loadEntriesCmd + tickCmd + monitorTickCmd
           ↓                 ↓              ↓
    entriesLoadedMsg     tickMsg     monitorRefreshMsg (30s)
           ↓
    monitorCmd (initial fetch from journalctl)
```

`saveCmd` writes the user crontab via `crontab -` stdin, then fires `savedMsg` which triggers `loadEntriesCmd` to reload.

### Screen routing

`Model.screen` is one of `ScreenList | ScreenDetail | ScreenEditor`. `Update` delegates key handling to `updateList`, `updateDetail`, or `updateEditor`. `View` delegates rendering to `ui.RenderList`, `ui.RenderDetail`, or `ui.RenderEditor` — passing plain params structs (no methods, no business logic in the `ui` package).

### Package responsibilities

- **`cron/`** — pure data layer: types (`CronEntry`, `CronSchedule`), parsing, serialization, schedule computation via `robfig/cron/v3`
- **`monitor/`** — `RunHistory` ring buffer (50 records/entry) and `FetchRunRecords` which parses journalctl JSON CMD/CMDEND pairs by PID
- **`ui/`** — stateless rendering functions only; all styles live in `ui/styles.go` as package-level `lipgloss.Style` vars
- **`model.go`** — `Model`, `EditorState`, message types, `newModel`, `Init`, and all `tea.Cmd` factories
- **`update.go`** — `Update` and all `updateX` handlers including editor input delegation and file picker delegation
- **`view.go`** — `View` + params-builder helpers (`listParams`, `detailParams`, `editorParams`)

### Editor state

`EditorState` holds two `textinput.Model` instances (schedule + command) plus an optional `*filepicker.Model` (non-nil when the file picker is open). When the schedule input changes, `revalidateSchedule()` calls `cronpkg.ValidateSchedule` and updates `nextRun` / `scheduleErr` inline before returning from `Update`.

### Disabled entries

Disabled user crontab entries are serialized with the prefix `##CRONMA_DISABLED## ` (defined as `cron.DisabledPrefix`), matching the cronma format for interoperability.
