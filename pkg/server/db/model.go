package db

import (
	"time"

	"gorm.io/gorm"
)

type Run struct {
	gorm.Model
	ID            uint `gorm:"primarykey"`
	FinishedAt    time.Time
	Reason        uint
	ScheduleAfter time.Time
	StartedAt     time.Time
	Status        uint
}

type RunResult struct {
	Error          string
	ID             uint `gorm:"primarykey"`
	RepositoryName string
	Result         uint
	TaskName       string
}
