package db

import "time"

type Execution struct {
	ID            uint `gorm:"primarykey"`
	FinishedAt    time.Time
	Reason        uint
	ScheduleAfter time.Time
	StartedAt     time.Time
	Status        uint
}
