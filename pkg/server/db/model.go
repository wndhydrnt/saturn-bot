package db

import (
	"time"
)

type RunStatus uint

const (
	RunStatusPending RunStatus = iota
	RunStatusRunning
	RunStatusFinished
	RunStatusFailed
)

type RunReason uint

const (
	RunReasonManual RunReason = iota
	RunReasonChanged
	RunReasonNew
	RunReasonNext
	RunReasonWebhook
	RunReasonCron
)

type Run struct {
	Error           *string
	ID              uint `gorm:"primarykey"`
	FinishedAt      *time.Time
	Reason          RunReason
	RepositoryNames StringList `gorm:"type:text"`
	ScheduleAfter   time.Time
	StartedAt       *time.Time
	Status          RunStatus
	TaskName        string
	RunData         StringMap `gorm:"type:text"`
}

type Task struct {
	Active bool
	ID     uint `gorm:"primarykey"`
	Name   string
	Hash   string
}

type TaskResultStatus string

const (
	TaskResultStatusClosed  = "closed"
	TaskResultStatusError   = "error"
	TaskResultStatusMerged  = "merged"
	TaskResultStatusOpen    = "open"
	TaskResultStatusUnknown = "unknown"
)

type TaskResult struct {
	CreatedAt      time.Time
	Error          *string
	ID             uint `gorm:"primarykey"`
	PullRequestUrl *string
	RepositoryName string
	Result         int
	Status         TaskResultStatus
	RunID          uint
}
