package cron

import "github.com/google/uuid"

const DisabledPrefix = "##CRONMA_DISABLED## "

type CronSource int

const (
	SourceUserCrontab CronSource = iota
	SourceSystemFile
)

type ScheduleKind int

const (
	KindExpression ScheduleKind = iota
	KindNamed
)

type CronSchedule struct {
	Kind  ScheduleKind
	Value string // the expression or @name
}

func (s CronSchedule) String() string { return s.Value }

type CronEntry struct {
	ID         uuid.UUID
	Source     CronSource
	SourceFile string
	Schedule   CronSchedule
	Command    string
	User       string
	Enabled    bool
}
