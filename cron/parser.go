package cron

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var namedSchedules = map[string]bool{
	"@daily":    true,
	"@weekly":   true,
	"@monthly":  true,
	"@yearly":   true,
	"@annually": true,
	"@hourly":   true,
	"@reboot":   true,
	"@midnight": true,
}

// LoadUserCrontab runs crontab -l and parses the output.
// Returns empty slice if no crontab is installed.
func LoadUserCrontab() ([]CronEntry, error) {
	cmd := exec.Command("crontab", "-l")
	out, err := cmd.Output()
	if err != nil {
		// exit status 1 means no crontab for this user — treat as empty
		return []CronEntry{}, nil
	}
	return ParseUserContent(string(out)), nil
}

// LoadSystemCrontabs reads /etc/crontab and all files in /etc/cron.d/.
func LoadSystemCrontabs() ([]CronEntry, error) {
	var entries []CronEntry

	// Read /etc/crontab
	if data, err := os.ReadFile("/etc/crontab"); err == nil {
		parsed := parseSystemContent(string(data), "/etc/crontab")
		entries = append(entries, parsed...)
	}

	// Read /etc/cron.d/*
	matches, err := filepath.Glob("/etc/cron.d/*")
	if err == nil {
		for _, path := range matches {
			info, err := os.Stat(path)
			if err != nil || info.IsDir() {
				continue
			}
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			parsed := parseSystemContent(string(data), path)
			entries = append(entries, parsed...)
		}
	}

	return entries, nil
}

// ParseUserContent parses user-format crontab content (5-field format).
// Exported for testing.
func ParseUserContent(content string) []CronEntry {
	var entries []CronEntry
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		entry, ok := parseUserLine(line)
		if ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

// parseUserLine parses a single user crontab line.
func parseUserLine(line string) (CronEntry, bool) {
	original := line
	enabled := true

	// Check for disabled prefix
	if strings.HasPrefix(line, DisabledPrefix) {
		line = strings.TrimPrefix(line, DisabledPrefix)
		enabled = false
	}

	trimmed := strings.TrimSpace(line)

	// Skip blank lines and comments (but not disabled lines already processed)
	if trimmed == "" {
		return CronEntry{}, false
	}
	if enabled && strings.HasPrefix(trimmed, "#") {
		return CronEntry{}, false
	}
	_ = original

	// Skip variable assignments (KEY=VALUE with no spaces in key)
	if isVariableAssignment(trimmed) {
		return CronEntry{}, false
	}

	// Parse named schedule (@daily, @reboot, etc.)
	if strings.HasPrefix(trimmed, "@") {
		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			return CronEntry{}, false
		}
		sched := fields[0]
		if !namedSchedules[sched] {
			return CronEntry{}, false
		}
		command := strings.Join(fields[1:], " ")
		return CronEntry{
			ID:       uuid.New(),
			Source:   SourceUserCrontab,
			Schedule: CronSchedule{Kind: KindNamed, Value: sched},
			Command:  command,
			Enabled:  enabled,
		}, true
	}

	// Parse standard 5-field format: min hour dom month dow command
	fields := strings.Fields(trimmed)
	if len(fields) < 6 {
		return CronEntry{}, false
	}
	expr := strings.Join(fields[0:5], " ")
	command := strings.Join(fields[5:], " ")

	return CronEntry{
		ID:       uuid.New(),
		Source:   SourceUserCrontab,
		Schedule: CronSchedule{Kind: KindExpression, Value: expr},
		Command:  command,
		Enabled:  enabled,
	}, true
}

// parseSystemContent parses system crontab content (6-field format with username).
func parseSystemContent(content string, sourceFile string) []CronEntry {
	var entries []CronEntry
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines, comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip variable assignments
		if isVariableAssignment(line) {
			continue
		}

		entry, ok := parseSystemLine(line, sourceFile)
		if ok {
			entries = append(entries, entry)
		}
	}
	return entries
}

// parseSystemLine parses a single system crontab line (6-field format).
func parseSystemLine(line string, sourceFile string) (CronEntry, bool) {
	// Named schedule: @name user command...
	if strings.HasPrefix(line, "@") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return CronEntry{}, false
		}
		sched := fields[0]
		if !namedSchedules[sched] {
			return CronEntry{}, false
		}
		user := fields[1]
		command := strings.Join(fields[2:], " ")
		return CronEntry{
			ID:         uuid.New(),
			Source:     SourceSystemFile,
			SourceFile: sourceFile,
			Schedule:   CronSchedule{Kind: KindNamed, Value: sched},
			Command:    command,
			User:       user,
			Enabled:    true,
		}, true
	}

	// Standard 6-field: min hour dom month dow user command
	fields := strings.Fields(line)
	if len(fields) < 7 {
		return CronEntry{}, false
	}
	expr := strings.Join(fields[0:5], " ")
	user := fields[5]
	command := strings.Join(fields[6:], " ")

	return CronEntry{
		ID:         uuid.New(),
		Source:     SourceSystemFile,
		SourceFile: sourceFile,
		Schedule:   CronSchedule{Kind: KindExpression, Value: expr},
		Command:    command,
		User:       user,
		Enabled:    true,
	}, true
}

// isVariableAssignment returns true if the line looks like KEY=VALUE
// where KEY contains no spaces.
func isVariableAssignment(line string) bool {
	idx := strings.Index(line, "=")
	if idx <= 0 {
		return false
	}
	key := line[:idx]
	// Key must have no spaces
	return !strings.ContainsAny(key, " \t")
}
