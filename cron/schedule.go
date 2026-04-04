package cron

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

var dowNames = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

// DescribeSchedule returns a short human-readable interval description for a
// cron expression, e.g. "every 5 minutes", "every day at 09:00".
// Returns "" if the expression is invalid or too complex to describe simply.
func DescribeSchedule(expr string) string {
	expr = strings.TrimSpace(expr)
	switch expr {
	case "@reboot":
		return "on reboot"
	case "@yearly", "@annually":
		return "every year"
	case "@monthly":
		return "every month"
	case "@weekly":
		return "every week"
	case "@daily", "@midnight":
		return "every day"
	case "@hourly":
		return "every hour"
	case "":
		return ""
	}

	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return ""
	}
	min, hour, dom, month, dow := fields[0], fields[1], fields[2], fields[3], fields[4]

	wildAll := dom == "*" && month == "*" && dow == "*"

	// every minute
	if min == "*" && hour == "*" && wildAll {
		return "every minute"
	}

	// every N minutes: */N * * * *
	if strings.HasPrefix(min, "*/") && hour == "*" && wildAll {
		n := min[2:]
		return "every " + n + " minutes"
	}

	// every hour on a fixed minute: M * * * *
	if !strings.ContainsAny(min, "*/-,") && hour == "*" && wildAll {
		if min == "0" {
			return "every hour"
		}
		return "every hour at :" + zeroPad(min)
	}

	// every N hours: 0 */N * * *
	if strings.HasPrefix(hour, "*/") && wildAll {
		n := hour[2:]
		suffix := ""
		if min != "0" && !strings.ContainsAny(min, "*/-,") {
			suffix = " at :" + zeroPad(min)
		}
		return "every " + n + " hours" + suffix
	}

	// fixed minute+hour patterns (M H ...)
	if strings.ContainsAny(min, "*/-,") || strings.ContainsAny(hour, "*/-,") {
		return "" // too complex
	}

	timeStr := fmt.Sprintf("%02s:%s", zeroPad(hour), zeroPad(min))

	// every day at H:M
	if dom == "*" && month == "*" && dow == "*" {
		if min == "0" && hour == "0" {
			return "every day at midnight"
		}
		return "every day at " + timeStr
	}

	// specific day of week: M H * * D
	if dom == "*" && month == "*" && !strings.ContainsAny(dow, "*/-,") {
		n, err := parseInt(dow)
		if err == nil && n >= 0 && n <= 7 {
			if n == 7 {
				n = 0
			}
			return "every " + dowNames[n] + " at " + timeStr
		}
	}

	// weekdays: M H * * 1-5
	if dom == "*" && month == "*" && dow == "1-5" {
		return "every weekday at " + timeStr
	}

	// weekends: M H * * 6,0 or 0,6
	if dom == "*" && month == "*" && (dow == "6,0" || dow == "0,6" || dow == "6-7" || dow == "0,7") {
		return "every weekend at " + timeStr
	}

	// every month on day D: M H D * *
	if !strings.ContainsAny(dom, "*/-,") && month == "*" && dow == "*" {
		return "every month on day " + dom + " at " + timeStr
	}

	return ""
}

func zeroPad(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

var cronParser = cron.NewParser(
	cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

// NextRun returns the next scheduled time after now.
// Returns nil for @reboot (no predictable next run).
// Returns an error for invalid expressions.
func NextRun(s CronSchedule) (*time.Time, error) {
	if s.Value == "@reboot" {
		return nil, nil
	}

	sched, err := cronParser.Parse(s.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid schedule %q: %w", s.Value, err)
	}

	next := sched.Next(time.Now())
	return &next, nil
}

// ValidateSchedule returns an error if expr is not a valid cron expression.
func ValidateSchedule(expr string) error {
	if expr == "" {
		return fmt.Errorf("schedule cannot be empty")
	}
	if expr == "@reboot" {
		return nil
	}
	_, err := cronParser.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}
	return nil
}

// FormatNextRun returns a human-readable next run string like:
// "in 5m (14:30:00)" or "in 2d (Mon 14:30)" or "@reboot" etc.
func FormatNextRun(s CronSchedule) string {
	if s.Value == "@reboot" {
		return "@reboot (on boot)"
	}

	t, err := NextRun(s)
	if err != nil {
		return fmt.Sprintf("invalid: %v", err)
	}
	if t == nil {
		return "N/A"
	}

	now := time.Now()
	diff := t.Sub(now)

	var inStr string
	switch {
	case diff < time.Minute:
		inStr = fmt.Sprintf("in %ds", int(diff.Seconds()))
	case diff < time.Hour:
		mins := int(diff.Minutes())
		secs := int(diff.Seconds()) % 60
		if secs == 0 {
			inStr = fmt.Sprintf("in %dm", mins)
		} else {
			inStr = fmt.Sprintf("in %dm%ds", mins, secs)
		}
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		mins := int(diff.Minutes()) % 60
		if mins == 0 {
			inStr = fmt.Sprintf("in %dh", hours)
		} else {
			inStr = fmt.Sprintf("in %dh%dm", hours, mins)
		}
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		inStr = fmt.Sprintf("in %dd", days)
	default:
		weeks := int(diff.Hours() / 24 / 7)
		inStr = fmt.Sprintf("in %dw", weeks)
	}

	var timeStr string
	if diff < 24*time.Hour {
		timeStr = t.Format("15:04:05")
	} else {
		timeStr = t.Format("Mon 15:04")
	}

	return fmt.Sprintf("%s (%s)", inStr, timeStr)
}
