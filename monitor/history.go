package monitor

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

const MaxRuns = 50

// RunRecord represents a single execution of a cron job.
type RunRecord struct {
	EntryID    uuid.UUID
	Command    string
	StartedAt  time.Time
	FinishedAt *time.Time
	ExitStatus *int
	LogLines   []string
}

// Duration returns the duration of the run, or nil if still running.
func (r RunRecord) Duration() *time.Duration {
	if r.FinishedAt == nil {
		return nil
	}
	d := r.FinishedAt.Sub(r.StartedAt)
	return &d
}

// FormatDuration returns a human-readable duration string.
func (r RunRecord) FormatDuration() string {
	if r.IsRunning() {
		return "running"
	}
	d := r.FinishedAt.Sub(r.StartedAt)
	switch {
	case d < time.Second:
		return fmt.Sprintf("%dms", d.Milliseconds())
	case d < time.Minute:
		return fmt.Sprintf("%.1fs", d.Seconds())
	default:
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", mins, secs)
	}
}

// IsSuccess returns true if the run completed with exit status 0.
func (r RunRecord) IsSuccess() bool {
	return r.ExitStatus != nil && *r.ExitStatus == 0
}

// IsRunning returns true if the run has not yet finished.
func (r RunRecord) IsRunning() bool {
	return r.FinishedAt == nil
}

// RunHistory stores run records per cron entry ID.
type RunHistory struct {
	records map[uuid.UUID][]RunRecord
}

// NewRunHistory creates an empty RunHistory.
func NewRunHistory() RunHistory {
	return RunHistory{
		records: make(map[uuid.UUID][]RunRecord),
	}
}

// AddAll merges new records into the history, capping each entry at MaxRuns.
// Records are kept sorted most-recent-first.
func (h *RunHistory) AddAll(records []RunRecord) {
	if h.records == nil {
		h.records = make(map[uuid.UUID][]RunRecord)
	}

	// Group incoming records by EntryID
	byID := make(map[uuid.UUID][]RunRecord)
	for _, r := range records {
		byID[r.EntryID] = append(byID[r.EntryID], r)
	}

	for id, newRecs := range byID {
		// Replace entirely with new records (journal fetch is comprehensive)
		// Sort most recent first
		sorted := sortByStartedDesc(newRecs)
		if len(sorted) > MaxRuns {
			sorted = sorted[:MaxRuns]
		}
		h.records[id] = sorted
	}
}

// Get returns all stored records for the given entry ID (most recent first).
func (h *RunHistory) Get(id uuid.UUID) []RunRecord {
	if h.records == nil {
		return nil
	}
	return h.records[id]
}

// Recent returns the last n records for the entry (most recent first).
func (h *RunHistory) Recent(id uuid.UUID, n int) []RunRecord {
	all := h.Get(id)
	if len(all) <= n {
		return all
	}
	return all[:n]
}

// sortByStartedDesc sorts records most-recent-first (in-place copy returned).
func sortByStartedDesc(recs []RunRecord) []RunRecord {
	result := make([]RunRecord, len(recs))
	copy(result, recs)
	// Simple insertion sort — typically small slice
	for i := 1; i < len(result); i++ {
		for j := i; j > 0 && result[j].StartedAt.After(result[j-1].StartedAt); j-- {
			result[j], result[j-1] = result[j-1], result[j]
		}
	}
	return result
}
