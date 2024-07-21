package db

import (
	"time"
)

type RunStatus uint

const (
	RunStatusPending RunStatus = iota
	RunStatusExecuting
	RunStatusFinished
	RunStatusFailed
)

type Run struct {
	ID            uint `gorm:"primarykey"`
	FinishedAt    *time.Time
	Reason        uint
	ScheduleAfter time.Time
	StartedAt     *time.Time
	Status        RunStatus
	TaskName      string
}

type Task struct {
	ID   uint `gorm:"primarykey"`
	Name string
	Hash string
}

type TaskResult struct {
	CreatedAt      time.Time
	Error          *string
	ID             uint `gorm:"primarykey"`
	RepositoryName string
	Result         uint
	RunID          uint
	TaskName       string
}
