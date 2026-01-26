package model

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusSkipped    TaskStatus = "skipped"
)

type Task struct {
	Base
	SessionID uint       `json:"sessionId" gorm:"index"`
	ParentID  *uint      `json:"parentId" gorm:"index"`
	Content   string     `json:"content"`
	Status    TaskStatus `json:"status" gorm:"default:'pending'"`
	Priority  string     `json:"priority" gorm:"default:'medium'"` // high, medium, low
}
