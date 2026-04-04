package cron

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

// SerializeEntry converts a CronEntry to its crontab line representation.
func SerializeEntry(e CronEntry) string {
	var line string
	line = fmt.Sprintf("%s %s", e.Schedule.Value, e.Command)
	if !e.Enabled {
		line = DisabledPrefix + line
	}
	return line
}

// WriteUserCrontab writes the given entries to the user's crontab via `crontab -`.
func WriteUserCrontab(entries []CronEntry) error {
	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(SerializeEntry(e))
		sb.WriteString("\n")
	}

	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(sb.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("crontab write failed: %w\n%s", err, string(out))
	}
	return nil
}

// ToggleEntry flips the Enabled state of the entry with the given ID.
func ToggleEntry(entries []CronEntry, id uuid.UUID) []CronEntry {
	result := make([]CronEntry, len(entries))
	copy(result, entries)
	for i := range result {
		if result[i].ID == id {
			result[i].Enabled = !result[i].Enabled
			break
		}
	}
	return result
}

// DeleteEntry removes the entry with the given ID from the slice.
func DeleteEntry(entries []CronEntry, id uuid.UUID) []CronEntry {
	result := make([]CronEntry, 0, len(entries))
	for _, e := range entries {
		if e.ID != id {
			result = append(result, e)
		}
	}
	return result
}

// AddEntry creates a new CronEntry and appends it to entries. Returns the new entry.
func AddEntry(entries []CronEntry, schedule CronSchedule, command string) CronEntry {
	e := CronEntry{
		ID:       uuid.New(),
		Source:   SourceUserCrontab,
		Schedule: schedule,
		Command:  command,
		Enabled:  true,
	}
	return e
}

// UpdateEntry replaces the schedule and command of the entry with the given ID.
func UpdateEntry(entries []CronEntry, id uuid.UUID, schedule CronSchedule, command string) []CronEntry {
	result := make([]CronEntry, len(entries))
	copy(result, entries)
	for i := range result {
		if result[i].ID == id {
			result[i].Schedule = schedule
			result[i].Command = command
			break
		}
	}
	return result
}
