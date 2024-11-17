package db

import (
	"database/sql/driver"
	"fmt"
	"strings"
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
)

// StringList represents a database type that stores a list of strings.
type StringList []string

// Scan implements [sql.Scanner].
func (sl *StringList) Scan(value any) error {
	if value == nil {
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("scan StringList: %v", value)
	}

	parts := strings.Split(str, ";")
	*sl = StringList(parts)
	return nil
}

// Scan implements [sql.Valuer].
func (sl StringList) Value() (driver.Value, error) {
	if len(sl) == 0 {
		return nil, nil
	}

	return strings.Join(sl, ";"), nil
}

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
