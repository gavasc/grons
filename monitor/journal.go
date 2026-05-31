package monitor

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	cronpkg "github.com/gavasc/grons/cron"
)

// journalEntry represents a single JSON line from journalctl output.
type journalEntry struct {
	Message           string `json:"MESSAGE"`
	RealtimeTimestamp string `json:"__REALTIME_TIMESTAMP"`
	PID               string `json:"_PID"`
	SyslogIdentifier  string `json:"SYSLOG_IDENTIFIER"`
}

// pidRun tracks an in-progress or complete run keyed by PID.
type pidRun struct {
	entryID   uuid.UUID
	command   string
	startedAt time.Time
	logLines  []string
}

// FetchRunRecords runs journalctl and parses CMD/CMDEND pairs to build RunRecords.
func FetchRunRecords(entries []cronpkg.CronEntry) ([]RunRecord, error) {
	cmd := exec.Command(
		"journalctl", "-u", "crond",
		"--output=json",
		"--since", "7 days ago",
		"--no-pager",
	)
	out, err := cmd.Output()
	if err != nil {
		// journalctl may exit non-zero if no entries — not fatal
		if len(out) == 0 {
			return nil, nil
		}
	}

	// Build command -> entryID lookup
	cmdToID := buildCommandIndex(entries)

	// Parse journal lines
	inProgress := make(map[string]*pidRun) // PID -> run
	var completed []RunRecord

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var je journalEntry
		if err := json.Unmarshal([]byte(line), &je); err != nil {
			continue
		}

		ts := parseTimestamp(je.RealtimeTimestamp)
		msg := stripUserPrefix(je.Message)
		pid := je.PID

		// CMD (command) — start of a run
		if strings.HasPrefix(msg, "CMD (") && strings.HasSuffix(msg, ")") {
			cmdStr := msg[5 : len(msg)-1]
			entryID := matchCommand(cmdStr, cmdToID)

			run := &pidRun{
				entryID:   entryID,
				command:   cmdStr,
				startedAt: ts,
			}
			inProgress[pid] = run
			continue
		}

		// CMDEND (command) — end of a run
		if strings.HasPrefix(msg, "CMDEND (") {
			// Extract command and exit status
			// Format: "CMDEND (command) (returned exit status N)"
			cmdEnd := parseCMDEND(msg)
			if cmdEnd != nil {
				// Find the matching in-progress run
				var matchedPID string
				var matchedRun *pidRun
				for p, r := range inProgress {
					if r.command == cmdEnd.command {
						matchedPID = p
						matchedRun = r
						break
					}
				}
				if matchedRun == nil {
					// Try by PID
					if r, ok := inProgress[pid]; ok {
						matchedPID = pid
						matchedRun = r
					}
				}

				if matchedRun != nil {
					finishedAt := ts
					rec := RunRecord{
						EntryID:    matchedRun.entryID,
						Command:    matchedRun.command,
						StartedAt:  matchedRun.startedAt,
						FinishedAt: &finishedAt,
						ExitStatus: &cmdEnd.exitStatus,
						LogLines:   matchedRun.logLines,
					}
					completed = append(completed, rec)
					delete(inProgress, matchedPID)
				}
			}
			continue
		}

		// Intermediate log line — attach to in-progress run by PID
		if run, ok := inProgress[pid]; ok {
			if msg != "" {
				run.logLines = append(run.logLines, msg)
			}
		}
	}

	// Add still-running records
	for _, run := range inProgress {
		rec := RunRecord{
			EntryID:   run.entryID,
			Command:   run.command,
			StartedAt: run.startedAt,
			LogLines:  run.logLines,
		}
		completed = append(completed, rec)
	}

	return completed, nil
}

type cmdEndResult struct {
	command    string
	exitStatus int
}

// parseCMDEND parses "CMDEND (command) (returned exit status N)" messages.
func parseCMDEND(msg string) *cmdEndResult {
	// Strip "CMDEND (" prefix
	if !strings.HasPrefix(msg, "CMDEND (") {
		return nil
	}
	rest := msg[len("CMDEND ("):]

	// Find last ") (returned exit status "
	const suffix = ") (returned exit status "
	idx := strings.LastIndex(rest, suffix)
	if idx < 0 {
		// Try simpler format: "CMDEND (command)"
		if strings.HasSuffix(rest, ")") {
			cmd := rest[:len(rest)-1]
			return &cmdEndResult{command: cmd, exitStatus: 0}
		}
		return nil
	}

	command := rest[:idx]
	statusPart := rest[idx+len(suffix):]
	statusPart = strings.TrimSuffix(statusPart, ")")
	status, err := strconv.Atoi(strings.TrimSpace(statusPart))
	if err != nil {
		status = 0
	}
	return &cmdEndResult{command: command, exitStatus: status}
}

// parseTimestamp converts a microsecond Unix timestamp string to time.Time.
func parseTimestamp(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	us, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Now()
	}
	return time.Unix(us/1_000_000, (us%1_000_000)*1000)
}

// stripUserPrefix removes the "(username) " prefix that crond (Fedora/RHEL)
// prepends to CMD and CMDEND messages, e.g. "(root) CMD (/bin/foo)" → "CMD (/bin/foo)".
func stripUserPrefix(msg string) string {
	if len(msg) > 0 && msg[0] == '(' {
		if end := strings.Index(msg, ") "); end != -1 {
			return msg[end+2:]
		}
	}
	return msg
}

// buildCommandIndex creates a map from command string to entry ID.
func buildCommandIndex(entries []cronpkg.CronEntry) map[string]uuid.UUID {
	m := make(map[string]uuid.UUID, len(entries))
	for _, e := range entries {
		m[e.Command] = e.ID
		// Also index by basename for run-parts style matching
		parts := strings.Fields(e.Command)
		if len(parts) > 0 {
			m[parts[0]] = e.ID
		}
	}
	return m
}

// matchCommand finds the entry ID for the given command string.
// Returns uuid.Nil if no match found.
func matchCommand(cmdStr string, index map[string]uuid.UUID) uuid.UUID {
	// Exact match
	if id, ok := index[cmdStr]; ok {
		return id
	}
	// Partial match — command may be a run-parts invocation
	for cmd, id := range index {
		if strings.Contains(cmdStr, cmd) || strings.Contains(cmd, cmdStr) {
			return id
		}
	}
	return uuid.Nil
}
